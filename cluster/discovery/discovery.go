package discovery

// NodeEntry is used to discovery nodes in the cluster
type NodeEntry struct {
	Id            string
	Ip            string
	// TODO: Remove GossipVersion from here. Make sure gossip itself 
	// handles version checks when gossip starts one a node
	GossipVersion string
}

// ClusterInfo is the cluster info used for discovering nodes
type ClusterInfo struct {
	// Nodes is a list of nodes that are advertising their
	// presence in the cluster
	Nodes   map[string]NodeEntry
	// Version is a monotonically increasing number which gets
	// incremented for every update to the ClusterInfo structure.
	Version uint64
}

type WatchCB func(*ClusterInfo, error) error

type Cluster interface {
	// AddNode adds a new node into a cluster so that other can discover
	AddNode(dne NodeEntry) (*ClusterInfo, error)

	// RemoveNode removes a node from a cluster
	RemoveNode(dne NodeEntry) (*ClusterInfo, error)

	// Enumerate enumerates the nodes that have been discovered in the cluster
	Enumerate() (*ClusterInfo, error)

	// Watch starts a watch on the cluster and calls the provided
	// callback function when a node is added or removed.
	Watch(wcb WatchCB, index uint64) error
}
