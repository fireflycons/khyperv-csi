//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

func GetStorePath(runner powershell.Runner) (string, error) {

	store := &models.GetPVStoreResponse{}

	_, err := executeWithReturn(
		runner,
		store,
		powershell.NewCmdlet("Get-PVStore", nil),
	)

	if err != nil {
		return "", err
	}

	return store.PVStore, nil
}
