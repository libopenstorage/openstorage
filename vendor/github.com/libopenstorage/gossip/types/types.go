package types

import (
	"fmt"
	"time"
)

type NodeId string
type StoreKey string
type NodeStatus uint8
type StateEvent uint8
type NodeInfoMap map[NodeId]NodeInfo
type NodeValueMap map[NodeId]NodeValue
type StoreMap map[StoreKey]interface{}

const (
	NODE_STATUS_INVALID NodeStatus = iota
	NODE_STATUS_UP
	NODE_STATUS_DOWN
	NODE_STATUS_NEVER_GOSSIPED
	NODE_STATUS_NOT_IN_QUORUM
	NODE_STATUS_SUSPECT_NOT_IN_QUORUM
)

const (
	SELF_ALIVE StateEvent = iota
	NODE_ALIVE
	SELF_LEAVE
	NODE_LEAVE
	UPDATE_CLUSTER_SIZE
	TIMEOUT
)

type GossipDirection uint8

const (
	// Direction of gossip
	GD_ME_TO_PEER GossipDirection = iota
	GD_PEER_TO_ME
)

type GossipOp string

const (
	LocalPush   GossipOp = "Local Push"
	MergeRemote GossipOp = "Merge Remote"
	NotifyAlive GossipOp = "Notify Alive"
	NotifyJoin  GossipOp = "Notify Join"
	NotifyLeave GossipOp = "Notify Leave"
)

type GossipSessionInfo struct {
	Node string
	Ts   time.Time
	Dir  GossipDirection
	Err  string
	Op   GossipOp
}

type NodeMetaInfo struct {
	GossipVersion string
	Id            NodeId
	GenNumber     uint64
	LastUpdateTs  time.Time
}

type NodeInfo struct {
	Id                 NodeId
	GenNumber          uint64
	LastUpdateTs       time.Time
	WaitForGenUpdateTs time.Time
	Status             NodeStatus
	Value              StoreMap
}

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

const (
	DEFAULT_GOSSIP_INTERVAL    time.Duration = 2 * time.Second
	DEFAULT_PUSH_PULL_INTERVAL time.Duration = 2 * time.Second
	DEFAULT_PROBE_INTERVAL     time.Duration = 5 * time.Second
	DEFAULT_PROBE_TIMEOUT      time.Duration = 200 * time.Millisecond
	DEFAULT_QUORUM_TIMEOUT     time.Duration = 1 * time.Minute
	DEFAULT_GOSSIP_VERSION     string        = "v1"
)

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
	// QuorumTimout is the timeout for which a node will stay in the SUSPECT_NOT_IN_QUORUM
	// and then transition to NOT_IN_QUORUM (Not UP) if quorum is not satisfied
	QuorumTimeout time.Duration
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
