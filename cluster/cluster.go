//go:generate mockgen -package=mock -destination=mock/cluster.mock.go github.com/libopenstorage/openstorage/cluster Cluster
package cluster

import (
	"errors"
	"time"

	"github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/objectstore"
	"github.com/libopenstorage/openstorage/osdconfig"
	sched "github.com/libopenstorage/openstorage/schedpolicy"
	"github.com/libopenstorage/openstorage/secrets"
	"github.com/portworx/kvdb"
)

const (
	// APIVersion for cluster APIs
	APIVersion = "v1"
	// APIBase url for cluster APIs
	APIBase = "/var/lib/osd/cluster/"
)

var (
	// ErrNodeRemovePending is returned when Node remove does not succeed and is
	// kept in pending state
	ErrNodeRemovePending = errors.New("Node remove is pending")
	ErrInitNodeNotFound  = errors.New("This node is already initialized but " +
		"could not be found in the cluster map.")
	ErrNodeDecommissioned   = errors.New("Node is decomissioned.")
	ErrRemoveCausesDataLoss = errors.New("Cannot remove node without data loss")
	ErrNotImplemented       = errors.New("Not Implemented")
)

// ClusterServerConfiguration holds manager implementation
// Caller has to create the manager and passes it in
type ClusterServerConfiguration struct {
	// holds implementation to Secrets interface
	ConfigSecretManager secrets.Secrets
	// holds implementeation to SchedulePolicy interface
	ConfigSchedManager sched.SchedulePolicyProvider
	// holds implementation to ObjectStore interface
	ConfigObjectStoreManager objectstore.ObjectStore
}

// NodeEntry is used to discover other nodes in the cluster
// and setup the gossip protocol with them.
type NodeEntry struct {
	Id                string
	SchedulerNodeName string
	MgmtIp            string
	DataIp            string
	GenNumber         uint64
	StartTime         time.Time
	MemTotal          uint64
	Hostname          string
	Status            api.Status
	NodeLabels        map[string]string
	NonQuorumMember   bool
}

