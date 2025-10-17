package models

type GetVMIdResponse struct {
	VMId string `json:"vmId"`
}

type GetVMResponse struct {
	Name       string `json:"Name"`
	ID         string `json:"Id"`
	Path       string `json:"Path"`
	Generation int    `json:"Generation"`
}
