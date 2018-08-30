package cluster

import (
	time "time"

	api "github.com/libopenstorage/openstorage/api"
	osdconfig "github.com/libopenstorage/openstorage/osdconfig"
	schedpolicy "github.com/libopenstorage/openstorage/schedpolicy"
)

// nullClusterMgr is a NULL implementation of the Cluster interface
// It is primarily used for testing the ClusterManager as well as the
// ClusterListener interface
type NullClusterMgr struct {
}

// AddEventListener
func (m *NullClusterMgr) AddEventListener(arg0 ClusterListener) error {
	return nil
}

// ClearAlert
func (m *NullClusterMgr) ClearAlert(arg0 api.ResourceType, arg1 int64) error {
	return ErrNotImplemented
}

// CreatePair
func (m *NullClusterMgr) CreatePair(arg0 *api.ClusterPairCreateRequest) (*api.ClusterPairCreateResponse, error) {
	return nil, ErrNotImplemented
}

// DeleteNodeConf
func (m *NullClusterMgr) DeleteNodeConf(arg0 string) error {
	return ErrNotImplemented
}

// DeletePair
func (m *NullClusterMgr) DeletePair(arg0 string) error {
	return ErrNotImplemented
}

// DisableUpdates
func (m *NullClusterMgr) DisableUpdates() error {
	return ErrNotImplemented
}

// EnableUpdates
func (m *NullClusterMgr) EnableUpdates() error {
	return ErrNotImplemented
}

// Enumerate
func (m *NullClusterMgr) Enumerate() (api.Cluster, error) {
	return api.Cluster{}, ErrNotImplemented
}

// EnumerateAlerts
func (m *NullClusterMgr) EnumerateAlerts(arg0, arg1 time.Time, arg2 api.ResourceType) (*api.Alerts, error) {
	return nil, ErrNotImplemented
}

// EnumerateNodeConf
func (m *NullClusterMgr) EnumerateNodeConf() (*osdconfig.NodesConfig, error) {
	return nil, ErrNotImplemented
}

// EnumeratePairs
func (m *NullClusterMgr) EnumeratePairs() (*api.ClusterPairsEnumerateResponse, error) {
	return nil, ErrNotImplemented
}

// EraseAlert
func (m *NullClusterMgr) EraseAlert(arg0 api.ResourceType, arg1 int64) error {
	return ErrNotImplemented
}

// GetClusterConf
func (m *NullClusterMgr) GetClusterConf() (*osdconfig.ClusterConfig, error) {
	return nil, ErrNotImplemented
}

// GetData
func (m *NullClusterMgr) GetData() (map[string]*api.Node, error) {
	return nil, ErrNotImplemented
}

// GetGossipState
func (m *NullClusterMgr) GetGossipState() *ClusterState {
	return nil
}

// GetNodeConf
func (m *NullClusterMgr) GetNodeConf(arg0 string) (*osdconfig.NodeConfig, error) {
	return nil, ErrNotImplemented
}

// GetNodeIdFromIp
func (m *NullClusterMgr) GetNodeIdFromIp(arg0 string) (string, error) {
	return "", ErrNotImplemented
}

// GetPair
func (m *NullClusterMgr) GetPair(arg0 string) (*api.ClusterPairGetResponse, error) {
	return nil, ErrNotImplemented
}

// GetPairToken
func (m *NullClusterMgr) GetPairToken(arg0 bool) (*api.ClusterPairTokenGetResponse, error) {
	return nil, ErrNotImplemented
}

// Inspect
func (m *NullClusterMgr) Inspect(arg0 string) (api.Node, error) {
	return api.Node{}, ErrNotImplemented
}

// NodeRemoveDone
func (m *NullClusterMgr) NodeRemoveDone(arg0 string, arg1 error) {
	return
}

// Nodestatus
func (m *NullClusterMgr) NodeStatus() (api.Status, error) {
	return api.Status_STATUS_NONE, ErrNotImplemented
}

// ObjectStoreCreate
func (m *NullClusterMgr) ObjectStoreCreate(arg0 string) (*api.ObjectstoreInfo, error) {
	return nil, ErrNotImplemented
}

