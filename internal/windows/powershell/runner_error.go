//go:build windows

package powershell

import (
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"google.golang.org/grpc/codes"
)

// RunnerError represents an error raised from code in the PowerShell module
type RunnerError struct {
	ProcessError error
	Stderr       string
	Code         codes.Code
}

func (e *RunnerError) Error() string {
	return e.ProcessError.Error() + ": " + e.Stderr
}

func (*RunnerError) Is(err error) bool {
	_, ok := err.(*RunnerError)
	return ok
}

func (e *RunnerError) As(target any) bool {
	if t, ok := target.(**rest.Error); ok {
		*t = &rest.Error{
			Code:    e.Code,
			Message: e.Error(),
		}
		return true
	}
	return false
}
