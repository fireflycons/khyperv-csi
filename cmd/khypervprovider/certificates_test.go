//go:build windows

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateCertificates(t *testing.T) {
	// Simulate user input:
	// (Each corresponds to one prompt, pressing Enter accepts default)
	input := bytes.NewBufferString(`
My CA Org
US
My Root CA
My Server Org
US
myserver.local
myserver.local,localhost
192.168.0.1
`)

	// Clean up generated files after test
	defer func() {
		_ = os.Remove("ca.crt")
		_ = os.Remove("ca.key")
		_ = os.Remove("server.crt")
		_ = os.Remove("server.key")
	}()

	if err := generateCertificates(".", input); err != nil {
		require.NoError(t, err, "generateCertificates failed")
	}

	// Validate that files exist
	files := []string{"ca.crt", "ca.key", "server.crt", "server.key"}
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			require.NoError(t, err, "expected %s to exist", f)
		}
	}
}
