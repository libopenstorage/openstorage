package discovery

// NodeEntry is used to discovery nodes in the cluster
type NodeEntry struct {
	Id string
	Ip string
	GossipVersion string
}
// ClusterInfo is the cluster info used while discoveryping nodes
// and discovering peer nodes using gossip
type ClusterInfo struct {
	Size int
	Nodes map[string]NodeEntry
	Version uint64
}

type WatchClusterCB func(*ClusterInfo, error) (error)

type Cluster interface {
	// AddNode adds a new node into a cluster so that other can discover
	AddNode(dne NodeEntry) (*ClusterInfo, error)
	
	// RemoveNode removes a node from a cluster
	RemoveNode(dne NodeEntry) (*ClusterInfo, error)

	// Enumerate enumerates the nodes that have been discovered in the cluster
	Enumerate() (*ClusterInfo, error)

	// WatchCluster starts a watch on the cluster and calls the provided 
	// callback function when a node is added or removed.
	WatchCluster(wcb WatchClusterCB, index uint64) error
}
