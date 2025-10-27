//go:build windows

package vhd

import (
	"fmt"
	"strings"

	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/powershell"
	"google.golang.org/grpc/codes"
)

// GetVMs lists all the VMs defined by Hyper-V
func GetVMs(runner powershell.Runner) (*rest.ListVMResponse, error) {

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

// GetVM gets a virtual machine by ID
func GetVM(runner powershell.Runner, id string) (*rest.GetVMResponse, error) {

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

	for _, vm := range vms.VMs {
		if strings.EqualFold(vm.ID, id) {
			return vm, nil
		}
	}

	return nil, &rest.Error{
		Code:    codes.NotFound,
		Message: fmt.Sprintf("VM %s not found", id),
	}
}
