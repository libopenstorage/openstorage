package api

import (
	"time"
)

// VolumeID driver specific system wide unique volume identifier.
type VolumeID string

// SnapID driver specific system wide unique snap identifier.
type SnapID string

// VolumeCos a number representing class of servcie.
type VolumeCos int

const (
	// VolumeCosNone minmum level of CoS
	VolumeCosNone = VolumeCos(0)
	// VolumeCosMedum in-between level of Cos
	VolumeCosMedium = VolumeCos(5)
	// VolumeCosNone maximum level of CoS
	VolumeCosMax = VolumeCos(9)
)

// VolumeStatus a health status.
type VolumeStatus string

const (
	// NotPresent This volume is not present.
	NotPresent = VolumeStatus("NotPresent")
	// Up status healthy
	Up = VolumeStatus("Up")
	// Down status failure.
	Down = VolumeStatus("Down")
	// Down status up but with degraded performance. In a RAID group, this may indicate a problem with one or more drives
	Degraded = VolumeStatus("Degraded")
)

// VolumeState is one of the below enumerations and reflects the state
// of a volume.
type VolumeState int

const (
	// VolumePending is being created
	VolumePending VolumeState = 1 << iota
	// VolumeAvailable disk is ready to be assigned to a container
	VolumeAvailable
	// VolumeAttached is attached to container
	VolumeAttached
	// VolumeDetached is detached but associated with a container.
	VolumeDetached
	// VolumeError is in Error State
	VolumeError
	// VolumeDeleted is deleted, it will remain in this state while resources are
	// asynchronously reclaimed.
	VolumeDeleted
)

// VolumeStateAny a filter that selects all volumes
const VolumeStateAny = VolumePending | VolumeAvailable | VolumeAttached | VolumeDetached | VolumeError | VolumeDeleted

// Labels a name-value map
type Labels map[string]string

// VolumeLocator is a structure that is attached to a volume and is used to
// carry opaque metadata.
type VolumeLocator struct {
	// Name user friendly identifier
	Name string
	// VolumeLabels set of name-value pairs that acts as search filters.
	VolumeLabels Labels
}

// CreateOptions are passed in with a CreateRequest
type CreateOptions struct {
	// FailIfExists fail create request if a volume with matching Locator already exists.
	FailIfExists bool
	// CreateFromSnap will create a volume with specified SnapID
	CreateFromSnap SnapID
}

// Filesystem supported filesystems
type Filesystem string

const (
	// FsXfs the XFS filesystem
	FsXfs = Filesystem("xfs")
	// FsExt4 the EXT4 filesystem
	FsExt4 = Filesystem("ext4")
	// FsZfs the ZFS filesystem
	FsZfs = Filesystem("zfs")
	// FsNone no file system, applicable for raw block devices.
	FsNone = Filesystem("none")
)

// VolumeSpec has the properties needed to create a volume.
type VolumeSpec struct {
	// Ephemeral storage
	Ephemeral bool
	// Thin provisioned volume size in bytes
	Size uint64
	// Format disk with this FileSystem
	Format Filesystem
	// BlockSize for file system
	BlockSize int
	// HA Level specifies the number of nodes that are
	// allowed to fail, and yet data is availabel.
	// A value of 0 implies that data is not erasure coded,
	// a failure of a node will lead to data loss.
	HALevel int
	// This disk's CoS
	Cos VolumeCos
	// Perform dedupe on this disk
	Dedupe bool
	// SnapShotInterval in minutes, set to 0 to disable Snapshots
	SnapShotInterval int
	// Volume configuration labels
	ConfigLabels Labels
}

// Volume represents a live, created volume.
type Volume struct {
	// Self referential VolumeID
	ID VolumeID
	// User specified locator
	Locator VolumeLocator
	// Volume creation time
	Ctime time.Time
	// User specified disk configuration
	Spec *VolumeSpec
	// Volume usage
	Usage uint64
	// Last time a scan was run
	LastScan time.Time
	// Filesystem type if any
	Format Filesystem
	// Volume Status
	Status VolumeStatus
	// VolumeState
	State VolumeState
	// Attached On - for clustered storage arrays
	AttachedOn interface{}
	// Device path
	DevicePath string
	// Attach path
	AttachPath string
	// Set of nodes no which this Volume is erasure coded - for clustered storage arrays
	ReplicaSet []interface{}
	// Last Recorded Error
	ErrorNum int
	// Error String
	ErrorString string
}

// VolumeSnap identifies a volume snapshot.
type VolumeSnap struct {
	// System generated snap label
	SnapID SnapID
	// Volume identifier.
	VolumeID VolumeID
	// Snap creation time.
	Ctime time.Time
	// User specfied label
	UserLabel string
	// Usage
	Usage uint64
}

// VolumeStats
type VolumeStats struct {
	// TODO
}

// VolumeAlerts
type VolumeAlerts struct {
	// TODO
}
