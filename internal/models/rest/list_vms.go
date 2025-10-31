package rest

type GetVMResponse struct {
	Name       string `json:"Name"`
	ID         string `json:"Id"`
	Path       string `json:"Path"`
	Generation int    `json:"Generation"`
}

type ListVMResponse struct {
	VMs []*GetVMResponse
}
