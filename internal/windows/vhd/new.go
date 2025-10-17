//go:build windows

package vhd

import (
	"errors"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

var ErrCapacityExhausted = errors.New("capacity exhasted")

// New creates a new VHD file in the given directory with the given size.
// The filename of the VHD is set to the DiskIdentifier property returned by creation.
func New(runner powershell.Runner, name, pvStore string, size int64) (*models.GetVHDResponse, error) {

	return executeWithReturn(
		runner,
		&models.GetVHDResponse{},
		powershell.NewCmdlet(
			"New-PVDisk",
			map[string]any{
				"Name":    name,
				"PVStore": pvStore,
				"Size":    size,
				"VHDType": constants.VhdType,
			},
		),
	)
}
