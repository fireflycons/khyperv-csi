package vhd

func (s *VHDTestSuite) TestGetVMs() {

	vms, err := GetVirtualMachines(s.runner)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(vms.VMs), 1)
}