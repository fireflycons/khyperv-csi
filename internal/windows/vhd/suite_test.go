//go:build windows

package vhd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/fireflycons/hypervcsi/internal/constants"
	"github.com/fireflycons/hypervcsi/internal/models"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"github.com/fireflycons/hypervcsi/internal/windows/win32"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// VHDTestSuite tests low-level interaction with Hyper-V.
//
// Requires Hyper-V service to be running on the machine
// running the tests and the user account under which the
// tests are running to be a member of the Hyper-V Administrators
// group.
type VHDTestSuite struct {
	suite.Suite
	pvStore string
	vmStore string
	vm      *models.GetVMResponse
	runner  powershell.Runner
	logger  powershell.LogAdapter
}

const (
	HyperVHostComputeService  win32.ServiceName = "vmcompute"
	HyperVDataExchangeService win32.ServiceName = "vmickvpexchange"
)

var (
	_ suite.SetupAllSuite    = (*VHDTestSuite)(nil)
	_ suite.TearDownAllSuite = (*VHDTestSuite)(nil)
	_ suite.BeforeTest       = (*VHDTestSuite)(nil)
	_ suite.AfterTest        = (*VHDTestSuite)(nil)
)

// an adapter for standard library "log" package
type logAdapter struct {
	logger *log.Logger
}

func (l *logAdapter) Infof(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("INFO:  %s", msg)
}

func (l *logAdapter) Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("ERROR: %s", msg)
}

func TestVHDPackage(t *testing.T) {
	suite.Run(t, new(VHDTestSuite))
}

func (s *VHDTestSuite) SetupSuite() {

	// We need Hyper-V up for these tests
	hyperV, err := win32.IsServiceRunning(HyperVHostComputeService)
	s.Require().NoError(err)
	s.Require().True(hyperV, "Hyper-V Host Compute Service is not running")

	switchableLog := &powershell.LogAdapter{
		Logger: &logAdapter{
			logger: log.New(os.Stderr, "PS: ", 0),
		},
	}

	runner, err := powershell.NewRunner(
		powershell.WithModules(constants.PowerShellModule),
		powershell.WithLogger(switchableLog),
	)

	if err != nil {
		if strings.Contains(err.Error(), "no valid module file was found") {
			err = fmt.Errorf("khyperv-csi module must be installed into Windows PowerShell before running these tests: %w", err)
		}
		s.FailNow(err.Error())
	}

	s.runner = runner
	s.logger = *switchableLog

	st, err := win32.GetLongPathName(filepath.Join(os.TempDir(), "khypervcsi-test", "disks"))
	s.Require().NoError(err, "Cannot resolve PV store path")

	s.pvStore = st

	fmt.Printf("\n\nPVStore: %s\n\n", s.pvStore)

	if _, err := os.Stat(s.pvStore); os.IsNotExist(err) {
		err = os.MkdirAll(s.pvStore, 0755)
		s.NoError(err)
	}

	s.vmStore = filepath.Join(os.TempDir(), "khypervcsi-test", "vm")

	if _, err := os.Stat(s.vmStore); os.IsNotExist(err) {
		err = os.MkdirAll(s.vmStore, 0755)
		s.NoError(err)
	}

	s.createTestDisks()
	s.setupTestVM()
}

func (s *VHDTestSuite) TearDownSuite() {

	s.teardownTestVM()

	if _, err := os.Stat(s.pvStore); err == nil {
		// Clear down PV store
		_ = os.RemoveAll(s.pvStore)
	}

	if _, err := os.Stat(s.vmStore); err == nil {
		// Clear down PV store
		_ = os.RemoveAll(s.vmStore)
	}

	s.runner.Exit()
}

func (s *VHDTestSuite) BeforeTest(suiteName, testName string) {

	switch {
	case isTestFor(testName, s.TestDelete):
		s.MustNewDisk(
			"pvx",
			10*constants.MiB,
		)

	case isTestFor(testName, s.TestListWithAttachedVolume):

		disk, err := GetByName(s.runner, s.pvStore, "pv10")
		s.Require().NoError(err)

		_, err = Attach(s.runner, s.pvStore, disk.DiskIdentifier, s.vm.ID)
		s.Require().NoError(err)
		fmt.Printf("Attached disk %s to VM %s\n", disk.DiskIdentifier, s.vm.Name)
	}
}

func (s *VHDTestSuite) AfterTest(suiteName, testName string) {

	if isTestFor(testName, s.TestListWithAttachedVolume) {

		disk, err := GetByName(s.runner, s.pvStore, "pv10")
		s.Require().NoError(err)

		err = Detach(s.runner, s.pvStore, disk.DiskIdentifier, s.vm.ID)
		s.Require().NoError(err)
		fmt.Printf("Detached disk %s from VM %s\n", disk.DiskIdentifier, s.vm.Name)
	}
}

func isTestFor(name string, testFuncs ...any) bool {

	for _, f := range testFuncs {

		t := reflect.TypeOf(f)
		if t != nil && t.Kind() == reflect.Func {

			fn := getFunctionName(f)
			if getFunctionName(f) == name || strings.HasPrefix(fn, name+"-") {
				return true
			}
		}
	}
	return false
}

func getFunctionName(f any) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()), ".")
	return strs[len(strs)-1]
}

const numVhdsForListTests = 10

func (s *VHDTestSuite) createTestDisks() {

	for i := 1; i <= numVhdsForListTests; i++ {
		s.MustNewDisk(fmt.Sprintf("pv%02d", i), 10*constants.MiB)
	}
}

func (s *VHDTestSuite) setupTestVM() {

	vm := &models.GetVMResponse{}

	_, err := executeWithReturn(
		s.runner,
		vm,
		powershell.NewCmdlet(
			"New-PVTestVM",
			map[string]any{
				"Name":       uuid.New().String(),
				"Path":       s.vmStore,
				"Generation": 2,
			},
		),
	)

	s.Require().NoError(err, "cannot create test VM")
	s.vm = vm
	fmt.Printf("Created test VM %s\n", vm.Name)
}

func (s *VHDTestSuite) teardownTestVM() {

	err := s.runner.Run(
		powershell.NewCmdlet(
			"Get-VM",
			map[string]any{
				"Name": s.vm.Name,
			},
		),
		powershell.NewCmdlet(
			"Remove-VM",
			map[string]any{
				"Force": nil,
			},
		),
	)

	if err != nil {
		fmt.Printf("cannot remove test VM %s: %v\n", s.vm.Name, err)
	}
}

func (s *VHDTestSuite) MustNewDisk(name string, size int64) {
	_, err := New(
		s.runner,
		name,
		s.pvStore,
		size,
	)
	s.Require().NoError(err)
	fmt.Printf("Created disk %s\n", name)
}

func (s *VHDTestSuite) assertDiskExists(path string) {

	_, err := os.Stat(path)
	s.Require().NoError(err, "Expected to find %s", path)
}

func (s *VHDTestSuite) assertDiskNotExists(path string) {

	_, err := os.Stat(path)
	s.Require().Error(err, "Expected not to find %s", path)
}

func (s *VHDTestSuite) dumpJson(v any, out io.Writer) {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	s.Require().NoError(enc.Encode(v))
}
