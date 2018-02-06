// osdconfig is a package to work with distributed config parameters
package osdconfig

import "context"

// A config manager interface allows management of osdconfig parameters
// It defines setters, getters and callback management functions
type ConfigManager interface {
	// GetClusterConf fetches cluster configuration data from a backend such as kvdb
	GetClusterConf() (*ClusterConfig, error)

	// Fetch node configuration data using node id
	GetNodeConf(nodeID string) (*NodeConfig, error)

	// SetClusterConf pushes cluster configuration data to the backend
	// It is assumed that the backend will notify the implementor of this interface
	// when a change is triggered
	SetClusterConf(config *ClusterConfig) error

	// SetNodeConf pushes node configuration data to the backend
	// It is assumed that the backend will notify the implementor of this interface
	// when a change is triggered
	SetNodeConf(config *NodeConfig) error

	// Register callback functions that will get triggered on a change to either
	// cluster configuration of node configuration data in the backend.
	// The callback function returns two channels: A channel to write to and a
	// channel to read on. Context is provided to the callback function in order
	// to manage cleanup during context cancellations
	Register(name string,
		opt interface{},
		cb func(ctx context.Context, opt interface{}) (chan<- *DataToCallback,
			<-chan *DataFromCallback)) error

	// Run executes callback functions, however, it is expected that users never
	// call this function directly. Run is called behind the scenes using scheduler
	Run(wd *DataToCallback)

	// GetContext returns the context during execution of callback functions
	// Please note that the context is renewed prior to new callback execution cycle
	// A good practice is to call this function directly in select statements
	// e.g.: case <- manager.GetContext().Done():
	GetContext() context.Context

	// GetStatus returns execution status
	// Please note that this function delivers a copy of internal status map
	// and internal memory is cleared prior to every new callback execution cycle
	GetStatus() map[string]*Status

	// Error returns exec error if any
	// This function should return an error during occurred previous callback execution cycle
	Error() error

	// Abort sends context cancellation to execution of callback functions
	// This function should return immediately and not block
	Abort()

	// Wait waits while scheduler is busy processing callbacks
	// This function should block and only return when (say) internal locks are released
	Wait()

	// Close and perform cleanup, if any
	// This function should cancel all derivative contexts and clean up internal memory
	// It is expected that a call to Close() should lead to graceful shutdown
	Close()
}
