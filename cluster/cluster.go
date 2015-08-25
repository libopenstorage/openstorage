package cluster

import (
	"errors"
	"time"

	"github.com/portworx/kvdb"
	"github.com/portworx/systemutils"
)

type Status int

const (
	StatusInit Status = 1 << iota
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
	NodeId    string
	Cpu       float64 // percentage.
	Memory    float64 // percentage.
	Luns      map[string]systemutils.Lun
	Avgload   int
	Timestamp time.Time
	Status    Status
	Ip        string
}

type Node struct {
	Ip     string
	Status Status
}

type Info struct {
	Status    Status
	ClusterId string
}

type Database struct {
	Cluster Info
	Nodes   map[string]Node
}

// ClusterListener is an interface to be implemented by a storage driver
// if it is participating in a multi host environment.  It exposes events
// in the cluster state machine.  Your driver can do the needful when
// these events are provided.
type ClusterListener interface {
	// String returns a string representation of this listener.
	String() string

	// ClusterInit is called when a brand new cluster is initialized.
	ClusterInit(self *NodeInfo, db *Database) error

	// Init is called when this node is joining an existing cluster for the first time.
	Init(self *NodeInfo, db *Database) error

	// Join is called when this node is joining an existing cluster.
	Join(self *NodeInfo, db *Database) error

	// Add is called when a new node joins the cluster.
	Add(info *NodeInfo) error

	// Remove is called when a node leaves the cluster
	Remove(info *NodeInfo) error

	// Update is called when a node status changes significantly
	// in the cluster changes.
	Update(info *NodeInfo) error

	// Leave is called when this node leaves the cluster.
	Leave(info *NodeInfo) error
}

type Cluster interface {
	AddEventListener(ClusterListener) error
	Start() error
}

// New instantiates and starts a new cluster manager.
func New(cfg Config, kv kvdb.Kvdb) (*ClusterManager, error) {
	inst = &ClusterManager{config: cfg, kv: kv}

	err := inst.Start()
	if err != nil {
		inst = nil
		return nil, err
	}
	return inst, nil
}

// Inst returns an instance of an already instantiated cluster manager.
func Inst() (*ClusterManager, error) {
	if inst == nil {
		return nil, errors.New("Cluster is not initialized.")
	}
	return inst, nil
}
