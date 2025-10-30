package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Version    = "0.0.1+dev"
	CommitHash = "unknown"
	BuildDate  = "unknown"
)

func PrintVersion() {
	exe := filepath.Base(os.Args[0])

	app := func() string {
		if runtime.GOOS == "windows" {
			return "Windows service for Hyper-V"
		}
		return "Cluster CSI Plugin"
	}()

	fmt.Printf(`
%s %s

Hyper-V Container Storage Interface: %s

Commit Hash: %s
Build Date : %s

`,
		exe,
		Version,
		app,
		CommitHash,
		BuildDate,
	)
}
