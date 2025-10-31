//go:build windows

package psmodule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLatestPackage(t *testing.T) {

	tests := []struct {
		name     string
		packages []string
		expected string
	}{
		{
			name:     "No packages (empty list)",
			packages: []string{},
			expected: "",
		},
		{
			name:     "No packages (nil list)",
			packages: nil,
			expected: "",
		},
		{
			name: "One package",
			packages: []string{
				"cmd/khypervprovider/psmodule/khyperv-csi.1.0.0.nupkg",
			},
			expected: "cmd/khypervprovider/psmodule/khyperv-csi.1.0.0.nupkg",
		},
		{
			name: "Multiple packages",
			packages: []string{
				"cmd/khypervprovider/psmodule/khyperv-csi.1.1.0.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.1.0.0.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.2.0.1.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.3.99.2.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.4.0.1.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.0.0.1.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.1.40.7.nupkg",
				"cmd/khypervprovider/psmodule/khyperv-csi.2.7.8.nupkg",
			},
			expected: "cmd/khypervprovider/psmodule/khyperv-csi.4.0.1.nupkg",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := getLatestPackage(test.packages)
			require.Equal(t, test.expected, actual)
		})
	}
}
