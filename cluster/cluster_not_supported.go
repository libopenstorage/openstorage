package cluster

import (
	time "time"

	api "github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/objectstore"
	"github.com/libopenstorage/openstorage/osdconfig"
	schedpolicy "github.com/libopenstorage/openstorage/schedpolicy"
	"github.com/libopenstorage/openstorage/secrets"
)

// NullClusterManager is a NULL implementation of the Cluster interface
// It is primarily used for testing the ClusterManager as well as the
// ClusterListener interface
type NullClusterManager struct {
	NullClusterData
	NullClusterRemove
	NullClusterStatus
	NullClusterAlerts
	NullClusterPair
	osdconfig.NullConfigCaller
	secrets.NullSecrets
	schedpolicy.NullSchedMgr
	objectstore.NullObjectStoreMgr
}

func NewDefaultClusterManager() Cluster {
	return &NullClusterManager{}
}

// NullClusterData is a NULL implementation of the ClusterData interface
type NullClusterData struct {
}

func NewDefaultClusterData() ClusterData {
	return &NullClusterData{}
}

// NullClusterRemove is a NULL implementation of the ClusterRemove interface
type NullClusterRemove struct {
}

func NewDefaultClusterRemove() ClusterRemove {
	return &NullClusterRemove{}
}

// NullClusterStatus is a NULL implementation of the ClusterStatus interface
type NullClusterStatus struct {
}

func NewDefaultClusterStatus() ClusterStatus {
	return &NullClusterStatus{}
}

// NullClusterAlerts is a NULL implementation of the ClusterAlerts interface
type NullClusterAlerts struct {
}

func NewDefaultClusterAlerts() ClusterAlerts {
	return &NullClusterAlerts{}
}

// NullClusterPair is a NULL implementation of the ClusterPair interface
type NullClusterPair struct {
}

func NewDefaultCluterPair() ClusterPair {
	return &NullClusterPair{}
}

// NullClusterManager implementations

// Inspect
func (m *NullClusterManager) Inspect(arg0 string) (api.Node, error) {
	return api.Node{}, ErrNotImplemented
}

// AddEventListener
func (m *NullClusterManager) AddEventListener(arg0 ClusterListener) error {
	return nil
}

// Enumerate
func (m *NullClusterManager) Enumerate() (api.Cluster, error) {
	return api.Cluster{}, ErrNotImplemented
}

// SetSize
func (m *NullClusterManager) SetSize(arg0 int) error {
	return ErrNotImplemented
}

// Shutdown
func (m *NullClusterManager) Shutdown() error {
	return ErrNotImplemented
}

// Start
func (m *NullClusterManager) Start(arg0 int, arg1 bool, arg2 string) error {
	return ErrNotImplemented
}

// StartWithConfiguration
func (m *NullClusterManager) StartWithConfiguration(arg0 int, arg1 bool, arg2 string, arg3 *ClusterServerConfiguration) error {
	return ErrNotImplemented
}

func (n *NullClusterManager) Uuid() string {
	return ""
}

// NullClusterData implementations

// UpdateData
func (m *NullClusterData) UpdateData(arg0 map[string]interface{}) error {
	return ErrNotImplemented
}

// UpdateLabels
func (m *NullClusterData) UpdateLabels(arg0 map[string]string) error {
	return ErrNotImplemented
}

// UpdateSchedulerNodeName
func (m *NullClusterData) UpdateSchedulerNodeName(arg0 string) error {
	return ErrNotImplemented
}

// GetData
func (m *NullClusterData) GetData() (map[string]*api.Node, error) {
	return nil, ErrNotImplemented
}

// GetNodeIdFromIp
func (m *NullClusterData) GetNodeIdFromIp(arg0 string) (string, error) {
	return "", ErrNotImplemented
}

// EnableUpdates
func (m *NullClusterData) EnableUpdates() error {
	return ErrNotImplemented
}

// DisableUpdates
func (m *NullClusterData) DisableUpdates() error {
	return ErrNotImplemented
}

// GetGossipState
func (m *NullClusterData) GetGossipState() *ClusterState {
	return nil
}

// NullClusterRemove implementations

// Remove
func (m *NullClusterRemove) Remove(arg0 []api.Node, arg1 bool) error {
	return ErrNotImplemented
}

// NodeRemoveDone
func (m *NullClusterRemove) NodeRemoveDone(arg0 string, arg1 error) {
	return
}

// NullClusterStatus implementations

// Nodestatus
func (m *NullClusterStatus) NodeStatus() (api.Status, error) {
	return api.Status_STATUS_NONE, ErrNotImplemented
}

// PeerStatus
func (m *NullClusterStatus) PeerStatus(arg0 string) (map[string]api.Status, error) {
	return nil, ErrNotImplemented
}

// NullClusterAlerts implementations

// EnumerateAlerts
func (m *NullClusterAlerts) EnumerateAlerts(arg0, arg1 time.Time, arg2 api.ResourceType) (*api.Alerts, error) {
	return nil, ErrNotImplemented
}

// EraseAlert
func (m *NullClusterAlerts) EraseAlert(arg0 api.ResourceType, arg1 int64) error {
	return ErrNotImplemented
}

// NullClusterPair implementations

// CreatePair
func (m *NullClusterPair) CreatePair(arg0 *api.ClusterPairCreateRequest) (*api.ClusterPairCreateResponse, error) {
	return nil, ErrNotImplemented
}

// ProcessPairRequest
func (m *NullClusterPair) ProcessPairRequest(arg0 *api.ClusterPairProcessRequest) (*api.ClusterPairProcessResponse, error) {
	return nil, ErrNotImplemented
}

// GetPair
func (m *NullClusterPair) GetPair(arg0 string) (*api.ClusterPairGetResponse, error) {
	return nil, ErrNotImplemented
}

// RefreshPair
func (m *NullClusterPair) RefreshPair(arg0 string) error {
	return ErrNotImplemented
}

// EnumeratePairs
func (m *NullClusterPair) EnumeratePairs() (*api.ClusterPairsEnumerateResponse, error) {
	return nil, ErrNotImplemented
}

// ValidatePair
func (m *NullClusterPair) ValidatePair(arg0 string) error {
	return ErrNotImplemented
}

// DeletePair
func (m *NullClusterPair) DeletePair(arg0 string) error {
	return ErrNotImplemented
}

// GetPairToken
func (m *NullClusterPair) GetPairToken(arg0 bool) (*api.ClusterPairTokenGetResponse, error) {
	return nil, ErrNotImplemented
}
