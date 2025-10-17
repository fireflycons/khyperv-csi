package logging

import "github.com/sirupsen/logrus"

func New() *logrus.Logger {

	log := logrus.New()

	if hook, err := getLogHook(false); err == nil {
		log.AddHook(hook)
	}

	return log
}

func NewDebug() *logrus.Logger {

	log := logrus.New()

	if hook, err := getLogHook(true); err == nil {
		log.AddHook(hook)
	}

	return log
}
