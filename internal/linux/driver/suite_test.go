//go:build linux

package driver

import (
	"testing"

	"github.com/fireflycons/hypervcsi/internal/common"
	"github.com/stretchr/testify/suite"
)

type driverTestSuite struct {
	common.SuiteBase
}

func TestDriverPackage(t *testing.T) {
	suite.Run(t, new(driverTestSuite))
}
