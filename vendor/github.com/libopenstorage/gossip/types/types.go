package types

import (
	"fmt"
	"time"
)

type NodeId string
type StoreKey string
type NodeStatus uint8
type NodeInfoMap map[NodeId]NodeInfo

const (
	NODE_STATUS_INVALID NodeStatus = iota
	NODE_STATUS_UP
	NODE_STATUS_DOWN
)

type NodeMetaInfo struct {
	Id           NodeId
	LastUpdateTs time.Time
}

type NodeInfo struct {
	Id           NodeId
	LastUpdateTs time.Time
	Status       NodeStatus
	Value        interface{}
}

func (n NodeInfo) String() string {
	return fmt.Sprintf("\nId: %v\nLastUpdateTs: %v\nStatus: : %v\nValue: %v",
		n.Id, n.LastUpdateTs, n.Status, n.Value)
}

type NodeInfoList struct {
	List []NodeInfo
}

type NodeMetaInfoList struct {
	List []NodeMetaInfo
}

// StoreValue is a map where the key is the
// StoreKey and the value is the NodeInfoList.
// This list gives the latest available view with this node
// for the whole system
type StoreValue map[StoreKey]NodeInfoList

// Used by the Gossip protocol
type StoreMetaInfo map[StoreKey]NodeMetaInfoList
type StoreDiff map[StoreKey]map[NodeId]NodeInfo
type StoreNodes map[StoreKey][]NodeId

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
