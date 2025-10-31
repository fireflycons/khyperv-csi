//go:build linux

package main

import (
	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Run: func(*cobra.Command, []string) {
		common.PrintVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
