package cluster

import (
	"time"
)

const (
	StatusInit = 1 << iota
	StatusOk
	StatusOffline
	StatusError
)

type NodeInfo struct {
	UUID      string
	Id        int
	Cpu       int
	Memory    int
	Iops      int
	Avgload   int
	Timestamp time.Time
	Status    uint8
	Last      *NodeInfo
}

type Node struct {
	Id        int
	Timestamp time.Time
	UUID      string
	Ip        string
	Status    uint8
}

type Info struct {
	Status    uint8
	ClusterId string
}

type Database struct {
	Cluster Info
	Nodes   map[string]Node
}

// ClusterListener is an interface to be implemented by a storage driver
// if it is participating in a multi host environment.
type ClusterListener interface {
	// String returns a string representation of this listener.
	String() string

	// ClusterInit is called when a brand new cluster is initialized.
	ClusterInit(db *Database) error

	// Init is called when this node is joining an existing cluster for the first time.
	Init(node *Node, db *Database) error

	// Join is called when this node is joining an existing cluster.
	Join(node *Node, db *Database) error

	// Add is called when a new node joins the cluster.
	Add(newNode *Node, db *Database) error

	// Remove is called when a node leaves the cluster
	Remove(oldNode *Node, db *Database) error

	// Update is called when a node status changes significantly
	// in the cluster changes.
	Update(node *Node, info *NodeInfo, db *Database) error

	// Leave is called when this node leaves the cluster.
	Leave(node *Node, db *Database) error
}

type Cluster interface {
	AddEventListener(ClusterListener)
	Connect() error
	Start()
}