// ClusterInfo is the basic info about the cluster and its nodes
type ClusterInfo struct {
	Size        int
	Status      api.Status
	Id          string
	NodeEntries map[string]NodeEntry
	PairToken   string
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
// in the cluster state machine.
// The basic set of APIs determine the lifecycle of a node and comprises of two operations
// 1. Setup
// ClusterInit -> (Node)Init -> Join -> JoinComplete
// 2. Teardown
// Halt -> CleanupInit
// The other APIs are helpers for cluster manager.
type ClusterListener interface {
	// String returns a string representation of this listener.
	String() string

	// ClusterInit is called when a brand new cluster is initialized.
	ClusterInit(self *api.Node) error

	// Init is called when this node is joining an existing cluster for the first time.
	Init(self *api.Node, state *ClusterInfo) (FinalizeInitCb, error)

	// Join is called when this node is joining an existing cluster.
	Join(self *api.Node, state *ClusterInitState, clusterNotify ClusterNotify) error

	// JoinComplete is called when this node has successfully joined a cluster
	JoinComplete(self *api.Node) error

	// CleanupInit is called when Init failed.
	CleanupInit(self *api.Node, clusterInfo *ClusterInfo) error

	// Halt is called when a node is gracefully shutting down.
	Halt(self *api.Node, clusterInfo *ClusterInfo) error

	ClusterListenerNodeOps
	ClusterListenerStatusOps
	ClusterListenerGenericOps
	ClusterListenerAlertOps
	ClusterListenerPairOps
}

// ClusterListenerPairOps is an interface that must be implemented to support
// pairing of multiple clusters. It will be used at the destination cluster to
// listen for incoming pairing requests.
type ClusterListenerPairOps interface {
	// CreatePair is called when we are pairing with another cluster
	CreatePair(response *api.ClusterPairProcessResponse) error

	// ProcessPairRequest is called when we get a pair request from another cluster
	ProcessPairRequest(request *api.ClusterPairProcessRequest, response *api.ClusterPairProcessResponse) error

	// ValidatePair is called when we get a validate pair request
	ValidatePair(pair *api.ClusterPairInfo) error
}

// ClusterListenerAlertOps is a wrapper over ClusterAlerts interface
// which the listeners need to implement if they want to handle alerts
type ClusterListenerAlertOps interface {
	ClusterAlerts
}

// ClusterListenerGenericOps defines a set of generic helper APIs for
// listeners to implement
type ClusterListenerGenericOps interface {
	// ListenerData returns the data that the listener wants to share
	// with ClusterManager and would be stored in NodeData field.
	ListenerData() map[string]interface{}

	// QuorumMember returns true if the listener wants this node to
	// participate in quorum decisions.
	QuorumMember(node *api.Node) bool

	// UpdateClusterInfo is called when there is an update to the cluster.
	// XXX: Remove ClusterInfo from this API
	UpdateCluster(self *api.Node, clusterInfo *ClusterInfo) error

	// Enumerate updates listener specific data in Enumerate.
	Enumerate(cluster api.Cluster) error
}

// ClusterListenerStatusOps defines APIs that a listener needs to implement
// to indicate its own/peer statuses
type ClusterListenerStatusOps interface {
	// ListenerStatus returns the listener's Status
	ListenerStatus() api.Status

	// ListenerPeerStatus returns the peer Statuses for a listener
	ListenerPeerStatus() map[string]api.Status
}

// ClusterListenerNodeOps defines APIs that a listener needs to implement
// to handle various node operations/updates
type ClusterListenerNodeOps interface {
	// Add is called when a new node joins the cluster.
	Add(node *api.Node) error

	// Remove is called when a node leaves the cluster
	Remove(node *api.Node, forceRemove bool) error

	// CanNodeRemove test to see if we can remove this node
	CanNodeRemove(node *api.Node) (string, error)

	// MarkNodeDown marks the given node's status as down
	MarkNodeDown(node *api.Node) error

	// Update is called when a node status changes significantly
	// in the cluster changes.
	Update(node *api.Node) error

	// Leave is called when this node leaves the cluster.
	Leave(node *api.Node) error
}

// ClusterState is the gossip state of all nodes in the cluster
type ClusterState struct {
	NodeStatus []types.NodeValue
}

// ClusterData interface provides apis to handle data of the cluster
type ClusterData interface {
	// UpdateData updates node data associated with this node
	UpdateData(nodeData map[string]interface{}) error

	// UpdateLabels updates node labels associated with this node
	UpdateLabels(nodeLabels map[string]string) error

	// UpdateSchedulerNodeName updates the scheduler node name
	// associated with this node
	UpdateSchedulerNodeName(name string) error

	// GetData get sdata associated with all nodes.
	// Key is the node id
	GetData() (map[string]*api.Node, error)

	// GetNodeIdFromIp returns a Node Id given an IP.
	GetNodeIdFromIp(idIp string) (string, error)

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
	NodeStatus() (api.Status, error)

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

type ClusterAlerts interface {
	// Enumerate enumerates alerts on this cluster for the given resource
	// within a specific time range.
	EnumerateAlerts(timeStart, timeEnd time.Time, resource api.ResourceType) (*api.Alerts, error)
	// EraseAlert erases an alert for the given resource
	EraseAlert(resource api.ResourceType, alertID int64) error
}

type ClusterPair interface {
	// PairCreate with a remote cluster
	CreatePair(*api.ClusterPairCreateRequest) (*api.ClusterPairCreateResponse, error)

	// PairProcess handles an incoming pair request from a remote cluster
	ProcessPairRequest(*api.ClusterPairProcessRequest) (*api.ClusterPairProcessResponse, error)

	// GetPair returns pair information for a cluster
	GetPair(string) (*api.ClusterPairGetResponse, error)

	// EnumeratePairs returns list of cluster pairs
	EnumeratePairs() (*api.ClusterPairsEnumerateResponse, error)

	// RefreshPair Refreshes a cluster pairing by fetching latest information
	// from the remote cluster
	RefreshPair(string) error

	// DeletePair Delete a cluster pairing
	DeletePair(string) error

	// ValidatePair validates a cluster pair
	ValidatePair(string) error

	// GetPairToken gets the authentication token for this cluster
	GetPairToken(bool) (*api.ClusterPairTokenGetResponse, error)
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
	// All managers will default returning NotSupported.
	Start(clusterSize int, nodeInitialized bool, gossipPort string) error

	// Like Start, but have the ability to pass in managers to the cluster object
	StartWithConfiguration(clusterMaxSize int, nodeInitialized bool, gossipPort string, config *ClusterServerConfiguration) error

	// Get a unique identifier for this cluster. Depending on the implementation, this could
	// be different than the _id_ from ClusterInfo. This id _must_ be unique across
	// any cluster.
	Uuid() string

	ClusterData
	ClusterRemove
	ClusterStatus
	ClusterAlerts
	ClusterPair
	osdconfig.ConfigCaller
	secrets.Secrets
	sched.SchedulePolicyProvider
	objectstore.ObjectStore
}

// ClusterNotify is the callback function listeners can use to notify cluster manager
type ClusterNotify func(string, api.ClusterNotify) (string, error)

// NullClusterListener is a NULL implementation of ClusterListener functions
// ClusterListeners should use this as the base override functions they
// are interested in.
type NullClusterListener struct {
}

func (nc *NullClusterListener) String() string {
	return "NullClusterListener"
}
func (nc *NullClusterListener) ClusterInit(self *api.Node) error {
	return nil
}

func (nc *NullClusterListener) Init(self *api.Node, state *ClusterInfo) (FinalizeInitCb, error) {
	return nil, nil
}

func (nc *NullClusterListener) CleanupInit(
	self *api.Node,
	clusterInfo *ClusterInfo,
) error {
	return nil
}

func (nc *NullClusterListener) Enumerate(cluster api.Cluster) error {
	return nil
}

func (nc *NullClusterListener) Halt(
	self *api.Node,
	clusterInfo *ClusterInfo) error {
	return nil
}

func (nc *NullClusterListener) Join(
	self *api.Node,
	state *ClusterInitState,
	clusterNotify ClusterNotify,
) error {
	return nil
}

func (nc *NullClusterListener) JoinComplete(
	self *api.Node,
) error {
	return nil
}

func (nc *NullClusterListener) Add(node *api.Node) error {
	return nil
}

func (nc *NullClusterListener) Remove(node *api.Node, forceRemove bool) error {
	return nil
}

func (nc *NullClusterListener) CanNodeRemove(node *api.Node) (string, error) {
	return "", nil
}

func (nc *NullClusterListener) MarkNodeDown(node *api.Node) error {
	return nil
}

func (nc *NullClusterListener) Update(node *api.Node) error {
	return nil
}

func (nc *NullClusterListener) Leave(node *api.Node) error {
	return nil
}

func (nc *NullClusterListener) ListenerStatus() api.Status {
	return api.Status_STATUS_OK
}

func (nc *NullClusterListener) ListenerPeerStatus() map[string]api.Status {
	return nil
}

func (nc *NullClusterListener) ListenerData() map[string]interface{} {
	return nil
}

func (nc *NullClusterListener) QuorumMember(node *api.Node) bool {
	return false
}

func (nc *NullClusterListener) UpdateCluster(self *api.Node,
	clusterInfo *ClusterInfo,
) error {
	return nil
}

func (nc *NullClusterListener) EnumerateAlerts(
	timeStart, timeEnd time.Time,
	resource api.ResourceType,
) (*api.Alerts, error) {
	return nil, nil
}

func (nc *NullClusterListener) EraseAlert(
	resource api.ResourceType,
	alertID int64,
) error {
	return nil
}

func (nc *NullClusterListener) CreatePair(
	response *api.ClusterPairProcessResponse,
) error {
	return nil
}

func (nc *NullClusterListener) ProcessPairRequest(
	request *api.ClusterPairProcessRequest,
	response *api.ClusterPairProcessResponse,
) error {
	return nil
}

func (nc *NullClusterListener) ValidatePair(
	pair *api.ClusterPairInfo,
) error {
	return nil
}
