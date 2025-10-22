//go:build windows

// NOTES:
//
// Volume name will be the filename of the VHD file created in the PV storage directory

package controller

import (
	"errors"
	"fmt"

	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/shared"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

func (s *controllerServer) CreateVolume(name string, size int64) (*rest.CreateVolumeResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"volume_name":  name,
		"storage_size": shared.FormatBytes(size),
		"method":       "create_volume",
	})
	log.Info(messages.CONTROLLER_CREATE_VOLUME)

	vol, err := vhd.GetByName(s.runner, s.PVStore, name)

	if err != nil {
		restErr := s.processError(err, log, messages.CONTROLLER_CREATE_VOLUME_FAILED, codes.NotFound)

		// We generally expect the disk to not be found
		if restErr.Code != codes.NotFound {
			return nil, restErr
		}
	}

	if vol != nil {

		if vol.Size != size {
			log.Error(messages.CONTROLLER_VOLUME_EXISTS)
			return nil, rest.NewError(codes.AlreadyExists, fmt.Sprintf("invalid option requested size: %d", size))
		}

		log.Info(messages.CONTROLLER_VOLUME_ALREADY_CREATED)

		return &rest.CreateVolumeResponse{
			ID:   vol.DiskIdentifier,
			Size: vol.Size,
		}, nil
	}

	vol, err = vhd.New(
		s.runner,
		name,
		s.PVStore,
		size,
	)

	if err != nil {
		log.Error(err.Error())
		restErr := &rest.Error{}

		if errors.As(err, &restErr) && restErr.Code == codes.ResourceExhausted {
			log.Error(messages.CONTROLLER_STORAGE_FULL)
			return nil, restErr
		}

		return nil, rest.NewError(codes.Internal, err.Error())
	}

	resp := &rest.CreateVolumeResponse{
		ID:   vol.DiskIdentifier,
		Size: vol.Size,
	}

	log.WithField("response", resp).Info(messages.CONTROLLER_VOLUME_CREATED)

	return resp, nil
}
