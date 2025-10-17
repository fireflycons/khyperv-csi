//go:build windows

package win32

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSystemLocale(t *testing.T) {

	l, err := GetSystemLocale()
	require.NoError(t, err)
	fmt.Println("Got system locale:", l)
}
