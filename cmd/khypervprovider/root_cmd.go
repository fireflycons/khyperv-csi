//go:build windows

/*
Copyright © 2025 Alistair Mackay <a.m@ckay.me>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"errors"
	"log"
	"os"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
)

var (
	portFlag   uint32
	apiKeyFlag string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "khypervprovider",
	Short: "Kubernetes Hyper-V CSI Windows Service",
	Long: `
The without arguments form is intended to be invoked by the Windows service control manager`,
	RunE: func(*cobra.Command, []string) error {

		inService, err := svc.IsWindowsService()
		if err != nil {
			log.Fatalf("failed to determine if we are running in service: %v", err)
		}

		if !inService {
			return errors.New("this is the Windows service entrypoint - it can only be called by Start Service")
		}

		runService(constants.ServiceName, false)
		return nil
	},
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

	rootCmd.Flags().Uint32Var(&portFlag, "port", constants.DefaultServicePort, "Port services listens on")
	rootCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "API key to assert on REST interface")
	rootCmd.Flags().StringVar(&certFlag, "cert", "", "Certificate to use for HTTPS serving")
	rootCmd.Flags().StringVar(&keyFlag, "key", "", "Key to use for HTTPS serving")
	rootCmd.Flags().StringVar(&pvDirectoryFlag, "directory", "", "Directory to store PV disks in. Omit to have the service choose.")

}
