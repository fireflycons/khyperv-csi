package common_test

import (
	"testing"

	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		inputBytes int64
		want       string
	}{
		{
			name:       "zero bytes",
			inputBytes: 0,
			want:       "0",
		},
		{
			name:       "bytes",
			inputBytes: 512,
			want:       "512",
		},
		{
			name:       "kilobytes",
			inputBytes: 2048,
			want:       "2Ki",
		},
		{
			name:       "megabytes",
			inputBytes: 5 * 1024 * 1024,
			want:       "5Mi",
		},
		{
			name:       "gigabytes",
			inputBytes: 10 * 1024 * 1024 * 1024,
			want:       "10Gi",
		},
		{
			name:       "terabytes",
			inputBytes: 3 * 1024 * 1024 * 1024 * 1024,
			want:       "3Ti",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, common.FormatBytes(tt.inputBytes))
		})
	}
}
