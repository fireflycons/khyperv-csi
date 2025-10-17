//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
)

func (s *controllerServer) GetCapacity() (*rest.GetCapacityResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"method": "get_capacity",
	})

	log.Info(messages.CONTROLLER_GET_CAPACITY)

	free, err := vhd.GetCapacity(s.runner, s.PVStore)

	if err != nil {
		return nil, s.processError(err, log, messages.CONTROLLER_GET_CAPACITY_FAILED)
	}

	log.Info(messages.CONTROLLER_GOT_CAPACITY)

	return &rest.GetCapacityResponse{
		AvailableCapacity: free,
		MinimumVolumeSize: constants.MinimumVolumeSizeInBytes,
	}, nil
}
