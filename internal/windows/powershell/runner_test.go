//go:build windows

package powershell

import (
	"errors"
	"fmt"
)

func (s *PowershellTestSuite) TestConcreteRunner() {

	runner, err := NewRunner()
	s.Require().NoError(err)
	defer runner.Exit()

	cmdlet := NewCmdlet("Write-Host", map[string]any{
		"Object": `{"foo": "bar"}`,
	})
	outStr, err := runner.RunWithResult(cmdlet)

	if err != nil && errors.Is(err, &RunnerError{}) {
		fmt.Println(err.(*RunnerError).Stderr)
	}
	s.Require().NoError(err)
	s.Require().Equal(`{"foo": "bar"}`, outStr)
}
