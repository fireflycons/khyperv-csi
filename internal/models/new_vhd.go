package models

// NewVHDRequest is passed to the New-VHD call
type NewVHDRequest struct {
	Path string
	Size uint64
}

// // NewVHDResponse is returned by the New-VHD call
// type NewVHDResponse struct {
// 	ComputerName            string  `json:"ComputerName"`
// 	Path                    string  `json:"Path"`
// 	VhdFormat               int     `json:"VhdFormat"`
// 	VhdType                 int     `json:"VhdType"`
// 	FileSize                uint64  `json:"FileSize"`
// 	Size                    uint64  `json:"Size"`
// 	MinimumSize             *uint64 `json:"MinimumSize,omitempty"`
// 	LogicalSectorSize       int     `json:"LogicalSectorSize"`
// 	PhysicalSectorSize      int     `json:"PhysicalSectorSize"`
// 	BlockSize               int     `json:"BlockSize"`
// 	ParentPath              string  `json:"ParentPath"`
// 	DiskIdentifier          string  `json:"DiskIdentifier"`
// 	FragmentationPercentage int     `json:"FragmentationPercentage"`
// 	Alignment               int     `json:"Alignment"`
// 	Attached                bool    `json:"Attached"`
// 	DiskNumber              *int    `json:"DiskNumber,omitempty"`
// 	IsPMEMCompatible        bool    `json:"IsPMEMCompatible"`
// 	AddressAbstractionType  int     `json:"AddressAbstractionType"`
// 	Number                  *int    `json:"Number,omitempty"`
// }
