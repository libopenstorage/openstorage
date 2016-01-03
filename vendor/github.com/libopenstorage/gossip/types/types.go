package types

import (
	"fmt"
	"time"
)

type NodeId string
type StoreKey string
type NodeStatus uint8
type NodeInfoMap map[NodeId]NodeInfo
type NodeValueMap map[NodeId]NodeValue
type StoreMap map[StoreKey]interface{}

const (
	NODE_STATUS_INVALID NodeStatus = iota
	NODE_STATUS_UP
	NODE_STATUS_DOWN
	NODE_STATUS_NEVER_GOSSIPED
	NODE_STATUS_WAITING_FOR_NEW_UPDATE
	NODE_STATUS_DOWN_WAITING_FOR_NEW_UPDATE
)

type GossipDirection uint8

const (
	// Direction of gossip
	GD_ME_TO_PEER GossipDirection = iota
	GD_PEER_TO_ME
)

type GossipSessionInfo struct {
	Node string
	Ts   time.Time
	Dir  GossipDirection
	Err  string
}

type NodeMetaInfo struct {
	Id           NodeId
	GenNumber    uint64
	LastUpdateTs time.Time
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

// Used by the Gossip protocol
type StoreMetaInfo map[NodeId]NodeMetaInfo
type StoreNodes []NodeId

// OnMessageRcv is a handler that is invoked when
// message arrives on the message channel.
type OnMessageRcv func(c MessageChannel)

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
	RunOnRcvData()
	// Close terminates the message channel.
	Close()
}
