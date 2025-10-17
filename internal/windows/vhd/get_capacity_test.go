//go:build windows

package vhd

func (s *VHDTestSuite) TestGetCapacity() {

	capacity, err := GetCapacity(s.runner, s.pvStore)

	s.Require().NoError(err)
	s.Require().Greater(capacity, int64(0))
}

func (s *VHDTestSuite) TestGetCapacityInvalidStore() {

	capacity, err := GetCapacity(s.runner, "1:\\invalid")
	s.Require().Error(err)
	_ = capacity
}
