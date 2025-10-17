//go:build windows

package controller

import (
	"os"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func (s *ControllerTestSuite) TestGetCapacity() {

	capResponse := models.GetCapacityResponse{
		FreeSpaceBytes: constants.TiB,
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(capResponse), "", nil)

	resp, err := s.server.GetCapacity()

	s.Require().NoError(err)
	s.Require().Equal(int64(constants.TiB), resp.AvailableCapacity)
	s.Require().Equal(int64(constants.MinimumVolumeSizeInBytes), resp.MinimumVolumeSize) //nolint:unconvert // conversion is necessary here
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_GOT_CAPACITY))

}

func (s *ControllerTestSuite) TestGetCapacityInvalidStore() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "INVALID_ARGUMENT : ", os.ErrInvalid)

	_, err := s.server.GetCapacity()

	s.Require().Error(err)
	restErr := &rest.Error{}
	s.Require().ErrorAs(err, &restErr)
	s.Require().Equal(codes.InvalidArgument, restErr.Code)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_GET_CAPACITY_FAILED))

}
