//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

func (s *controllerServer) GetVolume(name string) (*rest.GetVolumeResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"volume_name": name,
		"method":      "get_volume",
	})

	log.Info(messages.CONTROLLER_GET_VOLUME)

	vol, err := vhd.GetByID(s.runner, s.PVStore, name)

	if err != nil {
		restErr := s.processError(err, log, messages.CONTROLLER_GET_VOLUME_FAILED, codes.NotFound)

		if restErr.Code != codes.NotFound {
			return nil, restErr
		}

		// Now try by name
		vol, err = vhd.GetByName(s.runner, s.PVStore, name)
		if err != nil {
			restErr := s.processError(err, log, messages.CONTROLLER_GET_VOLUME_FAILED)
			return nil, restErr
		}
	}

	resp := &rest.GetVolumeResponse{
		Name: vol.Name,
		ID:   vol.DiskIdentifier,
		Size: vol.Size,
	}

	log.WithField("response", resp).Info(messages.CONTROLLER_GET_VOLUME_OK)

	return resp, nil
}
