//go:build windows

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/fireflycons/hypervcsi/internal/windows/win32"
)

var (
	serverFqdn = func() string {
		if s, err := win32.GetHostname(); err == nil {
			return s
		}
		return "unknown.example.com"
	}()
	countryCode = func() string {
		locale, err := win32.GetSystemLocale()
		if err != nil {
			return "US"
		}
		return strings.Split(locale, "-")[1]
	}()
)

func generateCertificates(certsPath string) error {

	const caLifetimeYears = 10

	if certificatesExist(certsPath) {
		return nil
	}

	ipAddreses, err := win32.GetIPv4Addresses()
	if err != nil {
		return err
	}

	caName := DecodeDistinguishedNameRFC4514(distinguishedNameCAFlag)
	if caName["CN"] == nil {
		return errors.New("CA cerfiticate must have Common Name")
	}
	certName := DecodeDistinguishedNameRFC4514(distinguishedNameCertFlag)
	if certName["CN"] == nil {
		return errors.New("server cerfiticate must have Common Name")
	}

	// === 1. Generate CA private key ===
	caPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// === 2. Create CA certificate template ===
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       caName["O"],
			OrganizationalUnit: caName["OU"],
			Country:            caName["C"],
			Locality:           caName["L"],
			Province:           caName["ST"],
			StreetAddress:      caName["STREET"],
			CommonName:         caName["CN"][0],
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(caLifetimeYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// === 3. Self-sign CA certificate ===
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPriv.PublicKey, caPriv)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	//nolint:govet // intentional redeclaration of err
	if err := writePemFile(filepath.Join(certsPath, "ca.crt"), "CERTIFICATE", caDER); err != nil {
		return err
	}

	//nolint:govet // intentional redeclaration of err
	if err := writePemFile(filepath.Join(certsPath, "ca.key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(caPriv)); err != nil {
		return err
	}

	fmt.Println("Generated CA certificate and key.")

	// === 4. Generate server private key ===
	serverPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate server private key: %w", err)
	}

	netIps := []net.IP{}
	From(ipAddreses).
		Select(func(i any) any {
			return net.ParseIP(i.(string))
		}).
		Where(func(i any) bool {
			return i.(net.IP) != nil
		}).ToSlice(&netIps)

	// === 5. Create server certificate template ===
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:       certName["O"],
			OrganizationalUnit: certName["OU"],
			Country:            certName["C"],
			Locality:           certName["L"],
			Province:           certName["ST"],
			StreetAddress:      certName["STREET"],
			CommonName:         certName["CN"][0],
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{serverFqdn, "localhost"},
		IPAddresses: netIps,
	}

	// === 6. Sign server certificate with CA ===
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverPriv.PublicKey, caPriv)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %w", err)
	}

	if err := writePemFile(filepath.Join(certsPath, "server.crt"), "CERTIFICATE", serverDER); err != nil {
		return err
	}
	if err := writePemFile(filepath.Join(certsPath, "server.key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverPriv)); err != nil {
		return err
	}

	fmt.Println("Generated server certificate and key signed by CA.")
	return nil
}

func writePemFile(filename, blockType string, bytes []byte) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: bytes}); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}
	return nil
}

func certificatesExist(path string) bool {

	for _, f := range []string{"ca", "server"} {
		for _, e := range []string{".crt", ".key"} {
			//nolint:govet // intentional redeclaration of path
			path := filepath.Join(path, f+e)
			if !fileExists(path) {
				return false
			}
		}
	}

	return true
}

func autoDistinguishedName(cn, org, country string) string {

	dn := []string{}

	if cn != "" {
		dn = append(dn, "CN="+cn)
	}

	if org != "" {
		dn = append(dn, "O="+org)
	}

	if country != "" {
		dn = append(dn, "C="+country)
	}

	return strings.Join(dn, ",")
}

// DecodeDistinguishedNameRFC4514 parses a DN (Distinguished Name)
// string into a map[string][]string according to RFC 4514 rules.
//
// Supported X.500 attribute short names:
// CN, L, ST, O, OU, C, STREET
//
// Returns nil slices for attributes not present.
func DecodeDistinguishedNameRFC4514(dn string) map[string][]string {
	// Known X.500 attributes we care about
	attrOrder := []string{"CN", "L", "ST", "O", "OU", "C", "STREET"}

	result := make(map[string][]string)
	for _, attr := range attrOrder {
		result[attr] = nil
	}

	// Split into RDNs (respect escaped commas)
	rdns := splitEscaped(dn, ',')

	for _, rdn := range rdns {
		// Split multi-valued RDNs like CN=foo+OU=bar
		parts := splitEscaped(rdn, '+')
		for _, part := range parts {
			key, val, ok := splitKeyVal(part)
			if !ok {
				continue
			}
			key = strings.ToUpper(strings.TrimSpace(key))
			val = strings.TrimSpace(val)
			val = unescape(val)
			if _, known := result[key]; known {
				result[key] = append(result[key], val)
			}
		}
	}

	return result
}

// splitEscaped splits a string on unescaped separators.
func splitEscaped(s string, sep rune) []string {
	var out []string
	var part strings.Builder
	escaped := false

	for _, r := range s {
		switch {
		case escaped:
			part.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == sep:
			out = append(out, strings.TrimSpace(part.String()))
			part.Reset()
		default:
			part.WriteRune(r)
		}
	}
	out = append(out, strings.TrimSpace(part.String()))
	return out
}

// splitKeyVal splits "key=value" into key and value, handling escaped '='.
func splitKeyVal(s string) (key, val string, ok bool) {
	var b strings.Builder
	escaped := false
	for i, r := range s {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '=':
			key = strings.TrimSpace(b.String())
			val = strings.TrimSpace(s[i+1:])
			return key, val, true
		default:
			b.WriteRune(r)
		}
	}
	return "", "", false
}

// unescape handles RFC 4514 escaping: converts "\," -> ",", "\+" -> "+", "\=" -> "=" etc.
func unescape(s string) string {
	var b strings.Builder
	escaped := false
	for _, r := range s {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
