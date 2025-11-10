//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
)

func (s *controllerServer) ExpandVolume(volumeId string, size int64) (*rest.ExpandVolumeResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"volume_id": volumeId,
		"new_size":  common.FormatBytes(size),
		"method":    "expand_volume",
	})
	log.Info(messages.CONTROLLER_EXPAND_VOLUME)

	origVol, err := vhd.GetByID(s.runner, s.PVStore, volumeId)

	if err != nil {
		restErr := s.processError(err, log, messages.CONTROLLER_EXPAND_VOLUME_FAILED)
		return nil, restErr
	}

	vol, err := vhd.Resize(s.runner, s.PVStore, volumeId, size)

	if err != nil {
		return nil, s.processError(err, log, messages.CONTROLLER_EXPAND_VOLUME_FAILED)
	}

	log.Info(messages.CONTROLLER_VOLUME_EXPANDED)

	return &rest.ExpandVolumeResponse{
		CapacityBytes:         vol.Size,
		NodeExpansionRequired: vol.Size > origVol.Size,
	}, nil

}
