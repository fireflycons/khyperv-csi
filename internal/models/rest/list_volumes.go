package rest

import "github.com/fireflycons/hypervcsi/internal/models"

type ListVolumesResponse struct {

	// List of volumes found in the PV Store
	Volumes []models.GetVHDResponse `json:"volumes"`

	// If there are more entries in the list, this token can be used to fetch the next set of entries.
	NextToken string `json:"next_token,omitempty"`
}
