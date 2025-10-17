//go:build windows

package wineventlog

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	_SOURCE = "khypervcsi"
)

type WinEventLogHook struct {
	log          debug.Log
	isRegistered bool
}

func New(isDebug bool) (*WinEventLogHook, error) {

	wel := &WinEventLogHook{}

	var el debug.Log

	if isDebug {
		el = debug.New(_SOURCE)
	} else {

		err := eventlog.InstallAsEventCreate(
			_SOURCE,
			eventlog.Info|eventlog.Warning|eventlog.Error,
		)
		if err != nil {
			// Already registered? (common case)
			if !errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
				return wel, fmt.Errorf("failed to register event source %q: %w", _SOURCE, err)
			}
		}

		el, err = eventlog.Open(_SOURCE)
		if err != nil {
			return wel, err
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
		_SOURCE,
		eventlog.Info|eventlog.Warning|eventlog.Error,
	)
	if err != nil {
		// Already registered? (common case)
		if errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
			return nil
		}
		return fmt.Errorf("failed to register event source %q: %w", _SOURCE, err)
	}
	return nil
}

// UnregisterEventSource removes an event source from the Windows Event Log.
// Call this during uninstall if needed.
func UnregisterEventSource() error {
	if err := eventlog.Remove(_SOURCE); err != nil {
		return fmt.Errorf("failed to remove event source %q: %w", _SOURCE, err)
	}
	return nil
}
