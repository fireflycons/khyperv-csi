//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

func (s *controllerServer) DeleteVolume(volId string) error {
	log := s.log.WithFields(logrus.Fields{
		"volume_id": volId,
		"method":    "delete_volume",
	})

	log.Info(messages.CONTROLLER_VOLUME_DELETE)

	if volId == "" {
		log.Error()
		return rest.NewError(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	err := vhd.Delete(s.runner, s.PVStore, volId)
	if err != nil {
		return s.processError(err, log, messages.CONTROLLER_VOLUME_DELETE_FAILED)
	}

	log.Info(messages.CONTROLLER_VOLUME_DELETED)
	return nil
}
