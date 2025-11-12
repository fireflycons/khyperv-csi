package logging

import "github.com/sirupsen/logrus"

func New(level logrus.Level) *logrus.Logger {

	log := logrus.New()
	log.Level = level

	if hook, err := getLogHook(false); err == nil {
		log.AddHook(hook)
	}

	return log
}

func NewDebug() *logrus.Logger {

	log := logrus.New()
	log.Level = logrus.DebugLevel

	if hook, err := getLogHook(true); err == nil {
		log.AddHook(hook)
	}

	return log
}
