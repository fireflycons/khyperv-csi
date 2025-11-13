//go:build windows

package controller

import (
	"fmt"
	"os"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func (s *ControllerTestSuite) TestExpandVolume() {

	newSize := 10 * constants.MiB

	disk := models.GetVHDResponse{
		Name:           "pv1",
		DiskIdentifier: constants.ZeroUUID,
		Size:           5 * constants.MiB,
		Path:           fmt.Sprintf("C:\\Temp\\pv1;%s.vhdx", constants.ZeroUUID),
	}

	resized := models.GetVHDResponse{
		Name:           "pv1",
		DiskIdentifier: constants.ZeroUUID,
		Size:           int64(newSize),
		Path:           fmt.Sprintf("C:\\Temp\\pv1;%s.vhdx", constants.ZeroUUID),
	}

	expected := &rest.ExpandVolumeResponse{
		CapacityBytes:         int64(newSize),
		NodeExpansionRequired: true,
	}

	// Disk will be looked up
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "", nil).Once()

	// and resized
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(resized), "", nil).Once()

	actual, err := s.server.ExpandVolume(disk.DiskIdentifier, int64(newSize))

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_EXPANDED))
}

func (s *ControllerTestSuite) TestExpandVolumeIdempotentSameSize() {

	disk := models.GetVHDResponse{
		Name:           "pv1",
		DiskIdentifier: constants.ZeroUUID,
		Size:           5 * constants.MiB,
		Path:           fmt.Sprintf("C:\\Temp\\pv1;%s.vhdx", constants.ZeroUUID),
	}

	newSize := disk.Size

	resized := models.GetVHDResponse{
		Name:           "pv1",
		DiskIdentifier: constants.ZeroUUID,
		Size:           newSize,
		Path:           fmt.Sprintf("C:\\Temp\\pv1;%s.vhdx", constants.ZeroUUID),
	}

	expected := &rest.ExpandVolumeResponse{
		CapacityBytes:         newSize,
		NodeExpansionRequired: false,
	}

	// Disk will be looked up
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "", nil).Once()

	// and resized
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(resized), "", nil).Once()

	actual, err := s.server.ExpandVolume(disk.DiskIdentifier, newSize)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_EXPANDED))
}

func (s *ControllerTestSuite) TestExpandVolumeIdempotentSmaller() {

	originalSize := int64(5 * constants.MiB)

	disk := models.GetVHDResponse{
		Name:           "pv1",
		DiskIdentifier: constants.ZeroUUID,
		Size:           originalSize,
		Path:           fmt.Sprintf("C:\\Temp\\pv1;%s.vhdx", constants.ZeroUUID),
	}

	newSize := disk.Size - constants.MiB

	expected := &rest.ExpandVolumeResponse{
		CapacityBytes:         originalSize,
		NodeExpansionRequired: false,
	}

	// Disk will be looked up
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "", nil).Once()

	// and resized
	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "", nil).Once()

	actual, err := s.server.ExpandVolume(disk.DiskIdentifier, newSize)

	s.Require().NoError(err)
	s.Require().Equal(expected, actual)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_EXPANDED))
}

func (s *ControllerTestSuite) TestExpandVolumeNotFound() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "NOT_FOUND : ", os.ErrNotExist).Once()

	_, err := s.server.ExpandVolume(constants.ZeroUUID, constants.GiB)
	targetErr := &rest.Error{}
	s.Require().ErrorAs(err, &targetErr)
	s.Require().Equal(targetErr.Code, codes.NotFound)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_EXPAND_VOLUME_FAILED), "%q not found in log", messages.CONTROLLER_EXPAND_VOLUME_FAILED)
}

func (s *ControllerTestSuite) TestExpandVolumeStorageFull() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "OUT_OF_RANGE : ", os.ErrNotExist).Once()

	_, err := s.server.ExpandVolume(constants.ZeroUUID, constants.TiB)
	targetErr := &rest.Error{}
	s.Require().ErrorAs(err, &targetErr)
	s.Require().Equal(targetErr.Code, codes.OutOfRange)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_EXPAND_VOLUME_FAILED), "%q not found in log", messages.CONTROLLER_EXPAND_VOLUME_FAILED)
}
