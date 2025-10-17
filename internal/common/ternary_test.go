package common_test

import (
	"testing"

	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/stretchr/testify/require"
)

func TestTernaryf(t *testing.T) {

	var trueCalled, falseCalled bool

	trueVal := func() string {
		trueCalled = true
		return "true"
	}

	falseVal := func() string {
		falseCalled = true
		return "false"
	}

	tests := []struct {
		name      string
		condition bool
		want      string
	}{
		{
			name:      "True branch",
			condition: true,
			want:      "true",
		},
		{
			name:      "False branch",
			condition: false,
			want:      "false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trueCalled = false
			falseCalled = false
			got := common.Ternaryf(tt.condition, trueVal, falseVal)

			if tt.condition {
				require.True(t, trueCalled, "'true' function was not called")
				require.False(t, falseCalled, "'false' function was called")
				require.Equal(t, tt.want, got)
			} else {
				require.False(t, trueCalled, "'true' function was called")
				require.True(t, falseCalled, "'false' function was not called")
				require.Equal(t, tt.want, got)
			}
		})
	}
}
