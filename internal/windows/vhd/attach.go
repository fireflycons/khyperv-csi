//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

func Attach(runner powershell.Runner, store, diskId, nodeId string) (*models.AttachedDrive, error) {

	// Get the disk path from the ID
	disk, err := GetByID(runner, store, diskId)
	if err != nil {
		return nil, err
	}

	if disk == nil {
		return nil, ErrInvalidDiskId
	}

	// TODO Assert disk is not attached to another node (FAILED_PRECONDITION)

	drive := &models.AttachedDrive{}

	return executeWithReturn(
		runner,
		drive,
		powershell.NewCmdlet(
			"Mount-PVDisk",
			map[string]any{
				"VMId":     nodeId,
				"DiskPath": disk.Path,
			},
		),
	)
}
