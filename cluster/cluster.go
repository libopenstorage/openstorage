package cluster

import (
	"errors"
	"time"

	"github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

const (
	StatusInit = 1 << iota
	StatusOk
	StatusOffline
	StatusError
)

var (
	inst *ClusterManager
)

type Config struct {
	ClusterId string
	NodeId    string
}

// NodeInfo describes the physical parameters of a node.
type NodeInfo struct {
	Config    Config
	Cpu       float64 // percentage.
	Memory    float64 // percentage.
	Luns      map[string]systemutils.Lun
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
	AddEventListener(ClusterListener) error
	Start() error
}

func New(cfg Config, kv kvdb.Kvdb) (*ClusterManager, error) {
	inst = &ClusterManager{config: cfg, kv: kv}

	err := inst.Start()
	if err != nil {
		inst = nil
		return nil, err
	}
	return inst, nil
}

func Inst() (*ClusterManager, error) {
	if inst == nil {
		return nil, errors.New("Cluster is not initialized.")
	}
	return inst, nil
}
