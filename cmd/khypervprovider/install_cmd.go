//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fireflycons/hypervcsi/cmd/khypervprovider/psmodule"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/logging/wineventlog"
	"github.com/fireflycons/hypervcsi/internal/windows/win32"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	sslFlag         bool
	certFlag        string
	keyFlag         string
	pvDirectoryFlag string
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs this application as a Windows service",
	Long: `
You can serve via HTTPS by either
* Generating self signed certs with --ssl
* Providing a pre-created cert with --cert and --key

If you generate them here, all the cert files will be stored in the same
directory as this EXE.

Use --directory to specify where the service will store VHD volumes it creates.
If you omit this flag, the service will locate the locally attached drive with
the most free storage and create directory "Kubernetes Persistent Volumes" at
its root.

`,

	Run: executeInstall,
}

func init() {

	installCmd.Flags().Uint32VarP(&portFlag, "port", "p", constants.DefaultServicePort, "Port service will listen on")
	installCmd.Flags().BoolVarP(&sslFlag, "ssl", "s", false, "Generate self-signed CA and server certificates to use with service")
	installCmd.Flags().StringVarP(&certFlag, "cert", "c", "", "Certificate to use for HTTPS serving")
	installCmd.Flags().StringVarP(&keyFlag, "key", "k", "", "Key associated with the certificate")
	installCmd.Flags().StringVarP(&pvDirectoryFlag, "directory", "d", "", "Directory to store PV disks in. Omit to have the service choose.")

	installCmd.MarkFlagsRequiredTogether("cert", "key")
	installCmd.MarkFlagsMutuallyExclusive("ssl", "cert")

	rootCmd.AddCommand(installCmd)
}

func executeInstall(*cobra.Command, []string) {

	if err := doInstall(); err != nil {
		log.Fatalf("%v", err)
	}
}

func doInstall() error { //nolint:gocyclo // not worth splitting up

	serviceArgs := []string{}

	useSSL := sslFlag || certFlag != ""

	hostname, err := win32.GetHostname()
	if err != nil {
		return fmt.Errorf("cannot determine hostname of this comuter: %w", err)
	}

	if pvDirectoryFlag != "" {
		//nolint:govet // intentional redeclaration of err
		if err := os.MkdirAll(pvDirectoryFlag, 0755); err != nil {
			if !errors.Is(err, os.ErrExist) {
				return fmt.Errorf("cannot create directory %s: %w", pvDirectoryFlag, err)
			}
		}

		serviceArgs = append(
			serviceArgs, []string{
				"--directory",
				pvDirectoryFlag,
			}...)
	}

	assertElevatedPrivilege()

	//nolint:govet // intentional redeclaration of err
	if err := psmodule.InstallModule(); err != nil {
		return err
	}

	log.Println("PowerShell module installed")

	exepath := mustGetExePath()

	serviceControlManager, err := mgr.Connect()

	if err != nil {
		return fmt.Errorf("cannot connect to service manager: %w", err)
	}

	defer func() {
		_ = serviceControlManager.Disconnect()
	}()

	//nolint:govet // intentional redeclaration of err
	if err := assertService(serviceControlManager, constants.HyperVServiceName); err != nil {
		return err
	}

	// Now do the SSL stuff, if requested.
	if useSSL {
		log.Println("Create self-signed certs:")

		//nolint:govet // intentional redeclaration of err
		if err := setupCerts(filepath.Dir(exepath)); err != nil {
			return err
		}
	}

	log.Println("Installing Windows service")

	//nolint:govet // intentional redeclaration of err
	if err := assertService(serviceControlManager, constants.ServiceName); err == nil {
		return fmt.Errorf("service %s already exists", constants.ServiceName)
	}

	// Generate a random API key
	apiKey := uuid.NewString()

	serviceArgs = append(
		serviceArgs, []string{
			"--api-key",
			apiKey,
		}...)

	if portFlag != constants.DefaultServicePort {
		serviceArgs = append(
			serviceArgs,
			[]string{
				"--port",
				strconv.Itoa(int(portFlag)),
			}...,
		)
	}

	endpoint := fmt.Sprintf("http://%s:%d", hostname, portFlag)
	if useSSL {
		serviceArgs = append(
			serviceArgs,
			[]string{
				"--cert",
				certFlag,
				"--key",
				keyFlag,
			}...,
		)

		endpoint = fmt.Sprintf("https://%s:%d", hostname, portFlag)
	}

	theService, err := serviceControlManager.CreateService(
		constants.ServiceName,
		exepath,
		mgr.Config{
			DisplayName: constants.ServiceDisplayName,
			Description: constants.ServiceDescription,
			StartType:   mgr.StartAutomatic,
		},
		serviceArgs...,
	)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	defer func() {

		_ = theService.Close()
	}()

	err = wineventlog.RegisterEventSource()
	if err != nil {
		_ = theService.Delete()
		return fmt.Errorf("cannot set up event log: %w: service not installed", err)
	}

	log.Println("Service installed")

	// Attempt to start the service
	if err := startService(theService); err != nil {
		return err
	}

	log.Printf(`Service Installed with the following configuration:

API Key         : %s
Service Endpoint: %s

`,
		apiKey,
		endpoint,
	)

	return nil
}

func startService(s *mgr.Service) error {

	const maxServiceWaitTime = time.Second * 30

	if err := s.Start(); err != nil {
		return fmt.Errorf("could not start service - check event log for details: %w", err)
	}

	log.Println("Service starting")

	ctx, cancel := context.WithTimeout(context.Background(), maxServiceWaitTime)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timed out waiting for service to start")

		case <-ticker.C:
			stat, _ := s.Query()
			status, ok := stateMap[stat.State]
			if !ok {
				status = fmt.Sprintf("Unknown state %d", stat.State)
			}
			log.Printf("- service status: %s", status)

			if stat.State == svc.Stopped {
				return errors.New("service stopped unexpectedly - check eventlog for errors")
			}

			if stat.State == svc.Running {
				log.Println("Service running")
				return nil
			}
		}
	}
}

func setupCerts(certsPath string) error {

	switch {
	case sslFlag:

		if err := generateCertificates(certsPath, os.Stdin); err != nil {
			return fmt.Errorf("error generating certificates: %w", err)
		}

		certFlag = filepath.Join(certsPath, "server.crt")
		keyFlag = filepath.Join(certsPath, "server.key")

	case certFlag != "" && keyFlag != "":

		certFlag, _ = filepath.Abs(certFlag)
		keyFlag, _ = filepath.Abs(keyFlag)

	default:
		return errors.New("both --cert and --key must have values")
	}

	if !fileExists(certFlag) {
		return fmt.Errorf("certificate %s not found", certFlag)
	}

	if !fileExists(keyFlag) {
		return fmt.Errorf("key file %s not found", keyFlag)
	}

	return nil
}

func fileExists(path string) bool {

	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}
