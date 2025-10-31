package rest

// GetVolumeResponse is the response returned when a volume is created or fetched.
type GetVolumeResponse struct {

	// The name of the volume.
	Name string `json:"name"`

	// The GUID ID assigned to the volume by Hyper-V
	ID string `json:"id"`

	// Actual size of the created volume.
	// If caller requests less than the minimum VHD size,
	// then this will be the minimum VHD size.
	Size int64 `json:"size"`
}
