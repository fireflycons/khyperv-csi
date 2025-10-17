//go:build windows

package win32

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func GetSystemLocale() (string, error) {
	const localeNameMax = 85 // LOCALE_NAME_MAX_LENGTH

	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procGetUserDefaultLocaleName := kernel32.NewProc("GetUserDefaultLocaleName")

	if err := kernel32.Load(); err != nil {
		return "", fmt.Errorf("failed to load kernel32.dll: %w", err)
	}

	buf := make([]uint16, localeNameMax)
	ret, _, err := procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(localeNameMax),
	)

	if ret == 0 {
		// err is often syscall.Errno(0) even on failure, so check both
		if !errors.Is(err, windows.ERROR_SUCCESS) && err != nil {
			return "", fmt.Errorf("GetUserDefaultLocaleName failed: %w", err)
		}
		return "", fmt.Errorf("GetUserDefaultLocaleName failed")
	}

	return windows.UTF16ToString(buf), nil
}
