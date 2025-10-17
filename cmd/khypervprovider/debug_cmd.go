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
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run in foreground for debugging",
	Run:   executeDebug,
}

func init() {

	debugCmd.Flags().Uint32VarP(&portFlag, "port", "p", constants.DefaultServicePort, "Port services listens on")
	debugCmd.Flags().StringVarP(&apiKeyFlag, "api-key", "k", "debug", "API key to assert on REST interface")

	rootCmd.AddCommand(debugCmd)
}

func executeDebug(*cobra.Command, []string) {

	runService(constants.ServiceName, true)
}
