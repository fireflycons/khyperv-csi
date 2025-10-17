package models

type GetSCSIControllerResopnse struct {
	ControllerNumber int             `json:"ControllerNumber"`
	IsTemplate       bool            `json:"IsTemplate"`
	Drives           []AttachedDrive `json:"Drives,omitempty"`
	Name             string          `json:"Name"`
	ID               string          `json:"Id"`
	VMID             string          `json:"VMId"`
	VMName           string          `json:"VMName"`
	VMSnapshotID     string          `json:"VMSnapshotId"`
	VMSnapshotName   string          `json:"VMSnapshotName"`
	ComputerName     string          `json:"ComputerName"`
	IsDeleted        bool            `json:"IsDeleted"`
	VMCheckpointID   string          `json:"VMCheckpointId"`
	VMCheckpointName string          `json:"VMCheckpointName"`
}
