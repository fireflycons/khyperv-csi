package constants

import "time"

const (
	_   = iota
	KiB = 1 << (10 * iota)
	MiB
	GiB
	TiB
)

const (
	ServiceName = "khypervprovider"

	ServiceDisplayName = "Persistent Volume provider for Kubernetes"

	ServiceDescription = "Provides an interface to Hyper-V virtual disks for Kubernetes Persistent Volumes"

	HyperVServiceName = "vmcompute"
)

const (
	ZeroUUID = "00000000-0000-0000-0000-000000000000"
)

const (
	// minimumVolumeSizeInBytes is used to validate that the user is not trying
	// to create a volume that is smaller than what we support
	MinimumVolumeSizeInBytes int64 = 5 * MiB

	// maximumVolumeSizeInBytes is used to validate that the user is not trying
	// to create a volume that is larger than what we support
	MaximumVolumeSizeInBytes int64 = 2 * TiB

	// defaultVolumeSizeInBytes is used when the user did not provide a size or
	// the size they provided did not satisfy our requirements
	DefaultVolumeSizeInBytes int64 = 16 * GiB

	// createdByDO is used to tag volumes that are created by this CSI plugin
	CreatedByDO = "Created by DigitalOcean CSI driver"

	// doAPITimeout sets the timeout we will use when communicating with the
	// Digital Ocean API. NOTE: some queries inherit the context timeout
	DoAPITimeout = 10 * time.Second

	// maxVolumesPerDropletErrorLegacyMessage is the old error message returned by
	// the DO API when the per-droplet volume limit would be exceeded.
	MaxVolumesPerDropletErrorLegacyMessage = "cannot attach more than 7 volumes to a single Droplet"

	// maxVolumesPerDropletErrorMessage is the error message returned by the DO API
	// when the per-droplet volume limit would be exceeded.
	MaxVolumesPerDropletErrorMessage = "cannot attach more volumes to the Droplet"
)

const (
	// Default VHD format
	VhdType = ".vhdx"
)

const (
	DefaultServicePort = 8080
)

const (
	RepoName = "khyperv-csi"
	PowerShellModule = RepoName
)
