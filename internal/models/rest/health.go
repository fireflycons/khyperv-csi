package rest

type HealthyResponse struct {
	// Status indicates the health status of the service
	Status string `json:"status"`
}
