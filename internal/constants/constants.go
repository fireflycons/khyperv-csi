package constants

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
)

const (
	// Default VHD format
	VhdType = ".vhdx"
)

const (
	DefaultServicePort = 8080
)

const (
	RepoName         = "khyperv-csi"
	PowerShellModule = RepoName
)

const (
	//nolint:gosec // this is a header name not a value
	ApiKeyHeader = "X-Api-Key"
)
