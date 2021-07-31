package correlation

// Component represents a control plane component for
// correlating requests
type Component string

const (
	ComponentUnknown   = Component("unknown")
	ComponentCSIDriver = Component("csi-driver")
	ComponentSDK       = Component("sdk-server")
)
