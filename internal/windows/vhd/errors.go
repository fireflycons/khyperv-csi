//go:build windows

package vhd

import "errors"

var (
	ErrNilRequest      = errors.New("request is nil")
	ErrBadRequest      = errors.New("invalid request format")
	ErrDiskAttached    = errors.New("disk is attached to a host")
	ErrInvalidDiskPath = errors.New("invalid disk path")
	ErrInvalidDiskId   = errors.New("invalid disk id")
)
