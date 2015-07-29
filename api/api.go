package api

type Options map[OptionKey]interface{}
type OptionKey string

const (
	OptName        = OptionKey("Name")
	OptID          = OptionKey("ID")
	OptVolumeLabel = OptionKey("VolumeLabel")
	OptConfigLabel = OptionKey("ConfigLabel")
)

type DriverStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type VolumeCreateRequest struct {
	Locator VolumeLocator  `json:"locator"`
	Options *CreateOptions `json:"options,omitempty"`
	Spec    *VolumeSpec    `json:"spec,omitempty"`
}

type VolumeCreateResponse struct {
	ID     VolumeID `json:"id"`
	Status string   `json:"status"`
}

type VolumeStateRequest struct {
	Format bool   `json:"format"`
	Attach bool   `json:"attach"`
	Mount  bool   `json:"mount"`
	Path   string `json:"path"`
}

type VolumeStateResponse struct {
	VolumeStateRequest
	Status string `json:"status"`
}

type VolumeResponse struct {
	Status string `json:"status"`
}

type SnapCreateRequest struct {
	ID     VolumeID `json:"id"`
	Labels Labels   `json:"labels"`
}

type SnapCreateResponse struct {
	ID     SnapID `json:"id"`
	Status string `json:"status"`
}

func ResponseStatusNew(err error) VolumeResponse {
	if err == nil {
		return VolumeResponse{}
	}
	return VolumeResponse{Status: err.Error()}
}
