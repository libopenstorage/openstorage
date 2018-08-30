package osdconfig

import "errors"

var (
	// ErrNotImplemented is returned when any of the ConfigCaller APIs is not
	// implemented
	ErrNotImplemented = errors.New("configCaller API not implemented")
)

// NullConfigCaller is a NULL implementation of the ConfigCaller interface
type NullConfigCaller struct {
}

// GetClusterConf fetches cluster configuration data from a backend such as kvdb
func (n *NullConfigCaller) GetClusterConf() (*ClusterConfig, error) {
	return nil, ErrNotImplemented
}

// GetNodeConf fetches node configuration data using node id
func (n *NullConfigCaller) GetNodeConf(nodeID string) (*NodeConfig, error) {
	return nil, ErrNotImplemented
}

// EnumerateNodeConf fetches data for all nodes
func (n *NullConfigCaller) EnumerateNodeConf() (*NodesConfig, error) {
	return nil, ErrNotImplemented
}

// SetClusterConf pushes cluster configuration data to the backend
// It is assumed that the backend will notify the implementor of this interface
// when a change is triggered
func (n *NullConfigCaller) SetClusterConf(config *ClusterConfig) error {
	return ErrNotImplemented
}

// SetNodeConf pushes node configuration data to the backend
// It is assumed that the backend will notify the implementor of this interface
// when a change is triggered
func (n *NullConfigCaller) SetNodeConf(config *NodeConfig) error {
	return ErrNotImplemented
}

// DeleteNodeConf removes node config for a particular node
func (n *NullConfigCaller) DeleteNodeConf(nodeID string) error {
	return ErrNotImplemented
}
