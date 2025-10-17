//go:build windows

package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/fireflycons/hypervcsi/internal/external_mocks/mock_shell"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	runner    powershell.Runner
	shell     *mock_shell.MockShell
	server    *controllerServer
	logBuffer *logBuffer
}

var (
	_ suite.BeforeTest = (*ControllerTestSuite)(nil)
	_ suite.AfterTest  = (*ControllerTestSuite)(nil)
)

func TestControllerPackage(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (s *ControllerTestSuite) BeforeTest(suiteName, testName string) {

	s.logBuffer = &logBuffer{}
	s.shell = mock_shell.NewMockShell(s.T())

	// Close shell
	s.shell.EXPECT().Exit().Return(nil).Once()

	r, err := powershell.NewRunner(powershell.WithShell(s.shell))

	s.Require().NoError(err)
	s.runner = r

	s.server = &controllerServer{
		runner:  s.runner,
		PVStore: os.TempDir(),
		log: &logrus.Logger{
			Out:          s.logBuffer,
			Formatter:    new(logrus.TextFormatter),
			Hooks:        make(logrus.LevelHooks),
			Level:        logrus.InfoLevel,
			ExitFunc:     os.Exit,
			ReportCaller: false,
		},
	}

}

func (s *ControllerTestSuite) AfterTest(suiteName, testName string) {
	s.logBuffer = nil
	s.server.Close()
}

func (s *ControllerTestSuite) JSON(obj any) string {

	b, err := json.Marshal(obj)
	s.Require().NoError(err, "Cannot marshal object of type %T", obj)
	return string(b)
}

// logBuffer implements io.Writer and collects log lines.
type logBuffer struct {
	mu    sync.Mutex
	lines []string
	buf   bytes.Buffer
}

func (lb *logBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	n, err = lb.buf.Write(p)
	for {
		line, errLine := lb.buf.ReadString('\n')
		if errLine != nil {
			break
		}
		lb.lines = append(lb.lines, strings.TrimRight(line, "\r\n"))
	}
	return n, err
}

func (lb *logBuffer) Lines() []string {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	cpy := make([]string, len(lb.lines))
	copy(cpy, lb.lines)
	return cpy
}

// Extract retrieves the value for a given key from a log line.
// Handles quoted or unquoted values. Returns "" if the key is not found.
func (*logBuffer) Extract(line, key string) string {
	keyEq := key + "="
	inQuotes := false
	var cur strings.Builder
	var tokens []string

	for i := range len(line) {
		ch := line[i]

		switch ch {
		case ' ':
			if inQuotes {
				cur.WriteByte(ch)
			} else if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		case '"':
			inQuotes = !inQuotes
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}

	for _, token := range tokens {
		if strings.HasPrefix(token, keyEq) {
			return strings.TrimPrefix(token, keyEq)
		}
	}
	return ""
}

// ContainsMessage checks whether any log line's msg field contains substr.
func (lb *logBuffer) ContainsMessage(substr string) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, line := range lb.lines {
		msg := lb.Extract(line, "msg")
		if strings.Contains(msg, substr) {
			return true
		}
	}
	return false
}

func (lb *logBuffer) Dump() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	fmt.Println(strings.Join(lb.lines, "\n"))
}
