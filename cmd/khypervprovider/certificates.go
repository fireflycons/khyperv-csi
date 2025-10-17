//go:build windows

package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/windows/win32"
)

func generateCertificates(certsPath string, input io.Reader) error {

	const caLifetimeYears = 10

	reader := bufio.NewReader(input)

	if certificatesExist(certsPath) && promptString(reader, "Certificates exist. Use them?", "yes") == "yes" {
		return nil
	}

	serverFqdn, err := win32.GetHostname()

	if err != nil {
		return err
	}

	ipAddreses, err := win32.GetIPv4Addresses()
	if err != nil {
		return err
	}

	countryCode := func() string {
		//nolint:govet // intentional redeclaration of err
		locale, err := win32.GetSystemLocale()
		if err != nil {
			return "US"
		}
		return strings.Split(locale, "-")[1]
	}()

	// === Prompt for CA info ===
	fmt.Println("\n--- CA Certificate Information ---")
	caOrg := promptString(reader, "CA Organization", "Example CA Org")
	caCountry := promptString(reader, "CA Country", countryCode)
	caCommonName := promptString(reader, "CA Common Name", "Example Root CA")

	// === Prompt for server info ===
	fmt.Println("\n--- Server Certificate Information ---")
	srvOrg := promptString(reader, "Server Organization", constants.ServiceName)
	srvCountry := promptString(reader, "Server Country", countryCode)
	srvCommonName := promptString(reader, "Server Common Name", serverFqdn)
	srvDNS := promptList(reader, "DNS Names", []string{serverFqdn, "localhost"})
	srvIP := promptList(reader, "IP Addreses", ipAddreses)

	// === 1. Generate CA private key ===
	caPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// === 2. Create CA certificate template ===
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{caOrg},
			Country:      []string{caCountry},
			CommonName:   caCommonName,
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
	From(srvIP).
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
			Organization: []string{srvOrg},
			Country:      []string{srvCountry},
			CommonName:   srvCommonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    srvDNS,
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

// promptString prompts the user for a single string value, showing a default.
func promptString(reader *bufio.Reader, prompt, def string) string {
	fmt.Printf("%s [%s]: ", prompt, def)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

// promptList prompts the user for a comma-separated list of values, showing defaults.
func promptList(reader *bufio.Reader, prompt string, defaults []string) []string {
	fmt.Printf("%s [%s]: ", prompt, strings.Join(defaults, ", "))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaults
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		val := strings.TrimSpace(p)
		if val != "" {
			result = append(result, val)
		}
	}
	return result
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
