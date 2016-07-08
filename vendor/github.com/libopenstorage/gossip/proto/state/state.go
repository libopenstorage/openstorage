package state

import (
	"github.com/libopenstorage/gossip/types"
)

// State is a node's state at any given point of time.
type State interface {
	// The functions defined by the interface are the events that might
	// cause the State to change. Following are the current states
	// - UP
	// - DOWN
	// - NOT_IN_QUORUM
	// - SUSPECT_NOT_IN_QUORUM

	// SelfAlive is an event when memberlist indicates self node is alive
	SelfAlive(nodeInfoMap types.NodeInfoMap) (State, error)

	// NodeAlive is an event when memberlist indicates another node is alive
	NodeAlive(nodeInfoMap types.NodeInfoMap) (State, error)

	// SelfLeave is an event when memberlist indicates self node leaves
	SelfLeave() (State, error)

	// NodeLeave is an event when memberlist indicates another node has left
	NodeLeave(nodeInfoMap types.NodeInfoMap) (State, error)

	// UpdateClusterSize is an event indicating the change in cluster size
	UpdateClusterSize(clusterSize int, nodeInfoMap types.NodeInfoMap) (State, error)

	// Timeout is an event triggered when quorum timeout has reached
	Timeout(clusterSize int, nodeInfoMap types.NodeInfoMap) (State, error)

	// String
	String() string

	// NodeStatus
	NodeStatus() types.NodeStatus
}
