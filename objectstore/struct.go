package objectstore

const (
	Enable        = "enable"
	VolumeName    = "name"
	ObjectStoreID = "id"
)

// Objectstoreinfo returns current objectstore details
// swagger:model
type ObjectstoreInfo struct {
	// UUID of objectstore
	UUID string
	// VolumeID of volume used by object store
	VolumeID string
	// Enable/Disable created objectstore
	Enabled bool
	// Status of objectstore running/failed
	Status string
	// Action being taken on this objectstore
	Action int
	// AccessKey for login into objectstore
	AccessKey string
	// SecretKey for login into objectstore
	SecretKey string
	// Endpoints for accessing objectstore
	Endpoints []string
	// CurrentEndpoint on which objectstore server is accessible
	CurrentEndpoint string
	// AccessPort is objectstore server port
	AccessPort int
	// Region for this objectstore
	Region string
}
