//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
)

func (s *controllerServer) ListVolumes(maxEntries int32, nextToken string) (*models.ListVHDResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"max_entries":        maxEntries,
		"req_starting_token": nextToken,
		"method":             "list_volumes",
	})

	log.Info(messages.CONTROLLER_LIST_VOLUMES)

	disks, err := vhd.List(s.runner, s.PVStore, maxEntries, nextToken)

	if err != nil {
		return nil, s.processError(err, log, messages.CONTROLLER_LIST_VOLUMES_FAILED)
	}

	log.Info(messages.CONTROLLER_VOLUMES_LISTED)

	return disks, nil
}
