// osdconfig is a package to work with distributed config parameters
package osdconfig

// A config manager interface allows management of osdconfig parameters
// It defines setters, getters and callback management functions
type ConfigManager interface {
	// GetClusterConf fetches cluster configuration data from a backend such as kvdb
	GetClusterConf() (*ClusterConfig, error)

	// GetNodeConf fetches node configuration data using node id
	GetNodeConf(nodeID string) (*NodeConfig, error)

	// SetClusterConf pushes cluster configuration data to the backend
	// It is assumed that the backend will notify the implementor of this interface
	// when a change is triggered
	SetClusterConf(config *ClusterConfig) error

	// SetNodeConf pushes node configuration data to the backend
	// It is assumed that the backend will notify the implementor of this interface
	// when a change is triggered
	SetNodeConf(config *NodeConfig) error

	// WatchCluster registers a user defined function as callback watching for changes
	// in the cluster configuration
	WatchCluster(name string, cb func(config *ClusterConfig) error) error

	// WatchNode registers a user defined function as callback watching for changes
	// in the node configuration
	WatchNode(name string, cb func(config *NodeConfig) error) error

	// Close performs internal cleanup
	Close()
}
