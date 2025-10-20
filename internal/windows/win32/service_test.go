//go:build windows

package win32

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsServiceRunning(t *testing.T) {

	tests := []struct {
		name            string
		serviceName     string
		shouldBeRunning bool
		expectError     bool
	}{
		{
			name:            "EventLog service is running",
			serviceName:     "EventLog",
			shouldBeRunning: true,
		},
		{
			// It is highly unlikely anyone would be running this
			name:            "BITS service is not running",
			serviceName:     "BITS",
			shouldBeRunning: false,
		},
		{
			name:        "Non-existent service is error",
			serviceName: "sdagsdbgesgerferfdfgbdv",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isRunning, err := IsServiceRunning(ServiceName(test.serviceName))

			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if test.shouldBeRunning {
				require.True(t, isRunning, "%s should be running, but isn't", test.serviceName)
			} else {
				require.False(t, isRunning, "%s should not running, but is", test.serviceName)
			}
		})
	}
}
