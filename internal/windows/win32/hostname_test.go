//go:build windows

package win32

import (
	"os"
	"strings"
	"testing"

	"github.com/julien040/go-ternary"
	"github.com/stretchr/testify/require"
)

func TestGetHostname(t *testing.T) {

	computerName := os.Getenv("COMPUTERNAME")
	domain := os.Getenv("USERDNSDOMAIN")

	expectedFqdn := ternary.If(
		domain != "",
		strings.ToLower(computerName+"."+domain),
		strings.ToLower(computerName),
	)

	actualFqdn, err := GetHostname()

	require.NoError(t, err)
	require.Equal(t, expectedFqdn, actualFqdn)
}
