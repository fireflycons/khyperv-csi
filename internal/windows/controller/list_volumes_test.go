//go:build windows

package controller

import (
	"os"

	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func (s *ControllerTestSuite) TestListVolumes() {

	vols := &models.ListVHDResponse{
		VHDs: make([]models.GetVHDResponse, 10),
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(vols), "", nil).Once()

	disks, err := s.server.ListVolumes(0, "")

	s.Require().NoError(err)
	s.Require().Len(disks.VHDs, len(vols.VHDs))
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUMES_LISTED))
}

func (s *ControllerTestSuite) TestListVolumesInvalidPath() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "INVALID_ARGUMENT :", os.ErrInvalid).Once()

	_, err := s.server.ListVolumes(0, "")

	s.Require().Error(err)

	restErr := &rest.Error{}
	s.Require().ErrorAs(err, &restErr)
	s.Require().Equal(codes.InvalidArgument, restErr.Code)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_LIST_VOLUMES_FAILED))
}
