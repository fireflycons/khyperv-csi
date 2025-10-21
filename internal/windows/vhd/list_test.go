//go:build windows

package vhd

import (
	"os"
	"slices"
	"strings"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/fireflycons/hypervcsi/internal/models"
)

func (s *VHDTestSuite) TestList() {

	disks, err := List(s.runner, s.pvStore, 5, "")

	s.Require().NoError(err)
	assertCompleteVolumeInfo(s, disks)
	s.Require().Len(disks.VHDs, 5)
	s.Require().NotEmpty(disks.NextToken)

	disks2, err := List(s.runner, s.pvStore, 5, disks.NextToken)

	assertCompleteVolumeInfo(s, disks2)
	s.Require().NoError(err)
	s.Require().Len(disks2.VHDs, 5)

	if !slices.ContainsFunc(disks.VHDs, func(v models.GetVHDResponse) bool {
		return v.Host != nil
	}) {
		// Only check if no attached disks are found, which would increase the total count
		s.Require().Empty(disks2.NextToken)
	}

	for d1 := range disks.VHDs {
		s.Require().NotContains(disks2.VHDs, disks.VHDs[d1])
	}
}

func (s *VHDTestSuite) TestListWithAttachedVolume() {

	disks, err := List(s.runner, s.pvStore, 0, "")

	s.Require().NoError(err)
	s.Require().NotEmpty(disks.VHDs)
	assertCompleteVolumeInfo(s, disks)

	found := From(disks.VHDs).
		Where(func(i any) bool {
			h := i.(models.GetVHDResponse).Host
			if h == nil {
				return false
			}
			return strings.EqualFold(*h, s.vm.ID)
		}).First()

	if found == nil {
		s.dumpJson(disks, os.Stdout)
	}
	s.Require().NotNil(found, "Could not find attached disk")
}

func assertCompleteVolumeInfo(suite *VHDTestSuite, l *models.ListVHDResponse) {

	for _, v := range l.VHDs {
		suite.Assert().NotEmpty(v.Path, "Path is empty")
		suite.Assert().NotEmpty(v.DiskIdentifier, "DiskIdentifier is empty")
		suite.Assert().NotZero(v.Size, "Size is zero")
	}
}
