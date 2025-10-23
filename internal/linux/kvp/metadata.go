//go:build linux

package kvp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const (
	VM_NAME_KEY = "VirtualMachineName"
	VM_ID_KEY   = "VirtualMachineId"
)

const (
	kvpDir     = "/var/lib/hyperv"
	keySize    = 512
	valueSize  = 2048
	recordSize = keySize + valueSize
)

// MetadataService defines methods to interact with the Hyper-V KVP metadata service.
type MetadataService interface {

	// IsPresent checks if the Hyper-V KVP metadata service is available.
	IsPresent() bool

	// Find searches all KVP pools for the given key and returns its value.
	Find(key string) (string, error)

	// Read reads the value for a given key from a specific KVP pool.
	Read(poolNumber int, key string) (string, error)
}

type kvpMetadataService struct{}

// New creates a new instance of the Hyper-V KVP metadata service.
func New() *kvpMetadataService {
	return &kvpMetadataService{}
}

// IsPresent checks if the Hyper-V KVP metadata service is available.
func (k *kvpMetadataService) IsPresent() bool {
	_, err := os.Stat(kvpDir)
	return err == nil
}

// Find searches all KVP pools for the given key and returns its value.
func (k *kvpMetadataService) Find(key string) (string, error) {

	results := make([]string, 0, 1)

	for _, poolNum := range getPoolNumbers() {
		val, err := k.Read(poolNum, key)
		if err == nil {
			results = append(results, val)
		}
	}

	switch len(results) {
	case 0:
		return "", fmt.Errorf("key %q not found in any pool", key)
	case 1:
		return results[0], nil
	default:
		return "", fmt.Errorf("multiple values found for key %q in different pools", key)
	}
}

// Read scans the given Hyper-V KVP pool file for a key and returns its value if found.
//
//	vmname, err := Read(3, "VirtualMachineName")
//	vmid, err := Read(3, "VirtualMachineId")
func (k *kvpMetadataService) Read(poolNumber int, key string) (string, error) {
	poolFile := filepath.Join(kvpDir, fmt.Sprintf(".kvp_pool_%d", poolNumber))

	f, err := os.Open(poolFile)
	if err != nil {
		return "", fmt.Errorf("failed to open pool file %s: %w", poolFile, err)
	}
	defer f.Close()

	buf := make([]byte, recordSize)

	for {
		_, err := f.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("error reading pool file: %w", err)
		}

		// Check if record is empty
		if bytes.Equal(buf, make([]byte, recordSize)) {
			continue
		}

		// Extract key and value
		keyBytes := buf[:keySize]
		valBytes := buf[keySize:]

		decodedKey := decodeCString(keyBytes)
		if decodedKey == "" {
			continue
		}

		if decodedKey == key {
			decodedVal := decodeCString(valBytes)
			return decodedVal, nil
		}
	}

	return "", fmt.Errorf("key %q not found in pool %d", key, poolNumber)
}

// decodeCString decodes a C-style NULL-terminated string from a fixed buffer.
// The KVP pool uses UTF-8 encoding padded with NULL bytes.
func decodeCString(buf []byte) string {
	n := bytes.IndexByte(buf, 0x00)
	if n == -1 {
		n = len(buf)
	}
	return string(buf[:n])
}

var pools []int

func getPoolNumbers() []int {
	if len(pools) == 0 {
		for i := 0; ; i++ {
			poolFile := filepath.Join(kvpDir, fmt.Sprintf(".kvp_pool_%d", i))
			if _, err := os.Stat(poolFile); os.IsNotExist(err) {
				break
			}
			pools = append(pools, i)
		}
	}
	return pools
}
