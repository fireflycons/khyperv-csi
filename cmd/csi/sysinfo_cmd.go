//go:build linux

package main

import (
	"fmt"

	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/spf13/cobra"
)

var sysinfoCmd = &cobra.Command{
	Use:   "sysinfo",
	Short: "Print system information and exit",
	Run: func(*cobra.Command, []string) {
		fmt.Println(common.GetSystemInfo())
	},
}

func init() {
	rootCmd.AddCommand(sysinfoCmd)
}
