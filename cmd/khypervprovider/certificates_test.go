//go:build windows

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCertificates(t *testing.T) {

	// Clean up generated files after test
	defer func() {
		_ = os.Remove("ca.crt")
		_ = os.Remove("ca.key")
		_ = os.Remove("server.crt")
		_ = os.Remove("server.key")
	}()

	if err := generateCertificates("."); err != nil {
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

func TestDecodeDistinguishedNameRFC4514(t *testing.T) {
	tests := []struct {
		name     string
		dn       string
		expected map[string][]string
	}{
		{
			name: "Simple DN",
			dn:   "CN=example.com, O=Example Corp, C=US",
			expected: map[string][]string{
				"CN":     {"example.com"},
				"L":      nil,
				"ST":     nil,
				"O":      {"Example Corp"},
				"OU":     nil,
				"C":      {"US"},
				"STREET": nil,
			},
		},
		{
			name: "Simple DN, no whitespace",
			dn:   "CN=example.com,O=Example Corp,C=US",
			expected: map[string][]string{
				"CN":     {"example.com"},
				"L":      nil,
				"ST":     nil,
				"O":      {"Example Corp"},
				"OU":     nil,
				"C":      {"US"},
				"STREET": nil,
			},
		},
		{
			name: "Multi-valued RDN and full attributes",
			dn:   "CN=John Doe+OU=Dev, O=Example Corp, L=NY, ST=NY, C=US, STREET=123 Main St",
			expected: map[string][]string{
				"CN":     {"John Doe"},
				"L":      {"NY"},
				"ST":     {"NY"},
				"O":      {"Example Corp"},
				"OU":     {"Dev"},
				"C":      {"US"},
				"STREET": {"123 Main St"},
			},
		},
		{
			name: "Escaped comma and equals",
			dn:   `CN=Foo\, Bar, O=Corp\=Inc., C=US`,
			expected: map[string][]string{
				"CN":     {"Foo, Bar"},
				"L":      nil,
				"ST":     nil,
				"O":      {"Corp=Inc."},
				"OU":     nil,
				"C":      {"US"},
				"STREET": nil,
			},
		},
		{
			name: "Extra spaces and mixed case keys",
			dn:   " cn = Alice , o = Wonderland  , c = GB ",
			expected: map[string][]string{
				"CN":     {"Alice"},
				"L":      nil,
				"ST":     nil,
				"O":      {"Wonderland"},
				"OU":     nil,
				"C":      {"GB"},
				"STREET": nil,
			},
		},
		{
			name: "Multiple values for same attribute",
			dn:   "OU=Dev, OU=Ops, CN=John Doe",
			expected: map[string][]string{
				"CN":     {"John Doe"},
				"L":      nil,
				"ST":     nil,
				"O":      nil,
				"OU":     {"Dev", "Ops"},
				"C":      nil,
				"STREET": nil,
			},
		},
		{
			name: "Auto DN",
			dn:   autoDistinguishedName("example.com", "Example Corp", "US"),
			expected: map[string][]string{
				"CN":     {"example.com"},
				"L":      nil,
				"ST":     nil,
				"O":      {"Example Corp"},
				"OU":     nil,
				"C":      {"US"},
				"STREET": nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeDistinguishedNameRFC4514(tt.dn)

			for k, exp := range tt.expected {
				assert.Equal(t, exp, got[k], "unexpected value for key %s", k)
			}
		})
	}
}
