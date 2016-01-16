package cluster

import (
	"container/list"
	"errors"

	"github.com/fsouza/go-dockerclient"
	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/kvdb"
)

var (
	inst *ClusterManager

	errClusterInitialized = errors.New("openstorage.cluster: already initialized")
	errClusterNotInitialized = errors.New("openstorage.cluster: not initialized")
)

type Config struct {
	ClusterId string
	NodeId    string
	MgtIface  string
	DataIface string
}

// NodeEntry is used to discover other nodes in the cluster
// and setup the gossip protocol with them.
type NodeEntry struct {
	Id        string
	Ip        string
	GenNumber uint64
}

type Database struct {
	Status      api.Status
	Id          string
	NodeEntries map[string]NodeEntry
}

type GossipStatus struct {
	History    []*types.GossipSessionInfo
	NodeStatus []types.NodeValue
}

// ClusterListener is an interface to be implemented by a storage driver
// if it is participating in a multi host environment.  It exposes events
// in the cluster state machine.  Your driver can do the needful when
// these events are provided.
type ClusterListener interface {
	// String returns a string representation of this listener.
	String() string

	// ClusterInit is called when a brand new cluster is initialized.
	ClusterInit(self *api.Node, db *Database) error

	// Init is called when this node is joining an existing cluster for the first time.
	Init(self *api.Node, db *Database) error

	// CleanupInit is called when Init failed.
	CleanupInit(self *api.Node, db *Database) error

	// Join is called when this node is joining an existing cluster.
	Join(self *api.Node, db *Database) error

	// Add is called when a new node joins the cluster.
	Add(node *api.Node) error

	// Remove is called when a node leaves the cluster
	Remove(node *api.Node) error

	// Update is called when a node status changes significantly
	// in the cluster changes.
	Update(node *api.Node) error

	// Leave is called when this node leaves the cluster.
	Leave(node *api.Node) error
}

// Cluster is the API that a cluster provider will implement.
type Cluster interface {
	// LocateNode find the node given a UUID.
	LocateNode(string) (api.Node, error)

	// AddEventListener adds an event listener and exposes cluster events.
	AddEventListener(ClusterListener) error

	// Enumerate lists all the nodes in the cluster.
	Enumerate() (api.Cluster, error)

	// Remove node(s) from the cluster permanently.
	Remove(nodes []api.Node) error

	// Shutdown node(s) or the entire cluster.
	Shutdown(cluster bool, nodes []api.Node) error

	// Start starts the cluster manager and state machine.
	// It also causes this node to join the cluster.
	Start() error

	// Update Node data associated with this node
	UpdateNodeData(dataKey string, value interface{})

	// Get Node data associated with all nodes. Key is
	// the node id.
	GetClusterNodeData() map[string]*api.Node

	// Enables notifications from gossip
	EnableGossipUpdates()

	// Disable notifications from gossip
	DisableGossipUpdates()

	// Status of nodes according to gossip
	GetGossipStatus() *GossipStatus
}

// Init instantiates a new cluster manager.
func Init(cfg Config, kv kvdb.Kvdb, dockerClient *docker.Client) error {
	if inst != nil {
		return errClusterInitialized
	}
	inst = &ClusterManager{
		listeners: list.New(),
		config:    cfg,
		kv:        kv,
		nodeCache: make(map[string]api.Node),
		docker:    dockerClient,
	}
	return nil
}

// Start will run the cluster manager daemon.
func Start() error {
	if inst == nil {
		return errClusterNotInitialized
	}
	if err := inst.start(); err != nil {
		return err
	}
	return nil
}

// Inst returns an instance of an already instantiated cluster manager.
func Inst() (*ClusterManager, error) {
	if inst == nil {
		return nil, errClusterNotInitialized
	}
	return inst, nil
}
