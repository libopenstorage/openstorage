package bootstrap

// BootstrapNodeEntry is used to bootstrap nodes in the cluster
type BootstrapNodeEntry struct {
	Id string
	Ip string
	GossipVersion string
}
// BootstrapClusterInfo is the cluster info used while bootstrapping nodes
// and discovering peer nodes using gossip
type BootstrapClusterInfo struct {
	Size int
	Nodes map[string]BootstrapNodeEntry
	Version uint64
}

type WatchClusterCB func(*BootstrapClusterInfo, error) (error)

type ClusterBootstrap interface {
	// AddNode bootstrap a new node into a cluster
	AddNode(bne BootstrapNodeEntry) (*BootstrapClusterInfo, error)
	
	// RemoveNode removes a node from a cluster
	RemoveNode(bne BootstrapNodeEntry) (*BootstrapClusterInfo, error)

	// Enumerate enumerates the nodes that have been bootstrapped in the cluster
	Enumerate() (*BootstrapClusterInfo, error)

	// WatchCluster starts a watch on the cluster and calls the provided 
	// callback function when a node is added or removed.
	WatchCluster(wcb WatchClusterCB, index uint64) error
}
