//go:build windows

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var stateMap = map[svc.State]string{
	svc.Stopped:         "Stopped",
	svc.StartPending:    "Start pending",
	svc.StopPending:     "Stop pending",
	svc.Running:         "Running",
	svc.ContinuePending: "Continue pending",
	svc.PausePending:    "Pause pending",
	svc.Paused:          "Paused",
}

// assertElevatedPrivilege ensures the process is running with
// elevated (administrator) privileges. If not, it exits with an error.
func assertElevatedPrivilege() {

	var token windows.Token

	// Open the access token for the current process
	err := windows.OpenProcessToken(
		windows.CurrentProcess(),
		windows.TOKEN_QUERY,
		&token,
	)
	if err != nil {
		log.Fatalf("failed to open process token: %v", err)
	}

	// --- Check elevation ---
	type tokenElevation struct {
		TokenIsElevated uint32
	}

	var elevation tokenElevation
	var outLen uint32
	if err := windows.GetTokenInformation(
		token,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&outLen,
	); err != nil {
		_ = token.Close()
		log.Fatalf("Failed to query token elevation: %v", err)
	}

	if elevation.TokenIsElevated == 0 {
		_ = token.Close()
		log.Fatalf("Administrator privileges required — please run this program As Administrator.")
	}
	_ = token.Close()
}

var serviceMessages = map[string]string{
	constants.HyperVServiceName: "Hyper-V compute service not found. Is this supposed to be a Hyper-V server?",
	constants.ServiceName:       "Kubernetes Persistent Volume service is not installed.",
}

// assertService asserts that a given service is installed locally,
// else exits with an error message.
func assertService(m *mgr.Mgr, name string) error {

	if s, err := m.OpenService(name); err != nil {
		msg, ok := serviceMessages[name]
		if !ok {
			msg = fmt.Sprintf("service %s not found", name)
		}
		return errors.New(msg)
	} else {
		_ = s.Close()
	}

	return nil
}

// mustGetExePath returns fully qualified path to this process's binary or exits with an error messsage.
func mustGetExePath() string {

	prog := os.Args[0]

	p, err := filepath.Abs(prog)

	if err != nil {
		log.Fatalf("Failed to get absolute path for %s: %v", prog, err)
	}

	fi, err := os.Stat(p)

	if err == nil {
		if !fi.Mode().IsDir() {
			return p
		}

		log.Fatalf("%s is directory", p)
	}

	if filepath.Ext(p) == "" {

		var fi os.FileInfo

		p += ".exe"
		fi, err = os.Stat(p)

		if err == nil {

			if !fi.Mode().IsDir() {
				return p
			}

			log.Fatalf("%s is directory", p)
		}
	}

	return p
}
