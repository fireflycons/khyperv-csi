//go:build windows

package win32

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	iphlpapi                 = syscall.NewLazyDLL("iphlpapi.dll")
	procGetAdaptersAddresses = iphlpapi.NewProc("GetAdaptersAddresses")
)

const (
	AF_INET                  = 2
	GAA_FLAG_SKIP_ANYCAST    = 0x2
	GAA_FLAG_SKIP_MULTICAST  = 0x4
	GAA_FLAG_SKIP_DNS_SERVER = 0x8

	IF_OPER_STATUS_UP = 1
)

type ipAdapterAddresses struct {
	Length                uint32
	IfIndex               uint32
	Next                  *ipAdapterAddresses
	AdapterName           *byte
	FirstUnicastAddress   *ipAdapterUnicastAddress
	FirstAnycastAddress   uintptr
	FirstMulticastAddress uintptr
	FirstDnsServerAddress uintptr
	DnsSuffix             *uint16
	Description           *uint16
	FriendlyName          *uint16
	PhysicalAddress       [8]byte
	PhysicalAddressLength uint32
	Flags                 uint32
	Mtu                   uint32
	IfType                uint32
	OperStatus            uint32
	Ipv6IfIndex           uint32
	ZoneIndices           [16]uint32
	FirstPrefix           uintptr
}

type ipAdapterUnicastAddress struct {
	Length             uint32
	Flags              uint32
	Next               *ipAdapterUnicastAddress
	Address            socketAddress
	PrefixOrigin       uint32
	SuffixOrigin       uint32
	DadState           uint32
	ValidLifetime      uint32
	PreferredLifetime  uint32
	LeaseLifetime      uint32
	OnLinkPrefixLength byte
}

type socketAddress struct {
	Addr *syscall.RawSockaddrAny
	Len  int32
}

// GetIPv4Addresses returns IPv4 addresses for "up" adapters,
// skipping link-local (169.254.x.x) and loopback interfaces.
func GetIPv4Addresses() ([]string, error) {
	var size uint32 = 15000
	buf := make([]byte, size)

	ret, _, _ := procGetAdaptersAddresses.Call(
		uintptr(AF_INET),
		uintptr(GAA_FLAG_SKIP_ANYCAST|GAA_FLAG_SKIP_MULTICAST|GAA_FLAG_SKIP_DNS_SERVER),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)

	const ERROR_BUFFER_OVERFLOW = 111
	if ret == ERROR_BUFFER_OVERFLOW {
		buf = make([]byte, size)
		ret, _, _ = procGetAdaptersAddresses.Call(
			uintptr(AF_INET),
			uintptr(GAA_FLAG_SKIP_ANYCAST|GAA_FLAG_SKIP_MULTICAST|GAA_FLAG_SKIP_DNS_SERVER),
			0,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
		)
	}

	const ERROR_SUCCESS = 0
	if ret != ERROR_SUCCESS {
		return nil, syscall.Errno(ret)
	}

	var results []string
	for aa := (*ipAdapterAddresses)(unsafe.Pointer(&buf[0])); aa != nil; aa = aa.Next {
		// Skip adapters that are not "up"
		if aa.OperStatus != IF_OPER_STATUS_UP {
			continue
		}

		for ua := aa.FirstUnicastAddress; ua != nil; ua = ua.Next {
			sa := ua.Address.Addr
			if sa == nil {
				continue
			}

			if sa.Addr.Family == syscall.AF_INET {
				sa4 := (*syscall.RawSockaddrInet4)(unsafe.Pointer(sa))
				ip := fmt.Sprintf("%d.%d.%d.%d",
					sa4.Addr[0], sa4.Addr[1], sa4.Addr[2], sa4.Addr[3])

				// Skip loopback (127.x.x.x) and link-local (169.254.x.x)
				if sa4.Addr[0] == 127 || (sa4.Addr[0] == 169 && sa4.Addr[1] == 254) {
					continue
				}

				results = append(results, ip)
			}
		}
	}

	return results, nil
}
