package models

type ControllerType int

const (
	IDE ControllerType = iota
	SCSI
	Unknown1
	Unknown2
)

type AttachedDrive struct {
	Path                          string `json:"Path"`
	DiskNumber                    *int   `json:"DiskNumber"`
	MaximumIOPS                   int    `json:"MaximumIOPS"`
	MinimumIOPS                   int    `json:"MinimumIOPS"`
	QoSPolicyID                   string `json:"QoSPolicyID"`
	SupportPersistentReservations bool   `json:"SupportPersistentReservations"`
	WriteHardeningMethod          int    `json:"WriteHardeningMethod"`
	ControllerLocation            int    `json:"ControllerLocation"`
	ControllerNumber              int    `json:"ControllerNumber"`
	ControllerType                int    `json:"ControllerType"`
	Name                          string `json:"Name"`
	PoolName                      string `json:"PoolName"`
	ID                            string `json:"Id"`
	VMID                          string `json:"VMId"`
	VMName                        string `json:"VMName"`
	VMSnapshotID                  string `json:"VMSnapshotId"`
	VMSnapshotName                string `json:"VMSnapshotName"`
	CimSession                    struct {
		ComputerName *string `json:"ComputerName"`
		InstanceID   string  `json:"InstanceId"`
	} `json:"CimSession"`
	ComputerName     string `json:"ComputerName"`
	IsDeleted        bool   `json:"IsDeleted"`
	VMCheckpointID   string `json:"VMCheckpointId"`
	VMCheckpointName string `json:"VMCheckpointName"`
}
