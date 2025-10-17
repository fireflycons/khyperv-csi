package models

type GetVHDRequest struct {
	Path       *string
	VmID       *string
	DiskNumber *int
}

type GetVHDResponse struct {
	// Path to the disk file
	Path string `json:"Path"`

	// Name of the disk
	Name string `json:"Name"`

	// Size in bytes of the disk
	Size int64 `json:"Size"`

	// UUID identifier of the disk
	DiskIdentifier string `json:"DiskIdentifier"`

	// UUID of the host to which the disk is attached, if it is attached.
	Host *string `json:"Host,omitempty"`
}

type ListVHDResponse struct {
	VHDs      []GetVHDResponse `json:"VHDs"`
	NextToken string           `json:"NextToken,omitempty"`
}
