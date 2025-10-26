package vhd

import (
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
)

// GetVirtualMachines lists all the VMs defined by Hyper-V
func GetVirtualMachines(runner powershell.Runner) (*rest.ListVMResponse, error) {

	vms, err := executeWithReturn(
		runner,
		&rest.ListVMResponse{},
		powershell.NewCmdlet(
			"Get-PVVirtualMachines",
			nil,
		),
	)

	if err != nil {
		return nil, err
	}

	return vms, nil
}