// ObjectStoreDelete
func (m *NullClusterMgr) ObjectStoreDelete(arg0 string) error {
	return ErrNotImplemented
}

// ObjectStoreInspect
func (m *NullClusterMgr) ObjectStoreInspect(arg0 string) (*api.ObjectstoreInfo, error) {
	return nil, ErrNotImplemented
}

// ObjectStoreUpdate
func (m *NullClusterMgr) ObjectStoreUpdate(arg0 string, arg1 bool) error {
	return ErrNotImplemented
}

// PeerStatus
func (m *NullClusterMgr) PeerStatus(arg0 string) (map[string]api.Status, error) {
	return nil, ErrNotImplemented
}

// ProcessPairRequest
func (m *NullClusterMgr) ProcessPairRequest(arg0 *api.ClusterPairProcessRequest) (*api.ClusterPairProcessResponse, error) {
	return nil, ErrNotImplemented
}

// Remove
func (m *NullClusterMgr) Remove(arg0 []api.Node, arg1 bool) error {
	return ErrNotImplemented
}

// SchedPolicyCreate
func (m *NullClusterMgr) SchedPolicyCreate(arg0, arg1 string) error {
	return ErrNotImplemented
}

// SchedPolicyDelete
func (m *NullClusterMgr) SchedPolicyDelete(arg0 string) error {
	return ErrNotImplemented
}

// SchedPolicyEnumerate
func (m *NullClusterMgr) SchedPolicyEnumerate() ([]*schedpolicy.SchedPolicy, error) {
	return nil, ErrNotImplemented
}

// SchedPolicyGet
func (m *NullClusterMgr) SchedPolicyGet(arg0 string) (*schedpolicy.SchedPolicy, error) {
	return nil, ErrNotImplemented
}

// SchedPolicyUpdate
func (m *NullClusterMgr) SchedPolicyUpdate(arg0, arg1 string) error {
	return ErrNotImplemented
}

// SecretCheckLogin
func (m *NullClusterMgr) SecretCheckLogin() error {
	return ErrNotImplemented
}

// SecretGet
func (m *NullClusterMgr) SecretGet(arg0 string) (interface{}, error) {
	return nil, ErrNotImplemented
}

// SecretGetDefaultSecretKey
func (m *NullClusterMgr) SecretGetDefaultSecretKey() (interface{}, error) {
	return nil, ErrNotImplemented
}

// SecretLogin
func (m *NullClusterMgr) SecretLogin(arg0 string, arg1 map[string]string) error {
	return ErrNotImplemented
}

// SecretSet
func (m *NullClusterMgr) SecretSet(arg0 string, arg1 interface{}) error {
	return ErrNotImplemented
}

// SecretSetDefaultSecretKey
func (m *NullClusterMgr) SecretSetDefaultSecretKey(arg0 string, arg1 bool) error {
	return ErrNotImplemented
}

// SetClusterConf
func (m *NullClusterMgr) SetClusterConf(arg0 *osdconfig.ClusterConfig) error {
	return ErrNotImplemented
}

// SetNodeConf
func (m *NullClusterMgr) SetNodeConf(arg0 *osdconfig.NodeConfig) error {
	return ErrNotImplemented
}

// SetSize
func (m *NullClusterMgr) SetSize(arg0 int) error {
	return ErrNotImplemented
}

// Shutdown
func (m *NullClusterMgr) Shutdown() error {
	return ErrNotImplemented
}

// Start
func (m *NullClusterMgr) Start(arg0 int, arg1 bool, arg2 string) error {
	return ErrNotImplemented
}

// StartWithConfiguration
func (m *NullClusterMgr) StartWithConfiguration(arg0 int, arg1 bool, arg2 string, arg3 *ClusterServerConfiguration) error {
	return ErrNotImplemented
}

// UpdateData
func (m *NullClusterMgr) UpdateData(arg0 map[string]interface{}) error {
	return ErrNotImplemented
}

// UpdateLabels
func (m *NullClusterMgr) UpdateLabels(arg0 map[string]string) error {
	return ErrNotImplemented
}
