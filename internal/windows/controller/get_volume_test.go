//go:build windows

package controller

import (
	"os"

	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/stretchr/testify/mock"
)

func (s *ControllerTestSuite) TestGetVolume() {

	disk := models.GetVHDResponse{
		Name:           "pv1",
		DiskIdentifier: "constants.ZeroUUID",
	}

	vols := &models.ListVHDResponse{
		VHDs: []models.GetVHDResponse{
			disk,
		},
	}

	s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(vols), "", nil).Once()
	allDisks, err := vhd.List(s.runner, s.server.PVStore, 0, "")
	s.Require().NoError(err)
	s.Require().NotEmpty(allDisks.VHDs)

	s.Run("by ID", func() {
		s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "", nil).Once()
		volResp, err := s.server.GetVolume(allDisks.VHDs[0].DiskIdentifier)
		s.Require().NoError(err)
		s.Require().NotNil(volResp)
		s.Require().Equal(allDisks.VHDs[0].DiskIdentifier, volResp.ID)
		s.Require().Equal(allDisks.VHDs[0].Name, volResp.Name)
	})

	s.Run("by Name", func() {
		s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "NOT_FOUND : dd", os.ErrNotExist).Once()
		s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "", nil).Once()
		volResp, err := s.server.GetVolume(allDisks.VHDs[0].Name)
		s.Require().NoError(err)
		s.Require().NotNil(volResp)
		s.Require().Equal(allDisks.VHDs[0].DiskIdentifier, volResp.ID)
		s.Require().Equal(allDisks.VHDs[0].Name, volResp.Name)
	})

	s.Run("not found", func() {
		s.shell.EXPECT().Execute(mock.Anything).Return(s.JSON(disk), "NOT_FOUND : dd", os.ErrNotExist).Times(2)
		_, err := s.server.GetVolume("non-existent-volume")
		s.Require().Error(err)
	})
}
