//go:build windows

package controller

import (
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
)

func (s *controllerServer) GetVm(nodeId string) (*rest.GetVMResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"method": "get_vm",
	})

	log.Info(messages.CONTROLLER_GET_VM)

	vm, err := vhd.GetVM(s.runner, nodeId)

	if err != nil {
		return nil, s.processError(err, log, messages.CONTROLLER_GET_VM_FAILED)
	}

	log.Info(messages.CONTROLLER_GOT_VM)

	return vm, nil
}
