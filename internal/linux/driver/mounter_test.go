//go:build linux

package driver

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

func (s *driverTestSuite) Test_mounter_IsAttached() {
	testSource := "test-source"
	testEvalSymlinkErr := errors.New("eval sym link err")
	testReadFileErr := errors.New("read file err")

	_ = testEvalSymlinkErr
	type readFileResult struct {
		content []byte
		err     error
	}

	type evalSymlinksResult struct {
		path string
		err  error
	}

	tests := []struct {
		name         string
		readFile     *readFileResult
		evalSymlinks *evalSymlinksResult
		errorMsg     string
	}{
		{
			name: "could not evaluate the symbolic link",
			readFile: &readFileResult{
				content: nil,
				err:     testReadFileErr,
			},
			evalSymlinks: &evalSymlinksResult{
				path: "",
				err:  testEvalSymlinkErr,
			},
			errorMsg: fmt.Sprintf("error evaluating the symbolic link %q: %s", testSource, testEvalSymlinkErr),
		},
		{
			name: "error reading the device state file",
			readFile: &readFileResult{
				content: nil,
				err:     testReadFileErr,
			},
			evalSymlinks: &evalSymlinksResult{
				path: "/dev/sda",
				err:  nil,
			},
			errorMsg: fmt.Sprintf("error reading the device state file \"/sys/class/block/sda/device/state\": %s", testReadFileErr),
		},
		{
			name:     "error device name is empty",
			readFile: nil,
			evalSymlinks: &evalSymlinksResult{
				path: "/",
				err:  nil,
			},
			errorMsg: "error device name is empty for path /",
		},
		{
			name: "state file content does not indicate a running state",
			readFile: &readFileResult{
				content: []byte("not-running"),
				err:     nil,
			},
			evalSymlinks: &evalSymlinksResult{
				path: "/dev/sda",
				err:  nil,
			},
			errorMsg: fmt.Sprintf("error comparing the state file content, expected: %s, got: %s", runningState, "not-running"),
		},
		{
			name: "state file content indicates a running state",
			readFile: &readFileResult{
				content: []byte(runningState),
				err:     nil,
			},
			evalSymlinks: &evalSymlinksResult{
				path: "/dev/sda",
				err:  nil,
			},
			errorMsg: "",
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {

			ttAv := NewMockAttachmentValidator(s.T())

			if test.evalSymlinks != nil {
				ttAv.EXPECT().evalSymlinks(testSource).
					Return(test.evalSymlinks.path, test.evalSymlinks.err)
			}

			if test.readFile != nil && test.evalSymlinks.err == nil {
				ttAv.EXPECT().readFile(mock.Anything).
					Return(test.readFile.content, test.readFile.err)
			}

			m := &mounter{
				log:                 logrus.NewEntry(logrus.New()),
				attachmentValidator: ttAv,
			}

			err := m.IsAttached(testSource)

			if test.errorMsg != "" {
				s.Require().ErrorContains(err, test.errorMsg)
			} else {
				s.Require().NoError(err, "should not received an error")
			}
		})
	}
}
