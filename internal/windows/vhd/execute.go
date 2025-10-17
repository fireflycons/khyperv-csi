//go:build windows

package vhd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

// execute runs the given cmdlets where output is not either required or expected.
//
// cmdlets are chained with |
func execute(runner powershell.Runner, cmdlets ...powershell.Cmdlet) error {

	if err := runner.Run(cmdlets...); err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	return nil
}

// executeWithReturn runs the given cmdlet with arguments and returns the response object
//
// cmdlets are chained with |
func executeWithReturn[T *Q, Q any](runner powershell.Runner, response T, cmdlets ...powershell.Cmdlet) (T, error) {

	stdout, err := runner.RunWithResult(cmdlets...)

	if err != nil {
		return nil, fmt.Errorf("executeWithReturn: %w", err)
	}

	if stdout != "" {
		if IsSlice(response) && strings.HasPrefix(stdout, "{") {
			// Convert to list
			stdout = "[" + stdout + "]"
		}

		if err := json.Unmarshal([]byte(stdout), response); err != nil {
			return nil, fmt.Errorf("cannot unmarshal %s response: %w", cmdlets, err)
		}
	}

	return response, nil
}

func IsSlice[T any](v T) bool {
	t := reflect.TypeOf(v)

	// If it's a pointer, look at what it points to
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Kind() == reflect.Slice
}
