package gossip

import (
	"github.com/libopenstorage/gossip/proto"
	"github.com/libopenstorage/gossip/types"
	"time"
)

type GossipStore interface {
	// types.NodeId of this Store
	NodeId() types.NodeId

	// Update updates the value for this node.
	// Side-effects include updating the last update ts
	// for this node.
	UpdateSelf(types.StoreKey, interface{})

	// GetStoreValue returns the StoreValue associated with
	// the given key
	GetStoreKeyValue(key types.StoreKey) types.NodeInfoMap

	// GetStoreKeys returns all the keys present in the store
	GetStoreKeys() []types.StoreKey

	// Used for gossiping

	// Update updates the current state of the gossip data
	// with the newly available data
	Update(newData types.StoreDiff)

	// Subset returns the available gossip data for the given
	// nodes. Node data is returned if there is none available
	// for a given node
	Subset(nodes types.StoreNodes) types.StoreDiff

	// MetaInfoMap returns meta information for the
	// current available data
	MetaInfo() types.StoreMetaInfo

	// Diff returns a tuple of lists, where
	// first list is of the names of node for which
	// the current data is older as compared to the
	// given meta info, and second list is the names
	// of nodes for which the current data is newer
	Diff(d types.StoreMetaInfo) (types.StoreNodes, types.StoreNodes)

	// UpdateNodeStatuses updates the statuses of
	// the nodes this node has information about
	UpdateNodeStatuses(time.Duration)
}

type Gossiper interface {
	// Gossiper has a gossip store
	GossipStore

	// Start begins the gossip protocol
	Start()

	// SetGossipInterval sets the gossip interval
	SetGossipInterval(time.Duration)
	// GossipInterval gets the gossip interval
	GossipInterval() time.Duration

	// SetNodeDeathInterval sets the duration which is used
	// to determine if peer node is alive. If the last update
	// timestamp of peer is older than this interval,
	// then we declare the node to be down
	SetNodeDeathInterval(t time.Duration)

	// NodeDeathInterval returns the duration which is
	// used to determine if the peer node is alive.
	NodeDeathInterval() time.Duration

	// Stop stops the gossiping
	Stop()

	// AddNode adds a node to gossip with
	AddNode(ip string) error

	// RemoveNode removes the node to gossip with
	RemoveNode(ip string) error

	// GetNodes returns a list of the connection addresses
	// added via AddNode
	GetNodes() []string
}

// New returns an initialized Gossip node
// which identifies itself with the given ip
func New(ip string, selfNodeId types.NodeId) Gossiper {
	g := new(proto.GossiperImpl)
	g.Init(ip, selfNodeId)
	return g
}
