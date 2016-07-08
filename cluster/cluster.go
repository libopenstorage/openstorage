package cluster

import (
	"container/list"
	"errors"
	"time"

	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/config"
	"github.com/portworx/kvdb"
)

var (
	inst *ClusterManager

	errClusterInitialized    = errors.New("openstorage.cluster: already initialized")
	errClusterNotInitialized = errors.New("openstorage.cluster: not initialized")
)

// NodeEntry is used to discover other nodes in the cluster
// and setup the gossip protocol with them.
type NodeEntry struct {
	Id         string
	MgmtIp     string
	DataIp     string
	GenNumber  uint64
	StartTime  time.Time
	MemTotal   uint64
	Hostname   string
	NodeLabels map[string]string
}

type ClusterInfo struct {
	Size        int
	Status      api.Status
	Id          string
	NodeEntries map[string]NodeEntry
}

// ClusterInitState is the snapshot state which should be used to initialize
type ClusterInitState struct {
	ClusterInfo *ClusterInfo
	InitDb      kvdb.Kvdb
	Version     uint64
}

// ClusterListener is an interface to be implemented by a storage driver
// if it is participating in a multi host environment.  It exposes events
// in the cluster state machine.  Your driver can do the needful when
// these events are provided.
type ClusterListener interface {
	// String returns a string representation of this listener.
	String() string

	// ClusterInit is called when a brand new cluster is initialized.
	ClusterInit(self *api.Node, state *ClusterInitState) error

	// Init is called when this node is joining an existing cluster for the first time.
	Init(self *api.Node, state *ClusterInitState) error

	// CleanupInit is called when Init failed.
	CleanupInit(self *api.Node, clusterInfo *ClusterInfo) error

	// Halt is called when a node is gracefully shutting down.
	Halt(self *api.Node, clusterInfo *ClusterInfo) error

	// Join is called when this node is joining an existing cluster.
	Join(self *api.Node, state *ClusterInitState, clusterNotify ClusterNotify) error

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

type ClusterState struct {
	History    []*types.GossipSessionInfo
	NodeStatus []types.NodeValue
}

type ClusterData interface {
	// Update node data associated with this node
	UpdateData(dataKey string, value interface{}) error

	// Get data associated with all nodes.
	// Key is the node id
	GetData() (map[string]*api.Node, error)

	// Enables cluster data updates to be sent to listeners
	EnableUpdates() error

	// Disables cluster data updates to be sent to listeners
	DisableUpdates() error

	// Status of nodes according to gossip
	GetState() (*ClusterState, error)
}

// Cluster is the API that a cluster provider will implement.
type Cluster interface {
	// Inspect the node given a UUID.
	Inspect(string) (api.Node, error)

	// AddEventListener adds an event listener and exposes cluster events.
	AddEventListener(ClusterListener) error

	// Enumerate lists all the nodes in the cluster.
	Enumerate() (api.Cluster, error)

	// Remove node(s) from the cluster permanently.
	Remove(nodes []api.Node) error

	// SetSize sets the maximum number of nodes in a cluster.
	SetSize(size int) error

	// Shutdown can be called when THIS node is gracefully shutting down.
	Shutdown() error

	// Start starts the cluster manager and state machine.
	// It also causes this node to join the cluster.
	Start() error

	ClusterData
}

type ClusterNotify func(string, api.ClusterNotify) (string, error)

// Init instantiates a new cluster manager.
func Init(cfg config.ClusterConfig) error {
	if inst != nil {
		return errClusterInitialized
	}

	kv := kvdb.Instance()
	if kv == nil {
		return errors.New("KVDB is not yet initialized.  " +
			"A valid KVDB instance required for the cluster to start.")
	}

	inst = &ClusterManager{
		listeners:    list.New(),
		config:       cfg,
		kv:           kv,
		nodeCache:    make(map[string]api.Node),
		nodeStatuses: make(map[string]api.Status),
	}

	return nil
}

// Inst returns an instance of an already instantiated cluster manager.
func Inst() (Cluster, error) {
	if inst == nil {
		return nil, errClusterNotInitialized
	}
	return inst, nil
}
