package types

import (
	"fmt"
	"time"
)

// NodeId identifies the node participating in the gossip cluster
type NodeId string

// StoreKey is the key fo the StoreMap
type StoreKey string

// NodeStatus indicates the status of the node
type NodeStatus uint8

// StateEvent is an event that triggeres a change in the state of the node
// which in turn could change the NodeStatus
type StateEvent uint8

// NodeInfoMap is a map of NodeId to NodeInfo object
type NodeInfoMap map[NodeId]NodeInfo

// NodeValueMap is a map of NodeId to NodeValue object
type NodeValueMap map[NodeId]NodeValue

// StoreMap is an opaque map which the users of gossip can use
// to transer data between nodes.
type StoreMap map[StoreKey]interface{}

// QuorumProvider identifies the algorithm used to determine
// quorum of a cluster
type QuorumProvider uint8

// ClusterDomainState identifies the state of cluster domain
type ClusterDomainState string

// ClusterDomainsActiveMap is a map of cluster domain to a boolean value
// indicating whether that domain is active or inactive
type ClusterDomainsActiveMap map[string]ClusterDomainState

// ClusterDomainsQuorumMembersMap is a map of cluster domains to the number
// of quorum members in that domain
type ClusterDomainsQuorumMembersMap map[string]int

// Constant Definitions

const (
	DEFAULT_GOSSIP_INTERVAL    time.Duration = 2 * time.Second
	DEFAULT_PUSH_PULL_INTERVAL time.Duration = 2 * time.Second
	DEFAULT_PROBE_INTERVAL     time.Duration = 5 * time.Second
	DEFAULT_PROBE_TIMEOUT      time.Duration = 200 * time.Millisecond
	DEFAULT_QUORUM_TIMEOUT     time.Duration = 1 * time.Minute
	DEFAULT_GOSSIP_VERSION     string        = "v1"
	GOSSIP_VERSION_2           string        = "v2"
)

const (
	NODE_STATUS_INVALID NodeStatus = iota
	NODE_STATUS_UP
	NODE_STATUS_DOWN
	NODE_STATUS_NEVER_GOSSIPED
	NODE_STATUS_NOT_IN_QUORUM
	NODE_STATUS_SUSPECT_NOT_IN_QUORUM
	NODE_STATUS_SUSPECT_DOWN
)

const (
	SELF_ALIVE StateEvent = iota
	NODE_ALIVE
	SELF_LEAVE
	NODE_LEAVE
	UPDATE_CLUSTER_SIZE
	TIMEOUT
	UPDATE_CLUSTER_DOMAINS_ACTIVE_MAP
)

const (
	QUORUM_PROVIDER_DEFAULT QuorumProvider = iota
	QUORUM_PROVIDER_FAILURE_DOMAINS
)

const (
	CLUSTER_DOMAIN_STATE_ACTIVE   = ClusterDomainState("Active")
	CLUSTER_DOMAIN_STATE_INACTIVE = ClusterDomainState("Inactive")
)

// NodeUpdate object is used for externally updating a node in gossip
type NodeUpdate struct {
	// Addr is the contact address for the node
	Addr string
	// QuorumMember is true if node participates in quorum decisions
	QuorumMember bool
	// ClusterDomain of the node
	ClusterDomain string
}

// NodeMetaInfo object is the node metadata information that gets stored in
// memberlist's Node object. Any node update that gossip receives from memberlist
// will have this metadata.
type NodeMetaInfo struct {
	// ClusterId of the gossip cluster
	ClusterId string
	// GossipVersion is the version of gossip protocol
	GossipVersion string
	// Id is the node id
	Id NodeId
	// GenNumber of the object
	GenNumber uint64
	// LastUpdateTs is the last updated timestamp for this object
	LastUpdateTs time.Time
}

// NodeInfo is the node object that is stored for each node
// in gossip's in memory datastructures
type NodeInfo struct {
	// Id of the node
	Id NodeId
	// GenNumber of the object
	GenNumber uint64
	// LastUpdateTs is the last updated timestamp for this object
	LastUpdateTs time.Time
	// WaitForGenUpdateTs
	WaitForGenUpdateTs time.Time
	// Status of the node as seen by gossip on this node
	Status NodeStatus
	// Value is the opaque key value map provided by the callers of gossip
	Value StoreMap
	// QuorumMember indicates if this node participates in quorum calculations
	QuorumMember bool
	// ClusterDomain indicates the cluster domain in which this node lies
	ClusterDomain string
	// Addr is the connection address for this node
	Addr string
}

// NodeValue is the node object that is returned to the callers of gossip.
// It essentially is a subset of the NodeInfo object
type NodeValue struct {
	Id           NodeId
	GenNumber    uint64
	LastUpdateTs time.Time
	Status       NodeStatus
	Value        interface{}
}

func (n NodeInfo) String() string {
	return fmt.Sprintf("\nId: %v\nLastUpdateTs: %v\nStatus: : %v\nValue: %v",
		n.Id, n.LastUpdateTs, n.Status, n.Value)
}

// GossipIntervals object defines the different tuning parameters for gossip intervals and timeouts
type GossipIntervals struct {
	// GossipInterval is the time interval within which the nodes gossip
	GossipInterval time.Duration
	// PushPullInterval is the time interval for full local state tcp sync amongst nodes
	PushPullInterval time.Duration
	// ProbeInterval is the time interval for probing other nodes.
	// Used for failure detection amongst peers and reap dead nodes.
	// It is also the interval for broadcasts (Broadcasts Not used currently)
	ProbeInterval time.Duration
	// ProbeTimeout used to determine if a node is down. Should be atleast twice the RTT of network
	ProbeTimeout time.Duration
	// QuorumTimeout is the timeout for which a node will stay in the SUSPECT_NOT_IN_QUORUM
	// and then transition to NOT_IN_QUORUM (Not UP) if quorum is not satisfied
	QuorumTimeout time.Duration
}

// GossipNodeConfiguration is the peer node configuration with which gossip on this
// node can start
type GossipNodeConfiguration struct {
	// KnownUrl is the ip of this peer node
	KnownUrl string
	// ClusterDomain is the failure domain of this peer node
	ClusterDomain string
}

// GossipStartConfiguration object provides the configuration with which gossip should start.
type GossipStartConfiguration struct {
	// Nodes is a map of known nodes and their failure domains
	Nodes map[NodeId]GossipNodeConfiguration
	// ActiveMap is a map of failure domains to a boolean indicating whether they
	// are active or inactive
	ActiveMap ClusterDomainsActiveMap
	// QuorumProviderType indicates which quorum calculation algorithm to use
	QuorumProviderType QuorumProvider
}

// Used by the Gossip protocol
type StoreMetaInfo map[NodeId]NodeMetaInfo
type StoreNodes []NodeId

// OnMessageRcv is a handler that is invoked when
// message arrives on the message channel.
type OnMessageRcv func(peerid string, c MessageChannel)

// MessageChanne defines an interface for sending and
// receiving messages between peer nodes. It abstracts
// the underlying mechanism used to exchange messages.
type MessageChannel interface {
	// SendData serialized the the message and sends it
	// to peer. The data must implement json.Marshal
	SendData(obj interface{}) error
	// RcvData recieves data from the peer and unmarshals
	// it into the given obj. obj must be a pointer to
	// effect change and must implement json.Unmarshal
	RcvData(obj interface{}) error
	// RunOnRcvData loops in continously and runs a handler
	// which is activated on receiving any data
	RunOnRcvData(time.Duration)
	// Close terminates the message channel.
	Close()
}
