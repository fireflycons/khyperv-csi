//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"google.golang.org/grpc/codes"
)

func (s *VHDTestSuite) TestGet() {

	allDisks, err := List(s.runner, s.pvStore, 0, "")
	s.Require().NoError(err)
	s.Require().NotEmpty(allDisks.VHDs)

	testDisk := allDisks.VHDs[0]

	s.Run("by ID", func() {
		disk, err := GetByID(s.runner, s.pvStore, testDisk.DiskIdentifier)
		s.Require().NoError(err)
		s.Require().NotNil(disk)
		s.Require().Equal(testDisk.DiskIdentifier, disk.DiskIdentifier)
	})

	s.Run("by Name", func() {
		disk, err := GetByName(s.runner, s.pvStore, testDisk.Name)
		s.Require().NoError(err)
		s.Require().NotNil(disk)
		s.Require().Equal(testDisk.DiskIdentifier, disk.DiskIdentifier)
	})

	s.Run("by Name not found", func() {
		_, err := GetByName(s.runner, s.pvStore, "non-existent-disk")
		runnerError := &powershell.RunnerError{}
		s.Require().ErrorAs(err, &runnerError)
		s.Require().Equal(codes.NotFound, runnerError.Code)
	})

	s.Run("by ID not found", func() {
		_, err := GetByID(s.runner, s.pvStore, constants.ZeroUUID)
		runnerError := &powershell.RunnerError{}
		s.Require().ErrorAs(err, &runnerError)
		s.Require().Equal(codes.NotFound, runnerError.Code)
	})
}
