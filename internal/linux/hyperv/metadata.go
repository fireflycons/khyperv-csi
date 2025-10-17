//go:build linux

package hyperv

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const (
	kvpDir     = "/var/lib/hyperv"
	keySize    = 512
	valueSize  = 2048
	recordSize = keySize + valueSize
)

// ReadMeta scans the given Hyper-V KVP pool file for a key and returns its value if found.
//
//	vmname, err := ReadMeta(3, "VirtualMachineName")
//	vmid, err := ReadMeta(3, "VirtualMachineId")
func ReadMeta(poolNumber int, key string) (string, error) {
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
