//go:build linux

package driver

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHypervDiskByID(t *testing.T) {

	tests := []struct {
		DiskIdentifier string
		ById           string
	}{
		{
			DiskIdentifier: "53B8F8D8-38B2-4479-8CA5-842C2CD44861",
			ById:           filepath.Join(diskIDPath, "scsi-360022480d8f8b853b238842c2cd44861"),
		},
		{
			DiskIdentifier: "E731E2AD-63D1-49CB-8F41-F4FD6275783E",
			ById:           filepath.Join(diskIDPath, "scsi-360022480ade231e7d163f4fd6275783e"),
		},
		{
			DiskIdentifier: "158F93EA-5C87-4DDA-A0F9-D85F5765DE1F",
			ById:           filepath.Join(diskIDPath, "scsi-360022480ea938f15875cd85f5765de1f"),
		},
		{
			DiskIdentifier: "F37E8C32-8063-4027-974F-43B258C5F9E2",
			ById:           filepath.Join(diskIDPath, "scsi-360022480328c7ef3638043b258c5f9e2"),
		},
	}

	for _, test := range tests {
		t.Run(test.DiskIdentifier, func(t *testing.T) {
			id, err := hypervDiskByID(test.DiskIdentifier)
			require.NoError(t, err)
			require.Equal(t, test.ById, id)
		})
	}
}
