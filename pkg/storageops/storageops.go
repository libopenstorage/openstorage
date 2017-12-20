package storageops

const (
	// SetIdentifierNone is a default identifier to group all disks from a
	// particular set
	SetIdentifierNone = "None"
)

// Custom storage operation error codes.
const (
	_ = iota + 5000
	// ErrVolDetached is code for a volume is detached on the instance
	ErrVolDetached
	// ErrVolInval is the code for a invalid volume
	ErrVolInval
	// ErrVolAttachedOnRemoteNode is code when a volume is not attached locally
	// but attached on a remote node
	ErrVolAttachedOnRemoteNode
)

// StorageError error returned for storage operations
type StorageError struct {
	// Code is one of storage operation driver error codes.
	Code int
	// Msg is human understandable error message.
	Msg string
	// Instance provides more information on the error.
	Instance string
}

// Ops interface to perform basic storage operations.
type Ops interface {
	// Name returns name of the storage operations driver
	Name() string
	// Create volume based on input template volume and also apply given labels.
	Create(template interface{}, labels map[string]string) (interface{}, error)
	// GetDeviceID returns ID/Name of the given device/disk
	GetDeviceID(template interface{}) string
	// Attach volumeID.
	// Return attach path.
	Attach(volumeID string) (string, error)
	// Detach volumeID.
	Detach(volumeID string) error
	// Delete volumeID.
	Delete(volumeID string) error
	// FreeDevices returns free block devices on the instance.
	// blockDeviceMappings is a data structure that contains all block devices on
	// the instance and where they are mapped to
	FreeDevices(blockDeviceMappings []interface{}, rootDeviceName string) ([]string, error)
	// Inspect volumes specified by volumeID
	Inspect(volumeIds []*string) ([]interface{}, error)
	// DeviceMappings returns map[local_attached_volume_path]->volume ID/NAME
	DeviceMappings() (map[string]string, error)
	// Enumerate volumes that match given filters. Organize them into
	// sets identified by setIdentifier.
	// labels can be nil, setIdentifier can be empty string.
	Enumerate(volumeIds []*string,
		labels map[string]string,
		setIdentifier string,
	) (map[string][]interface{}, error)
	// DevicePath for the given volume i.e path where it's attached
	DevicePath(volume interface{}) (string, error)
	// Snapshot the volume with given volumeID
	Snapshot(volumeID string, readonly bool) (interface{}, error)
	// ApplyTags will apply given labels/tags on the given volume
	ApplyTags(v interface{}, labels map[string]string) error
	// RemoveTags removes labels/tags from the given volume
	RemoveTags(v interface{}, labels map[string]string) error
	// Tags will list the existing labels/tags on the given volume
	Tags(v interface{}) (map[string]string, error)
}

// NewStorageError creates a new custom storage error instance
func NewStorageError(code int, msg string, instance string) error {
	return &StorageError{Code: code, Msg: msg, Instance: instance}
}

func (e *StorageError) Error() string {
	return e.Msg
}
