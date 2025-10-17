//go:build windows

package vhd

import (
	"os"
	"syscall"
	"time"
)

func (s *VHDTestSuite) TestGetStorePath() {

	runner := s.runner

	testStartTime := time.Now()
	time.Sleep(time.Millisecond)

	store, err := GetStorePath(runner)

	s.Require().NoError(err)
	s.Require().DirExists(store)

	// No error here, since DirExists above checks for existence
	st, _ := os.Stat(store)

	d := st.Sys().(*syscall.Win32FileAttributeData)
	creationTime := time.Unix(0, d.CreationTime.Nanoseconds())

	if creationTime.After(testStartTime) {
		// We created it, so trash it
		_ = os.RemoveAll(store)
	}
}
