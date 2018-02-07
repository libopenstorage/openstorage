// package configutil is a higher level abstraction on osdconfig.
// It allows easier interface to setup kvdb watches
package configutil

import (
	"context"

	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/portworx/kvdb"
)

// NewManager instantiates a new osdconfig manager
func NewManager(kv kvdb.Kvdb) (osdconfig.ConfigManager, error) {
	return osdconfig.NewManager(context.Background(), kv)
}

// NewManagerWithContext instantiates a new osdconfig manager
func NewManagerWithContext(ctx context.Context, kv kvdb.Kvdb) (osdconfig.ConfigManager, error) {
	return osdconfig.NewManager(ctx, kv)
}

// WatchCluster registers a function literal to watch on cluster level changes
func WatchCluster(manager osdconfig.ConfigManager, name string, cb func(clusterConfig *osdconfig.ClusterConfig) error) error {
	f, _ := osdconfig.GetCallback(name, cb)

	if err := manager.Register(name, osdconfig.ClusterWatcher, nil, f); err != nil {
		return err
	}

	return nil
}

// WatchNodes registers a function literal to watch on nodes level changes
func WatchNode(manager osdconfig.ConfigManager, name string, cb func(nodeConfig *osdconfig.NodeConfig) error) error {
	f, _ := osdconfig.GetCallback(name, cb)

	if err := manager.Register(name, osdconfig.NodeWatcher, nil, f); err != nil {
		return err
	}

	return nil
}
