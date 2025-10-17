//go:build windows

package vhd

import "github.com/fireflycons/hypervcsi/internal/windows/powershell"

func Detach(runner powershell.Runner, store, diskId, nodeId string) error {

	// Get the disk path from the ID
	disk, err := GetByID(runner, store, diskId)
	if err != nil {
		return err
	}

	if disk == nil {
		return ErrInvalidDiskId
	}

	return execute(
		runner,
		powershell.NewCmdlet(
			"Dismount-PVDisk",
			map[string]any{
				"VMId":     nodeId,
				"DiskPath": disk.Path,
			},
		),
	)
}
