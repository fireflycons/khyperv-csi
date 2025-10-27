//go:build windows

package vhd

func (s *VHDTestSuite) TestGetVMs() {

	vms, err := GetVMs(s.runner)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(vms.VMs), 1)
}

func (s *VHDTestSuite) TestGetVm() {

	vm, err := GetVM(s.runner, s.vm.ID)
	s.Require().NoError(err)
	s.Require().Equal(s.vm, vm)
}
