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

	// GetSelfStatus returns the node's status
	GetSelfStatus() types.NodeStatus

	// GetStoreValue returns the StoreValue associated with
	// the given key
	GetStoreKeyValue(key types.StoreKey) types.NodeValueMap

	// GetStoreKeys returns all the keys present in the store
	GetStoreKeys() []types.StoreKey

	// Used for gossiping

	// Update updates the current state of the gossip data
	// with the newly available data
	Update(newData types.NodeInfoMap)

	// UpdateSelfStatus
	UpdateSelfStatus(types.NodeStatus)

	// UpdateNodeStatus
	UpdateNodeStatus(types.NodeId, types.NodeStatus) error

	// MetaInfo returns meta information for the
	// current available data
	MetaInfo() types.NodeMetaInfo

	// GetLocalState returns our nodeInfoMap
	GetLocalState() types.NodeInfoMap

	// GetLocalNodeInfo returns
	GetLocalNodeInfo(types.NodeId) (types.NodeInfo, error)

	// Add a new node in the database
	AddNode(types.NodeId, types.NodeStatus)

	// Remove a node from the database
	RemoveNode(types.NodeId) error
}

type Gossiper interface {
	// Gossiper has a gossip store
	GossipStore

	// Start begins the gossip protocol using memberlist
	// To join an existing cluster provide atleast one ip of the known node.
	Start(knownIp []string) error

	// GossipInterval gets the gossip interval
	GossipInterval() time.Duration

	// Stop stops the gossiping. Leave timeout indicates the minimum time
	// required to successfully broadcast the leave message to all other nodes.
	Stop(leaveTimeout time.Duration) error

	// GetNodes returns a list of the connection addresses
	GetNodes() []string

	// GetGossipHistory returns the gossip records for last 20 sessions.
	GetGossipHistory() []*types.GossipSessionInfo

	// UpdateCluster updates gossip with latest peer nodes Id-Ip mapping
	UpdateCluster(map[types.NodeId]string)

	// ExternalNodeLeave is used to indicate gossip that one of the nodes might be down.
	// It checks quorum and appropriately marks either self down or the other node down.
	// It returns the nodeId that was marked down
	ExternalNodeLeave(nodeId types.NodeId) types.NodeId
}

// New returns an initialized Gossip node
// which identifies itself with the given ip
func New(ip string, selfNodeId types.NodeId, genNumber uint64, gossipIntervals types.GossipIntervals, gossipVersion string) Gossiper {
	g := new(proto.GossiperImpl)
	g.Init(ip, selfNodeId, genNumber, gossipIntervals, gossipVersion)
	return g
}
