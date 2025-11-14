//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fireflycons/hypervcsi/cmd/shared"
	"github.com/fireflycons/hypervcsi/internal/linux/driver"
	"github.com/fireflycons/hypervcsi/internal/linux/kvp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	endpointFlag   string
	urlFlag        string
	driverNameFlag string
	debugAddrFlag  string
	apiKeyFlag     string
	logLevelFlag   uint32
)

var rootCmd = &cobra.Command{
	Use:   "hyperv-csi-plugin",
	Short: "Kubernetes Hyper-V CSI Windows Service",
	Long: `
The without arguments form is intended to be invoked by the Windows service control manager`,
	Run: runDriver,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&endpointFlag, "endpoint", "e", envOrDefaultString("ENDPOINT", "unix:///var/lib/kubelet/plugins/"+driver.DefaultDriverName+"/csi.sock"), "CSI endpoint")
	rootCmd.Flags().StringVarP(&urlFlag, "url", "u", os.Getenv("URL"), "URL of khypervprovider Windows Service")
	rootCmd.Flags().StringVarP(&driverNameFlag, "driver-name", "n", driver.DefaultDriverName, "Name for the driver")
	rootCmd.Flags().StringVarP(&debugAddrFlag, "debug-addr", "d", "", "Address to serve the HTTP debug server on")
	rootCmd.Flags().StringVarP(&apiKeyFlag, "api-key", "k", os.Getenv("API_KEY"), "API key to access Hyper-V service backend")
	rootCmd.Flags().Uint32VarP(&logLevelFlag, "log-level", "v", envOrDefaultUint32("LOG_LEVEL", uint32(logrus.InfoLevel)), "Log level (higher = more verbose)")

	shared.InitDocCmd(rootCmd)
	shared.InitSysinfoCmd(rootCmd)
}

func envOrDefaultString(varname, defaultValue string) string {

	if v, present := os.LookupEnv(varname); present {
		return v
	}

	return defaultValue
}

func envOrDefaultUint32(varname string, defaultValue uint32) uint32 {

	if v, present := os.LookupEnv(varname); present {
		i, err := strconv.ParseUint(v, 10, 3)

		if err != nil {
			return defaultValue
		}

		return uint32(i)
	}

	return defaultValue
}

func runDriver(*cobra.Command, []string) {

	drv, err := driver.NewDriver(
		&driver.NewDriverParams{
			Endpoint:   endpointFlag,
			URL:        urlFlag,
			DriverName: driverNameFlag,
			DebugAddr:  debugAddrFlag,
			Metadata:   kvp.New(),
			ApiKey:     apiKeyFlag,
			LogLevel: func() logrus.Level {
				if logLevelFlag > uint32(logrus.TraceLevel) {
					return logrus.TraceLevel
				}
				return logrus.Level(logLevelFlag)
			}(),
		},
	)

	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
	}()

	if err := drv.Run(ctx); err != nil {
		log.Printf("Failed to start driver: %v", err)
	}

}
