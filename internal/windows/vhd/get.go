//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

func GetByName(runner powershell.Runner, store, name string) (*models.GetVHDResponse, error) {

	response, err := executeWithReturn(
		runner,
		&models.GetVHDResponse{},
		powershell.NewCmdlet(
			"Get-PVDisk",
			map[string]any{
				"PVStore": store,
				"Name":    name,
				"AsJson": nil,
			},
		),
	)

	return response, err
}

func GetByID(runner powershell.Runner, store, id string) (*models.GetVHDResponse, error) {

	return executeWithReturn(
		runner,
		&models.GetVHDResponse{},
		powershell.NewCmdlet(
			"Get-PVDisk",
			map[string]any{
				"PVStore": store,
				"Id":      id,
				"AsJson": nil,
			},
		),
	)
}

func GetByPath(runner powershell.Runner, fullPath string) (*models.GetVHDResponse, error) {

	return executeWithReturn(
		runner,
		&models.GetVHDResponse{},
		powershell.NewCmdlet(
			"Get-PVDisk",
			map[string]any{
				"PVStore": "",
				"Path":    fullPath,
				"AsJson": nil,
			},
		),
	)
}
