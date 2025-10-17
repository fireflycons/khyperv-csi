//go:build windows

package controller

import (
	"os"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func (s *ControllerTestSuite) TestUnpublishVolume() {

	var (
		volId  = uuid.NewString()
		nodeId = uuid.NewString()
		path   = "C:\\Temp\\test.vhdx"
	)

	getDiskResponse := &models.GetVHDResponse{
		Path:           path,
		DiskIdentifier: volId,
		Name:           "pv1",
		Size:           10 * constants.MiB,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(getDiskResponse), "", nil).Once()
	s.shell.EXPECT().Execute(mock.Anything).Return("", "", nil).Once()

	err := s.server.UnpublishVolume(volId, nodeId)

	s.Require().NoError(err)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_UNPUBLISHED))
}

func (s *ControllerTestSuite) TestUnpublishVolumeFailsIfDiskDoesNotExist() {

	var (
		volId  = uuid.NewString()
		nodeId = uuid.NewString()
	)

	s.shell.EXPECT().Execute(mock.Anything).Return("", "NOT_FOUND : ", os.ErrNotExist).Once()

	err := s.server.UnpublishVolume(volId, nodeId)

	s.Require().Error(err)
	restErr := &rest.Error{}
	s.Require().ErrorAs(err, &restErr)
	s.Require().Equal(codes.NotFound, restErr.Code)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_UNPUBLISH_VOLUME_FAILED))
}

func (s *ControllerTestSuite) TestUnpublishVolumeFailsIfVMDoesNotExist() {

	var (
		volId  = uuid.NewString()
		nodeId = uuid.NewString()
		path   = "C:\\Temp\\test.vhdx"
	)

	getDiskResponse := &models.GetVHDResponse{
		Path:           path,
		DiskIdentifier: volId,
		Name:           "pv1",
		Size:           10 * constants.MiB,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(getDiskResponse), "", nil).Once()
	s.shell.EXPECT().Execute(mock.Anything).Return("", "NOT_FOUND : VM does not exist", os.ErrNotExist).Once()

	err := s.server.UnpublishVolume(volId, nodeId)

	s.Require().Error(err)
	restErr := &rest.Error{}
	s.Require().ErrorAs(err, &restErr)
	s.Require().Equal(codes.NotFound, restErr.Code)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_UNPUBLISH_VOLUME_FAILED))
}
