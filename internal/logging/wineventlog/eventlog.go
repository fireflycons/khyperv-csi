//go:build windows

package wineventlog

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

type WinEventLogHook struct {
	log          debug.Log
	isRegistered bool
}

func New(isDebug bool) (*WinEventLogHook, error) {

	emptyHook := &WinEventLogHook{}

	var el debug.Log

	if isDebug {
		el = debug.New(constants.ServiceName)

	} else {

		if err := RegisterEventSource(); err != nil {
			return emptyHook, err
		}

		var err error

		el, err = eventlog.Open(constants.ServiceName)

		if err != nil {
			return emptyHook, err
		}
	}

	return &WinEventLogHook{log: el, isRegistered: true}, nil
}

func (*WinEventLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *WinEventLogHook) Fire(e *logrus.Entry) error {

	if !h.isRegistered {
		return nil
	}

	msg, err := e.String()
	if err != nil {
		return err
	}

	// TODO Assign message numbers
	switch e.Level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		return h.log.Error(1, msg)
	case logrus.WarnLevel:
		return h.log.Warning(1, msg)
	default:
		return h.log.Info(1, msg)
	}
}

func RegisterEventSource() error {
	err := eventlog.InstallAsEventCreate(
		constants.ServiceName,
		eventlog.Info|eventlog.Warning|eventlog.Error,
	)
	if err != nil {
		// Already registered? (common case)
		if errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
			return nil
		}
		return fmt.Errorf("failed to register event source %q: %w", constants.ServiceName, err)
	}
	return nil
}

// UnregisterEventSource removes an event source from the Windows Event Log.
// Call this during uninstall if needed.
func UnregisterEventSource() error {
	if err := eventlog.Remove(constants.ServiceName); err != nil {
		return fmt.Errorf("failed to remove event source %q: %w", constants.ServiceName, err)
	}
	return nil
}
