//go:build windows

package powershell

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/stretchr/testify/suite"
)

type PowershellTestSuite struct {
	suite.Suite
	runner Runner
}

var (
	_ suite.SetupAllSuite    = (*PowershellTestSuite)(nil)
	_ suite.TearDownAllSuite = (*PowershellTestSuite)(nil)
)

func TestPowershellPackage(t *testing.T) {
	suite.Run(t, new(PowershellTestSuite))
}

func (s *PowershellTestSuite) SetupSuite() {
	runner, err := NewRunner(WithModules(constants.PowerShellModule))

	if err != nil {
		if strings.Contains(err.Error(), "no valid module file was found") {
			err = fmt.Errorf("khyperv-csi module must be installed into Windows PowerShell before running these tests: %w", err)
		}
		s.FailNow(err.Error())
	}
	s.NoError(err)
	s.runner = runner
}

func (s *PowershellTestSuite) TearDownSuite() {
	s.runner.Exit()
}
