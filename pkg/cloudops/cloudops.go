package cloudops

// CloudObjectMeta encapsulates generic metadata information about a cloud object
type CloudObjectMeta struct {
	// Name is the name of the object
	Name string
	// ID is the object ID
	ID string
	// Labels are labels attached with the object
	Labels map[string]string
	// Zone is the cloud zone the object resides in
	Zone string
	// Region is the cloud regison the object resides in
	Region string
}

// InstanceGroupInfo encapsulates info for a cloud instance group
type InstanceGroupInfo struct {
	CloudObjectMeta
	// AutoscalingEnabled is true if auto scaling is turned on
	AutoscalingEnabled bool
	// Min is the min number of nodes in the instance group
	Min int64
	// Max is the max number of nodes in the instance group
	Max int64
	// Zones are the zones that the instance group is part of
	Zones []string
}

// InstanceInfo encapsulates info for a cloud instance
type InstanceInfo struct {
	CloudObjectMeta
}

// Ops is the interface to interact with cloud instances
type Ops interface {
	// InspectSelf inspects the node where the code is invoked. Hence this assumes
	// this is invoked from withing a cloud instance
	InspectSelf() (*InstanceInfo, error)
	// InspectSelfInstanceGroup inspects the instance group within which the current
	// instance resides
	InspectSelfInstanceGroup() (*InstanceGroupInfo, error)
}
