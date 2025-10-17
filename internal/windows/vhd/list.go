//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

// List lists all volumes created in the volume store
func List(runner powershell.Runner, store string, maxEntries int32, nextToken string) (*models.ListVHDResponse, error) {

	volumes, err := executeWithReturn(
		runner,
		&models.ListVHDResponse{},
		powershell.NewCmdlet(
			"Get-PVDisks",
			map[string]any{
				"PVStore":    store,
				"MaxEntries": maxEntries,
				"NextToken":  nextToken,
			}),
	)

	if err != nil {
		return nil, err
	}

	return volumes, nil
}
