//go:build windows

package powershell

import (
	"errors"
	"fmt"
	"strings"

	psg "github.com/fireflycons/go-powershell"
	"github.com/fireflycons/go-powershell/backend"
)

// Runner is the interface through which to execute PowerShell
// commands on the Hyper-V host.
type Runner interface {
	// Run executes the given cmdlet
	Run(cmdlets ...Cmdlet) error

	// RunWitResult executes the given cmdlet and collects a JSON response
	RunWithResult(cmdlets ...Cmdlet) (string, error)

	// Exit releases any resources associated with the runner
	Exit()
}

type concreteRunner struct {
	shell psg.Shell
}

var _ Runner = (*concreteRunner)(nil)

type runnerOptions struct {
	shell         psg.Shell
	logger        psg.Logger
	importModules []string
}

type RunnerOptionFunc func(*runnerOptions)

// NewRunner creates a concrete implementation of the Runner interface.
func NewRunner(opts ...RunnerOptionFunc) (*concreteRunner, error) {

	var s psg.Shell

	ro := runnerOptions{}
	for _, opt := range opts {
		opt(&ro)
	}

	if ro.shell != nil {
		s = ro.shell
	} else {
		var err error

		shopts := make([]psg.ShellOptionFunc, 0, 2)

		if ro.logger != nil {
			shopts = append(shopts, psg.WithLogger(ro.logger))
		}

		if len(ro.importModules) > 0 {
			shopts = append(shopts, psg.WithModules(ro.importModules...))
		}

		s, err = psg.New(&backend.Local{}, shopts...)

		if err != nil {
			return nil, fmt.Errorf("failed to start shell: %w", err)
		}
	}

	return &concreteRunner{
		shell: s,
	}, nil
}

func WithShell(s psg.Shell) RunnerOptionFunc {
	return func(ro *runnerOptions) {
		ro.shell = s
	}
}

func WithModules(modules ...string) RunnerOptionFunc {
	return func(ro *runnerOptions) {
		ro.importModules = modules
	}
}

func WithLogger(logger psg.Logger) RunnerOptionFunc {
	return func(ro *runnerOptions) {
		ro.logger = logger
	}
}

// Run executes the given cmdlet(s).
func (r *concreteRunner) Run(cmdlets ...Cmdlet) error {

	_, err := r.RunWithResult(cmdlets...)

	return err
}

// RunWitResult executes the given cmdlet and collects a JSON response
func (r *concreteRunner) RunWithResult(cmdlets ...Cmdlet) (string, error) {

	cmd, err := buildCommand(cmdlets)

	if err != nil {
		return "", err
	}

	stdout, stderr, err := r.shell.Execute(cmd)

	if err != nil {
		code := extractCsiErrorCode(stderr)
		err = &RunnerError{
			ProcessError: err,
			Stderr:       stderr,
			Code:         code,
		}
	}

	return strings.TrimSpace(stdout), err
}

// Exit releases any resources associated with the runner
func (r *concreteRunner) Exit() {
	_ = r.shell.Exit()
}

func buildCommand(cmdlets []Cmdlet) (string, error) {

	if len(cmdlets) == 0 {
		return "", errors.New("missing cmdlet")
	}

	commandString := make([]string, 0, len(cmdlets)+1)

	for _, c := range cmdlets {
		if c.Err != nil {
			return "", fmt.Errorf("%s: %w", c.Name, c.Err)
		}

		commandString = append(commandString, c.FullCommand)
	}

	return strings.Join(commandString, " | "), nil
}
