//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

func GetCapacity(runner powershell.Runner, pvstore string) (int64, error) {

	capacity := &models.GetCapacityResponse{}

	resp, err := executeWithReturn(
		runner,
		capacity,
		powershell.NewCmdlet("Get-PVCapacity", map[string]any{"PVStore": pvstore}),
	)

	if err != nil {
		return 0, err
	}

	return resp.FreeSpaceBytes, nil
}
