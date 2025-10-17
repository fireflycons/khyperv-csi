//go:build windows

package vhd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSlice(t *testing.T) {

	type intSlice []int

	var s1 []int
	var s2 intSlice
	var s3 *intSlice

	require.True(t, IsSlice(s1), "[]int is not slice")
	require.True(t, IsSlice(s2), "intSlice is not slice")
	require.True(t, IsSlice(s3), "*intSlice is not slice")

}
