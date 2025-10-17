//go:build windows

package vhd

import "path/filepath"

func (s *VHDTestSuite) TestDelete() {

	toDelete, _ := filepath.Glob(filepath.Join(s.pvStore, "pvx*.vhdx"))

	s.Require().NotEmpty(toDelete, "No disks found to delete")

	name, id, err := ParseDiskPath(toDelete[0])
	s.Require().NoError(err)
	s.Require().NotEmpty(name)
	s.Require().NotEmpty(id)

	for _, path := range toDelete {

		//nolint:govet // intentional redeclaration of err
		err := Delete(s.runner, s.pvStore, id)
		s.Require().NoError(err)
		s.assertDiskNotExists(path)
	}

	// Deleting a non-existing disk should not return an error
	err = Delete(s.runner, s.pvStore, id)
	s.Require().NoError(err)
}
