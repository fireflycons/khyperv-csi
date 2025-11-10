//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fireflycons/hypervcsi/cmd/khypervprovider/psmodule"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Uninstalls this service",
	Long:  ``,
	Run:   executeRemove,
}

const (
	maxOperationWaitTime = 30 * time.Second
)

func init() {

	rootCmd.AddCommand(removeCmd)
}

func executeRemove(*cobra.Command, []string) {

	if err := doRemove(); err != nil {
		log.Fatalf("cannot conmpletely uninstall service: %v", err)
	}
}

func doRemove() error {

	const pauseAfterStop = 500 * time.Millisecond

	assertElevatedPrivilege()

	serviceControlManager, err := mgr.Connect()

	if err != nil {
		return fmt.Errorf("cannot connect to service manager: %w", err)
	}

	defer func() {
		_ = serviceControlManager.Disconnect()
	}()

	psmodule.InstallLog.Println("Removing Windows service...")

	if err := assertService(serviceControlManager, constants.ServiceName); err != nil {
		psmodule.InstallLog.Println("service already removed")
	} else {
		// Won't error since assertService already successfully made this call
		s, _ := serviceControlManager.OpenService(constants.ServiceName)
		stat, _ := s.Control(svc.Stop)

		if stat.State != svc.Stopped {
			psmodule.InstallLog.Println("Waiting for service to stop...")
			time.Sleep(pauseAfterStop)
			ctx, cancel := context.WithTimeout(context.Background(), maxOperationWaitTime)
			defer cancel()

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

		STOP_LOOP:
			for {
				select {
				case <-ctx.Done():
					return errors.New("timed out waiting for service to stop")

				case <-ticker.C:
					stat, _ = s.Query()
					status, ok := stateMap[stat.State]
					if !ok {
						status = fmt.Sprintf("Unknown state %d", stat.State)
					}
					psmodule.InstallLog.Printf("- service status: %s", status)

					if stat.State == svc.Stopped {
						break STOP_LOOP
					}
				}
			}
		}

		if err = s.Delete(); err != nil {
			return fmt.Errorf("cannot remove service: %w", err)
		}
	}

	psmodule.InstallLog.Println("Service removed")

	// remove eventlog
	if err := eventlog.Remove(constants.ServiceName); err != nil {
		psmodule.InstallLog.Printf("Cannot remove eventlog: %v", err)
	}

	// remove module
	if err := psmodule.RemoveModule(); err != nil {
		return err
	}

	psmodule.InstallLog.Println("PowerShell module removed")

	return nil
}
