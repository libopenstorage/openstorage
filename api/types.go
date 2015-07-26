package api

import (
	"time"
)

type VolumeID string
type SnapID string
type VolumeCos int

const (
	VolumeCosNone = VolumeCos(0)
	VolumeCosMax  = VolumeCos(9)
)

type Status string

const (
	NotPresent = Status("NotPresent")
	Up         = Status("Up")
	Down       = Status("Down")
	Degraded   = Status("Degraded")
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

const VolumeStateAny = VolumePending | VolumeAvailable | VolumeAttached | VolumeDetached | VolumeError | VolumeDeleted

type Labels map[string]string

// VolumeLocator is a structure that is attached to a volume and is used to
// carry opaque metadata.
type VolumeLocator struct {
	Name         string
	VolumeLabels Labels
}

type CreateOptions struct {
	FailIfExists   bool
	CreateFromSnap SnapID
}

type Filesystem string

const (
	FsXfs  = Filesystem("xfs")
	FsExt4 = Filesystem("ext4")
	FsZfs  = Filesystem("zfs")
	FsNone = Filesystem("none")
)

type VolumeAction int

const (
	VolumeActionCreate = 1 << iota
	VolumeActionUpdate
	VolumeActionDelete
)

// VolumeInfo carrys the runtime information on a created volume.
type VolumeInfo struct {
	// RequestID to synchronize on outstanding requests
	RequestID int64
	// Action
	Action VolumeAction
	// Current State of the volume
	DeviceInfo Volume
}

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
	// Device Minor
	Minor int32
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
	VolumeStatus Status
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
	// usage
	Usage uint64
}

type VolumeStats struct {
	// TODO
}

type VolumeAlerts struct {
	// TODO
}
