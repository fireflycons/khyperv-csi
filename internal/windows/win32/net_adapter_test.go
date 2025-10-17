//go:build windows

package win32

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetIPAddresses(t *testing.T) {

	ips, err := GetIPv4Addresses()

	require.NoError(t, err)
	require.NotEmpty(t, ips, "Received no IP addresses")
	fmt.Printf("Local adapter IPs: %v\n", ips)
}
