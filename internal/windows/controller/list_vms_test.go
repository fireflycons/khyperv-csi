package controller

import (
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/stretchr/testify/mock"
)

func (s *ControllerTestSuite) TestListVms() {

	vms := &rest.ListVMResponse{
		VMs: []*models.GetVMResponse{
			{
				Name:       "pv1",
				ID:         "00000000-0000-0000-0000-0000000000000",
				Path:       "C:\\temp\\pv1;00000000-0000-0000-0000-0000000000000.vhdx",
				Generation: 2,
			},
		},
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(vms), "", nil).Once()

	actual, err := s.server.ListVms()
	s.Require().NoError(err)
	s.Require().Len(actual.VMs, len(vms.VMs))
	s.Require().True(s.logBuffer.ContainsMessage(messages.CONTROLLER_VMS_LISTED))
}
