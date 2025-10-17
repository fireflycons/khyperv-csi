package rest

// GetCapacityResponse is the REST objext reurned by the GetCapacity call
type GetCapacityResponse struct {
	// AvailableCapacity is the available space in bytes
	// on the disk where the PV Store resides for creating
	// new persistent volumes.
	AvailableCapacity int64

	// MinimumVolumeSize is the minimum size of a volume that can be provisioned.
	// Requests for smaller volumes will result in a volume of this size being provisioned.
	MinimumVolumeSize int64
}
