package api

// OptionKey specifies a set of recognized query params
type OptionKey string

const (
	// OptName query parameter used to lookup volume by name
	OptName = OptionKey("Name")
	// OptVolumeID query parameter used to lookup volume by ID.
	OptVolumeID = OptionKey("VolumeID")
	// OptLabel query parameter used to lookup volume by set of labels.
	OptLabel = OptionKey("Label")
	// OptConfigLabel query parameter used to lookup volume by set of labels.
	OptConfigLabel = OptionKey("ConfigLabel")
)

// VolumeCreateRequest is the body of create REST request
type VolumeCreateRequest struct {
	// Locator user specified volume name and labels.
	Locator VolumeLocator `json:"locator"`
	// Source to create volume
	Source *Source `json:"source,omitempty"`
	// Spec is the storage spec for the volume
	Spec *VolumeSpec `json:"spec,omitempty"`
}

// VolumeCreateResponse is the body of create REST response
type VolumeCreateResponse struct {
	// ID of the newly created volume
	ID VolumeID `json:"id"`
	VolumeResponse
}

// VolumeActionParam desired action on volume
type VolumeActionParam int

const (
	// ParamIgnore user should ignore the value of the parameter.
	ParamIgnore VolumeActionParam = iota
	// ParamOff maps to the boolean value false.
	ParamOff
	// ParamOn maps to the boolean value true.
	ParamOn
)

// VolumeStateAction is the body of the REST request to specify desired actions
type VolumeStateAction struct {
	// Format volume
	Format VolumeActionParam `json:"format"`
	// Attach or Detach volume
	Attach VolumeActionParam `json:"attach"`
	// Mount or unmount volume
	Mount VolumeActionParam `json:"mount"`
	// MountPath
	MountPath string `json:"mount_path"`
	// DevicePath returned in Attach
	DevicePath string `json:"device_path"`
}

// VolumeStateResponse is the body of the REST response
type VolumeStateResponse struct {
	// VolumeStateRequest the current state of the volume
	VolumeStateAction
	VolumeResponse
}

// VolumeResponse is embedded in all REST responses.
type VolumeResponse struct {
	// Error is "" on success or contains the error message on failure.
	Error string `json:"error"`
}

// SnapCreateRequest request body to create a snap.
type SnapCreateRequest struct {
	ID       VolumeID      `json:"id"`
	Locator  VolumeLocator `json:"locator"`
	Readonly bool          `json:"readonly"`
}

// SnapCreateResponse response body to SnapCreateRequest
type SnapCreateResponse struct {
	VolumeCreateResponse
}

// ResponseStatusNew create VolumeResponse from error
func ResponseStatusNew(err error) VolumeResponse {
	if err == nil {
		return VolumeResponse{}
	}
	return VolumeResponse{Error: err.Error()}
}
