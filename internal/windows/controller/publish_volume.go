//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
)

func (s *controllerServer) PublishVolume(volumeId, nodeId string) error {

	log := s.log.WithFields(logrus.Fields{
		"volume_id": volumeId,
		"node_id":   nodeId,
		"method":    "publish_volume",
	})

	log.Info(messages.CONTROLLER_PUBLISH_VOLUME)

	_, err := vhd.Attach(s.runner, s.PVStore, volumeId, nodeId)

	if err != nil {
		return s.processError(err, log, messages.CONTROLLER_PUBLISH_VOLUME_FAILED)
	}

	log.Info(messages.CONTROLLER_VOLUME_PUBLISHED)
	return nil
}
