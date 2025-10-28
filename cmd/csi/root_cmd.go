//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fireflycons/hypervcsi/internal/linux/driver"
	"github.com/fireflycons/hypervcsi/internal/linux/kvp"
	"github.com/spf13/cobra"
)

var (
	endpointFlag   string
	urlFlag        string
	driverNameFlag string
	debugAddrFlag  string
	apiKeyFlag     string
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
	rootCmd.Flags().StringVarP(&endpointFlag, "endpoint", "e", "unix:///var/lib/kubelet/plugins/"+driver.DefaultDriverName+"/csi.sock", "CSI endpoint")
	rootCmd.Flags().StringVarP(&urlFlag, "url", "u", "", "URL of khypervprovider Windows Service")
	rootCmd.Flags().StringVarP(&driverNameFlag, "driver-name", "n", driver.DefaultDriverName, "Name for the driver")
	rootCmd.Flags().StringVarP(&debugAddrFlag, "debug-addr", "d", "", "Address to serve the HTTP debug server on")
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
		log.Fatalln(err)
	}

}
