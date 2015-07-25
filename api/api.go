package api

type StatusReason string

const (
	StatusReasonUnknown StatusReason = "Unknown"
)

const (
	StatusSuccess = "OK"
	StatusFail    = "FAIL"
)

var ResponseSuccess ResponseStatus = ResponseStatus{
	Status:    StatusSuccess,
	Reason:    StatusReason(StatusSuccess),
	ErrorCode: 0,
}

type Options map[OptionKey]interface{}
type OptionKey string

const (
	OptName        = OptionKey("Name")
	OptID          = OptionKey("ID")
	OptDiskLabel   = OptionKey("DiskLabel")
	OptConfigLabel = OptionKey("ConfigLabel")
)

type DriverStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type ResponseStatus struct {
	Status    string       `json:"status"`
	Reason    StatusReason `json:"reason"`
	ErrorCode int          `json:"errorCode"`
}

type VolumeCreateRequest struct {
	Locator VolumeLocator  `json:"locator"`
	Options *CreateOptions `json:"options,omitempty"`
	Spec    *VolumeSpec    `json:"spec,omitempty"`
}

type VolumeCreateResponse struct {
	Status ResponseStatus
	ID     VolumeID
}

type VolumeAttachRequest struct {
	ID   VolumeID `json:"ID"`
	Path string   `json:"path,omitempty"`
}

type VolumeAttachResponse struct {
	Status ResponseStatus
	Path   string
}

func ResponseStatusNew(err error) ResponseStatus {
	if err == nil {
		return ResponseSuccess
	}
	return ResponseStatus{
		Status: StatusFail,
		Reason: StatusReason(err.Error())}
}
