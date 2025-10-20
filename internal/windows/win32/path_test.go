package win32

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func getShortPath(path string) (string, error) {
	pathUTF16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	buf := make([]uint16, syscall.MAX_PATH)
	n, err := syscall.GetShortPathName(pathUTF16, &buf[0], syscall.MAX_PATH)
	if n == 0 {
		if err != nil {
			return "", err
		}
		return path, nil
	}
	return syscall.UTF16ToString(buf[:n]), nil
}

func TestGetLongPath(t *testing.T) {
	// Create a temp file with a long-ish name to ensure it can have a short alias.
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "longfilename_for_testing_purpose_1234567890_*.txt")
	require.NoError(t, err, "Failed to create temp file")

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	tmpFile.Close()

	longPath, err := filepath.Abs(tmpFile.Name())
	require.NoError(t, err, "Failed to get absolute path")

	shortPath, err := getShortPath(longPath)
	require.NoError(t, err, "Failed to get short path")

	if shortPath == longPath {
		t.Skip("Filesystem does not support 8.3 short paths (short == long), skipping test")
	}

	resolved, err := GetLongPathName(shortPath)
	require.NoError(t, err, "getLongPath() returned error")
	require.Equal(t, longPath, resolved, "Expected long path:\n%s\nGot:\n%s", longPath, resolved)
}
