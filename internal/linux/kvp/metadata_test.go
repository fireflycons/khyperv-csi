//go:build linux

package kvp_test

import (
	"testing"

	"github.com/fireflycons/hypervcsi/internal/linux/kvp"
	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {

	s := kvp.New()

	if !s.IsPresent() {
		t.Skip("Hyper-V metadata service not available. Probably not running in a Hyper-V VM.")
	}

	vmName, err := s.Find(kvp.VM_NAME_KEY)
	require.NoError(t, err, "FindMeta VirtualMachineName failed")
	t.Logf("VirtualMachineName: %s", vmName)

	vmID, err := s.Find(kvp.VM_ID_KEY)
	require.NoError(t, err, "FindMeta VirtualMachineId failed")
	t.Logf("VirtualMachineId: %s", vmID)
}
