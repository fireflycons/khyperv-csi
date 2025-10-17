//go:build windows

package vhd

import (
	"path/filepath"
	"regexp"
)

var diskNameRx = regexp.MustCompile(`^(?P<name>[A-Za-z0-9._-]+);(?P<id>[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})\.vhdx?$`)

func ParseDiskPath(path string) (name, id string, err error) {

	filename := filepath.Base(path)
	matches := diskNameRx.FindStringSubmatch(filename)
	if len(matches) != 3 {
		return "", "", ErrInvalidDiskPath
	}
	return matches[1], matches[2], nil
}
