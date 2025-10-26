package rest

import "github.com/fireflycons/hypervcsi/internal/models"

type ListVMResponse struct {
	VMs []*models.GetVMResponse
}
