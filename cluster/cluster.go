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

const (
	// APIVersion for cluster APIs
	APIVersion = "v1"
	// APIBase url for cluster APIs
	APIBase = "/var/lib/osd/cluster/"
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
	Status     api.Status
	NodeLabels map[string]string
}

// ClusterInfo is the basic info about the cluster and its nodes
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
	Collector   kvdb.UpdatesCollector
}

// FinalizeInitCb is invoked when init is complete and is in the process of
// updating the cluster database. This callback is invoked under lock and must
// finish quickly, else it will slow down other node joins.
type FinalizeInitCb func() error

// ClusterListener is an interface to be implemented by a storage driver
// if it is participating in a multi host environment.  It exposes events
// in the cluster state machine.  Your driver can do the needful when
// these events are provided.
type ClusterListener interface {
	// String returns a string representation of this listener.
	String() string

	// ClusterInit is called when a brand new cluster is initialized.
	ClusterInit(self *api.Node) error

	// Init is called when this node is joining an existing cluster for the first time.
	Init(self *api.Node, state *ClusterInfo) (FinalizeInitCb, error)

	// CleanupInit is called when Init failed.
	CleanupInit(self *api.Node, clusterInfo *ClusterInfo) error

	// Halt is called when a node is gracefully shutting down.
	Halt(self *api.Node, clusterInfo *ClusterInfo) error

	// Join is called when this node is joining an existing cluster.
	Join(self *api.Node, state *ClusterInitState, clusterNotify ClusterNotify) error

	// Add is called when a new node joins the cluster.
	Add(node *api.Node) error

	// Remove is called when a node leaves the cluster
	Remove(node *api.Node, forceRemove bool) error

	// CanNodeRemove test to see if we can remove this node
	CanNodeRemove(node *api.Node) error

	// MarkNodeDown marks the given node's status as down
	MarkNodeDown(node *api.Node) error

	// Update is called when a node status changes significantly
	// in the cluster changes.
	Update(node *api.Node) error

	// Leave is called when this node leaves the cluster.
	Leave(node *api.Node) error

	// ListenerStatus returns the listener's Status
	ListenerStatus() api.Status

	// ListenerPeerStatus returns the peer Statuses for a listener
	ListenerPeerStatus() map[string]api.Status

	// ListenerData returns the data that the listener wants to share
	// with ClusterManaher and would be stored in NodeData field.
	ListenerData() map[string]interface{}
}

// ClusterState is the gossip state of all nodes in the cluster
type ClusterState struct {
	History    []*types.GossipSessionInfo
	NodeStatus []types.NodeValue
}

// ClusterData interface provides apis to handle data of the cluster
type ClusterData interface {
	// UpdateData updates node data associated with this node
	UpdateData(dataKey string, value interface{}) error

	// GetData get sdata associated with all nodes.
	// Key is the node id
	GetData() (map[string]*api.Node, error)

	// EnableUpdate cluster data updates to be sent to listeners
	EnableUpdates() error

	// DisableUpdates disables cluster data updates to be sent to listeners
	DisableUpdates() error

	// GetGossipState returns the state of nodes according to gossip
	GetGossipState() *ClusterState
}

// ClusterStatus interface provides apis for cluster and node status
type ClusterStatus interface {
	// NodeStatus returns the status of THIS node as seen by the Cluster Provider
	// for a given listener. If listenerName is empty it returns the status of
	// THIS node maintained by the Cluster Provider.
	// At any time the status of the Cluster Provider takes precedence over
	// the status of listener. Precedence is determined by the severity of the status.
	NodeStatus(listenerName string) (api.Status, error)

	// PeerStatus returns the statuses of all peer nodes as seen by the
	// Cluster Provider for a given listener. If listenerName is empty is returns the
	// statuses of all peer nodes as maintained by the ClusterProvider (gossip)
	PeerStatus(listenerName string) (map[string]api.Status, error)
}

// ClusterRemove interface provides apis for removing nodes from a cluster
type ClusterRemove interface {
	// Remove node(s) from the cluster permanently.
	Remove(nodes []api.Node, forceRemove bool) error
	// NodeRemoveDone notify cluster manager NodeRemove is done.
	NodeRemoveDone(nodeID string, result error)
}

// Cluster is the API that a cluster provider will implement.
type Cluster interface {
	// Inspect the node given a UUID.
	Inspect(string) (api.Node, error)

	// AddEventListener adds an event listener and exposes cluster events.
	AddEventListener(ClusterListener) error

	// Enumerate lists all the nodes in the cluster.
	Enumerate() (api.Cluster, error)

	// SetSize sets the maximum number of nodes in a cluster.
	SetSize(size int) error

	// Shutdown can be called when THIS node is gracefully shutting down.
	Shutdown() error

	// Start starts the cluster manager and state machine.
	// It also causes this node to join the cluster.
	// nodeInitialized indicates if the caller of this method expects the node
	// to have been in an already-initialized state.
	Start(clusterSize int, nodeInitialized bool) error

	ClusterData
	ClusterRemove
	ClusterStatus
}

// ClusterNotify is the callback function listeners can use to notify cluster manager
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
