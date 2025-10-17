//go:build windows

package vhd

import (
	"testing"

	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"google.golang.org/grpc/codes"
)

func (s *VHDTestSuite) TestGet() {

	allDisks, err := List(s.runner, s.pvStore, 0, "")
	s.Require().NoError(err)
	s.Require().NotEmpty(allDisks.VHDs)

	s.T().Run("by ID", func(*testing.T) {
		disk, err := GetByID(s.runner, s.pvStore, allDisks.VHDs[0].DiskIdentifier)
		s.Require().NoError(err)
		s.Require().NotNil(disk)
		s.Require().Equal(allDisks.VHDs[0].DiskIdentifier, disk.DiskIdentifier)
	})

	s.T().Run("by Name", func(*testing.T) {
		disk, err := GetByName(s.runner, s.pvStore, allDisks.VHDs[0].Name)
		s.Require().NoError(err)
		s.Require().NotNil(disk)
		s.Require().Equal(allDisks.VHDs[0].DiskIdentifier, disk.DiskIdentifier)
	})

	s.T().Run("by Name not found", func(*testing.T) {
		_, err := GetByName(s.runner, s.pvStore, "non-existent-disk")
		runnerError := &powershell.RunnerError{}
		s.Require().ErrorAs(err, &runnerError)
		s.Require().Equal(codes.NotFound, runnerError.Code)
	})

	s.T().Run("by ID not found", func(*testing.T) {
		_, err := GetByID(s.runner, s.pvStore, "00000000-0000-0000-0000-000000000000")
		runnerError := &powershell.RunnerError{}
		s.Require().ErrorAs(err, &runnerError)
		s.Require().Equal(codes.NotFound, runnerError.Code)
	})
}
