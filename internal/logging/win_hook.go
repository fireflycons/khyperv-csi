//go:build windows

package logging

import (
	"github.com/fireflycons/hypervcsi/internal/logging/wineventlog"
	"github.com/sirupsen/logrus"
)

func getLogHook(debug bool) (logrus.Hook, error) {
	return wineventlog.New(debug)
}
