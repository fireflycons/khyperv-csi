//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

func Delete(runner powershell.Runner, store, diskId string) error {

	return execute(
		runner,
		powershell.NewCmdlet(
			"Remove-PVDisk",
			map[string]any{
				"PVStore": store,
				"Id":      diskId,
			},
		),
	)
}
