//go:build windows

package vhd

import (
	"strings"

	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

func (s *VHDTestSuite) TestAttachDetach() {

	disk, err := GetByName(s.runner, s.pvStore, "pv01")

	s.Require().NoError(err)
	s.Require().NotNil(disk)

	// Attach
	// Do this twice for idempotency test
	for range 2 {
		//nolint:govet // intentional redeclaration of err
		attached, err := Attach(s.runner, s.pvStore, disk.DiskIdentifier, s.vm.ID)
		s.Require().NoError(err)

		disks, err := List(s.runner, s.pvStore, 0, "")
		s.Require().NoError(err)

		s.Require().True(diskAttached(disks, s.vm.ID, attached.Path), "Could not find attachment")
	}

	// Detach
	// Do this twice for idempotency test
	for range 2 {
		err = Detach(s.runner, s.pvStore, disk.DiskIdentifier, s.vm.ID)
		s.Require().NoError(err)

		disks, err := List(s.runner, s.pvStore, 0, "")
		s.Require().NoError(err)

		s.Require().False(diskAttached(disks, s.vm.ID, disk.Path), "Disk was not detached")
	}
}

func (s *VHDTestSuite) TestAttachFailsIfDiskNotFound() {

	_, err := Attach(s.runner, s.pvStore, uuid.NewString(), s.vm.ID)
	s.Require().Error(err)
	runnerErr := &powershell.RunnerError{}
	s.Require().ErrorAs(err, &runnerErr)
	s.Require().Equal(codes.NotFound, runnerErr.Code)
}

func (s *VHDTestSuite) TestAttachFailsIfVMNotFound() {

	disk, err := GetByName(s.runner, s.pvStore, "pv01")

	s.Require().NoError(err)
	s.Require().NotNil(disk)

	_, err = Attach(s.runner, s.pvStore, disk.DiskIdentifier, uuid.NewString())
	s.Require().Error(err)
	runnerErr := &powershell.RunnerError{}
	s.Require().ErrorAs(err, &runnerErr)
	s.Require().Equal(codes.NotFound, runnerErr.Code)
}

func (s *VHDTestSuite) TestDetachFailsIfDiskNotFound() {

	err := Detach(s.runner, s.pvStore, uuid.NewString(), s.vm.ID)
	s.Require().Error(err)
	runnerErr := &powershell.RunnerError{}
	s.Require().ErrorAs(err, &runnerErr)
	s.Require().Equal(codes.NotFound, runnerErr.Code)
}

func (s *VHDTestSuite) TestDetachFailsIfVMNotFound() {

	disk, err := GetByName(s.runner, s.pvStore, "pv01")

	s.Require().NoError(err)
	s.Require().NotNil(disk)

	_, err = Attach(s.runner, s.pvStore, disk.DiskIdentifier, uuid.NewString())
	s.Require().Error(err)
	runnerErr := &powershell.RunnerError{}
	s.Require().ErrorAs(err, &runnerErr)
	s.Require().Equal(codes.NotFound, runnerErr.Code)
}

func diskAttached(disks *models.ListVHDResponse, vmid, path string) bool {
	for _, d := range disks.VHDs {
		if d.Host != nil && strings.EqualFold(*d.Host, vmid) && strings.EqualFold(path, d.Path) {
			return true
		}
	}
	return false
}
