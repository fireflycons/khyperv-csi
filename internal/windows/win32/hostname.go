//go:build windows

package win32

import (
	"strings"
	"syscall"
	"unsafe"
)

// COMPUTER_NAME_FORMAT corresponds to the Windows API enum for GetComputerNameExW
type computerNameFormat uint32

const (
	ComputerNameNetBIOS                   computerNameFormat = 0
	ComputerNameDnsHostname               computerNameFormat = 1
	ComputerNameDnsDomain                 computerNameFormat = 2
	ComputerNameDnsFullyQualified         computerNameFormat = 3
	ComputerNamePhysicalNetBIOS           computerNameFormat = 4
	ComputerNamePhysicalDnsHostname       computerNameFormat = 5
	ComputerNamePhysicalDnsDomain         computerNameFormat = 6
	ComputerNamePhysicalDnsFullyQualified computerNameFormat = 7
)

// GetHostname returns the FQDN (Fully Qualified Domain Name) if the machine is domain-joined,
// otherwise it returns the hostname.
func GetHostname() (string, error) {
	// Load kernel32.dll
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetComputerNameExW := kernel32.NewProc("GetComputerNameExW")

	// First call to get required buffer size
	var size uint32
	ret, _, _ := procGetComputerNameExW.Call(
		uintptr(ComputerNamePhysicalDnsFullyQualified),
		uintptr(0),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 && size == 0 {
		// fallback to simpler name if FQDN not available
		return fallbackHostname()
	}

	// Allocate buffer for name
	buf := make([]uint16, size)
	ret, _, _ = procGetComputerNameExW.Call(
		uintptr(ComputerNamePhysicalDnsFullyQualified),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return fallbackHostname()
	}

	name := syscall.UTF16ToString(buf)
	if name == "" {
		return fallbackHostname()
	}

	return strings.ToLower(name), nil
}

// fallbackHostname uses the simpler ComputerNameDnsHostname if the full FQDN isn’t available
func fallbackHostname() (string, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetComputerNameExW := kernel32.NewProc("GetComputerNameExW")

	var size uint32
	_, _, _ = procGetComputerNameExW.Call(
		uintptr(ComputerNamePhysicalDnsHostname),
		uintptr(0),
		uintptr(unsafe.Pointer(&size)),
	)

	buf := make([]uint16, size)
	ret, _, err := procGetComputerNameExW.Call(
		uintptr(ComputerNamePhysicalDnsHostname),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return "", err
	}

	return strings.ToLower(syscall.UTF16ToString(buf)), nil
}
