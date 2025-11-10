//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"google.golang.org/grpc/codes"
)

func (s *VHDTestSuite) TestResize() {

	allDisks, err := List(s.runner, s.pvStore, 0, "")
	s.Require().NoError(err)
	s.Require().NotEmpty(allDisks.VHDs)

	resizeDisk := allDisks.VHDs[len(allDisks.VHDs)-1]
	newSize := resizeDisk.Size + (5 * constants.MiB)

	idempotentDisk := allDisks.VHDs[0]

	// Don't want to run separate tests on same volume
	s.Assert().NotEqual(resizeDisk, idempotentDisk)

	s.Run("Disk is expanded", func() {
		disk, err := Resize(s.runner, s.pvStore, resizeDisk.DiskIdentifier, newSize)
		s.Require().NoError(err)
		s.Require().NotNil(disk)
		s.Require().Equal(disk.Size, newSize)
	})

	s.Run("Expansion is idempotent", func() {
		disk, err := Resize(s.runner, s.pvStore, idempotentDisk.DiskIdentifier, idempotentDisk.Size-constants.MiB)
		s.Require().NoError(err)
		s.Require().NotNil(disk)
		s.Require().Equal(disk.Size, idempotentDisk.Size)
	})

	s.Run("Expansion fails if insufficient space", func() {
		_, err := Resize(s.runner, s.pvStore, resizeDisk.DiskIdentifier, 50*constants.TiB)
		runnerError := &powershell.RunnerError{}
		s.Require().ErrorAs(err, &runnerError)
		s.Require().Equal(codes.OutOfRange, runnerError.Code)
	})

	s.Run("Disk not found", func() {
		_, err := Resize(s.runner, s.pvStore, constants.ZeroUUID, constants.GiB)
		runnerError := &powershell.RunnerError{}
		s.Require().ErrorAs(err, &runnerError)
		s.Require().Equal(codes.NotFound, runnerError.Code)
	})
}
