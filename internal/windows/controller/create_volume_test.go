//go:build windows

package controller

import (
	"os"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func (s *ControllerTestSuite) TestCreate() {

	const (
		size   = 10 * constants.MiB
		diskId = "00000000-0000-0000-0000-000000000000"
	)

	newVhdResponse := &models.GetVHDResponse{
		Path:           "C:\\Temp\\pv1;00000000-0000-0000-0000-000000000000.vhdx",
		Name:           "pv1",
		Size:           size,
		DiskIdentifier: diskId,
	}

	expected := &rest.CreateVolumeResponse{
		ID:   diskId,
		Size: size,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return("", "NOT_FOUND : ", os.ErrNotExist).Once()
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(newVhdResponse), "", nil).Once()

	actual, err := s.server.CreateVolume("pv1", size)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_CREATED))
}

func (s *ControllerTestSuite) TestCreateUnderMinSize() {

	const (
		size   = 10 * constants.MiB
		diskId = "00000000-0000-0000-0000-000000000000"
	)

	newVhdResponse := &models.GetVHDResponse{
		Path:           "C:\\Temp\\pv1;00000000-0000-0000-0000-000000000000.vhdx",
		Name:           "pv1",
		Size:           constants.MinimumVolumeSizeInBytes,
		DiskIdentifier: diskId,
	}

	expected := &rest.CreateVolumeResponse{
		ID:   diskId,
		Size: constants.MinimumVolumeSizeInBytes,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return("", "NOT_FOUND : ", os.ErrNotExist).Once()
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(newVhdResponse), "", nil).Once()

	actual, err := s.server.CreateVolume("pv1", size)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_CREATED))
}

func (s *ControllerTestSuite) TestCreateIdempotent() {

	const (
		size   = 10 * constants.MiB
		diskId = "00000000-0000-0000-0000-000000000000"
	)

	exitingVhdResponse := &models.GetVHDResponse{
		Path:           "C:\\Temp\\pv1;00000000-0000-0000-0000-000000000000.vhdx",
		Name:           "pv1",
		Size:           size,
		DiskIdentifier: diskId,
	}

	expected := &rest.CreateVolumeResponse{
		ID:   diskId,
		Size: size,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(exitingVhdResponse), "", nil).Once()

	actual, err := s.server.CreateVolume("pv1", size)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_ALREADY_CREATED))
}

func (s *ControllerTestSuite) TestCreateResourceExhausted() {

	const (
		size   = 10 * constants.MiB
		diskId = "00000000-0000-0000-0000-000000000000"
	)

	s.shell.EXPECT().Execute(mock.Anything).Return("", "NOT_FOUND : ", os.ErrNotExist).Once()
	s.shell.EXPECT().Execute(mock.Anything).Return("", "RESOURCE_EXHAUSTED : Insufficient storage", vhd.ErrCapacityExhausted).Once()

	actual, err := s.server.CreateVolume("pv1", size)
	s.Require().Nil(actual)

	targetErr := &rest.Error{}
	s.Require().ErrorAs(err, &targetErr)
	s.Require().Equal(targetErr.Code, codes.ResourceExhausted)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_STORAGE_FULL))
}

func (s *ControllerTestSuite) TestCreateWithDifferentSizeWhenDiskExistsIsError() {

	const (
		size   = 10 * constants.MiB
		diskId = "00000000-0000-0000-0000-000000000000"
	)

	exitingVhdResponse := &models.GetVHDResponse{
		Path:           "C:\\Temp\\pv1;00000000-0000-0000-0000-000000000000.vhdx",
		Name:           "pv1",
		Size:           size * 2,
		DiskIdentifier: diskId,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(exitingVhdResponse), "", nil).Once()

	actual, err := s.server.CreateVolume("pv1", size)
	s.Require().Nil(actual)

	targetErr := &rest.Error{}
	s.Require().ErrorAs(err, &targetErr)
	s.Require().Equal(targetErr.Code, codes.AlreadyExists)
}
