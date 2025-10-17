//go:build windows

package controller

import (
	"os"

	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func (s *ControllerTestSuite) TestDelete() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "", nil)

	err := s.server.DeleteVolume("pv1")
	s.Require().NoError(err)
	s.Require().True(s.logBuffer.ContainsMessage("volume was deleted"))
}

func (s *ControllerTestSuite) TestDeleteDuplicateVolume() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "INTERNAL : Duplicate disks found", os.ErrInvalid)

	err := s.server.DeleteVolume("pv1")

	targetError := &rest.Error{}
	s.Require().Error(err)
	s.Require().ErrorAs(err, &targetError)
	s.Require().Equal(targetError.Code, codes.Internal)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_DELETE_FAILED))
}

func (s *ControllerTestSuite) TestDeleteAttachedVolume() {

	s.shell.EXPECT().Execute(mock.Anything).Return("", "FAILED_PRECONDITION : Disk is attached", os.ErrInvalid)

	err := s.server.DeleteVolume("pv1")

	targetError := &rest.Error{}
	s.Require().Error(err)
	s.Require().ErrorAs(err, &targetError)
	s.Require().Equal(targetError.Code, codes.FailedPrecondition)
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VOLUME_DELETE_FAILED))
}
