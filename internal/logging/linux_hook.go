//go:build linux

package logging

import "github.com/sirupsen/logrus"

type nullHook struct{}

func (*nullHook) Levels() []logrus.Level {
	return []logrus.Level{}
}

func (*nullHook) Fire(*logrus.Entry) error {
	return nil
}

func getLogHook(bool) (logrus.Hook, error) {
	return &nullHook{}, nil
}
