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

	vmName, err := s.FindMeta("VirtualMachineName")
	require.NoError(t, err, "FindMeta VirtualMachineName failed")
	t.Logf("VirtualMachineName: %s", vmName)

	vmID, err := s.FindMeta("VirtualMachineId")
	require.NoError(t, err, "FindMeta VirtualMachineId failed")
	t.Logf("VirtualMachineId: %s", vmID)
}
