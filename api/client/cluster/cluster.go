package cluster

import (
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/api/client"
)

// ClusterManager returns a REST wrapper for the Cluster interface.
func ClusterManager(c *client.Client) cluster.Cluster {
	return newClusterClient(c)
}
