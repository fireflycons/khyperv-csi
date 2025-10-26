package controller

import (
	"github.com/fireflycons/hypervcsi/internal/models/rest"
	"github.com/fireflycons/hypervcsi/internal/windows/messages"
	"github.com/fireflycons/hypervcsi/internal/windows/vhd"
	"github.com/sirupsen/logrus"
)

func (s *controllerServer) ListVms() (*rest.ListVMResponse, error) {

	log := s.log.WithFields(logrus.Fields{
		"method": "list_vms",
	})

	log.Info(messages.CONTROLLER_LIST_VMS)

	vms, err := vhd.GetVirtualMachines(s.runner)

	if err != nil {
		return nil, s.processError(err, log, messages.CONTROLLER_LIST_VMS_FAILED)
	}

	log.Info(messages.CONTROLLER_VMS_LISTED)

	return vms, nil
}
