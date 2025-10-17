package powershell

import psg "github.com/fireflycons/go-powershell"

// LogAdapter implements the psg.Logging interface providing
// a wrapper for it that enables you to turn logging on and off
// selectively. Goo for debugging.
type LogAdapter struct {
	Logger  psg.Logger
	enabled bool
}

func (l *LogAdapter) Infof(format string, v ...any) {

	if !l.enabled {
		return
	}

	l.Logger.Infof(format, v...)
}

func (l *LogAdapter) Errorf(format string, v ...any) {

	if !l.enabled {
		return
	}

	l.Logger.Errorf(format, v...)
}

// On enables logging
func (l *LogAdapter) On() {
	l.enabled = true
}

// Off disables logging
func (l *LogAdapter) Off() {
	l.enabled = true
}
