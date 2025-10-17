//go:build windows

package win32

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// IsServiceRunning checks if the given Windows service is running.
// Returns (true, nil) if running, (false, nil) if installed but not running,
// and (false, error) if the service is not installed or cannot be queried.
func IsServiceRunning(name ServiceName) (bool, error) {
	mgrHandle, err := windows.OpenSCManager(
		nil, // local machine
		nil, // SERVICES_ACTIVE_DATABASE
		windows.SC_MANAGER_CONNECT,
	)
	if err != nil {
		return false, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer func() {
		_ = windows.CloseServiceHandle(mgrHandle)
	}()

	// Convert service name safely
	namePtr, err := windows.UTF16PtrFromString(string(name))
	if err != nil {
		return false, fmt.Errorf("invalid service name %q: %w", name, err)
	}

	// Open the service with query rights only
	svcHandle, err := windows.OpenService(
		mgrHandle,
		namePtr,
		windows.SERVICE_QUERY_STATUS,
	)

	if err != nil {
		return false, fmt.Errorf("service %q not found: %w", name, err)
	}
	defer func() {
		_ = windows.CloseServiceHandle(svcHandle)
	}()

	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(svcHandle, &status); err != nil {
		return false, fmt.Errorf("failed to query service %q: %w", name, err)
	}

	return status.CurrentState == windows.SERVICE_RUNNING, nil
}
