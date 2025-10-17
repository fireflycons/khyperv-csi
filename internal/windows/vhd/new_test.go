//go:build windows

package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/constants"
)

func (s *VHDTestSuite) TestNew() {

	disk, err := New(
		s.runner,
		"pv1",
		s.pvStore,
		10*constants.MiB,
	)

	s.Require().NoError(err)
	s.Require().NotNil(disk)
	s.assertDiskExists(disk.Path)
}
