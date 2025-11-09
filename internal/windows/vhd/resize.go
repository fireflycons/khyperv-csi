//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

// New creates a new VHD file in the given directory with the given size.
// The filename of the VHD is set to the DiskIdentifier property returned by creation.
func Resize(runner powershell.Runner, pvStore, id string, size int64) (*models.GetVHDResponse, error) {

	return executeWithReturn(
		runner,
		&models.GetVHDResponse{},
		powershell.NewCmdlet(
			"Resize-PVDisk",
			map[string]any{
				"Id":      id,
				"PVStore": pvStore,
				"Size":    size,
			},
		),
	)
}
