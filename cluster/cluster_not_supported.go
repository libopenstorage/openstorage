package cluster

import (
	time "time"

	api "github.com/libopenstorage/openstorage/api"
	osdconfig "github.com/libopenstorage/openstorage/osdconfig"
	schedpolicy "github.com/libopenstorage/openstorage/schedpolicy"
)

// ClusterNotSupported is a NULL implementation of the Cluster interface
// It is primarily used for testing the ClusterManager as well as the
// ClusterListener interface
type ClusterNotSupported struct {
}

// AddEventListener
func (m *ClusterNotSupported) AddEventListener(arg0 ClusterListener) error {
	return nil
}

// ClearAlert
func (m *ClusterNotSupported) ClearAlert(arg0 api.ResourceType, arg1 int64) error {
	return nil
}

// CreatePair
func (m *ClusterNotSupported) CreatePair(arg0 *api.ClusterPairCreateRequest) (*api.ClusterPairCreateResponse, error) {
	return nil, nil
}

// DeleteNodeConf
func (m *ClusterNotSupported) DeleteNodeConf(arg0 string) error {
	return nil
}

// DeletePair
func (m *ClusterNotSupported) DeletePair(arg0 string) error {
	return nil
}

// DisableUpdates
func (m *ClusterNotSupported) DisableUpdates() error {
	return nil
}

// EnableUpdates
func (m *ClusterNotSupported) EnableUpdates() error {
	return nil
}

// Enumerate
func (m *ClusterNotSupported) Enumerate() (api.Cluster, error) {
	return api.Cluster{}, nil
}

// EnumerateAlerts
func (m *ClusterNotSupported) EnumerateAlerts(arg0, arg1 time.Time, arg2 api.ResourceType) (*api.Alerts, error) {
	return nil, nil
}

// EnumerateNodeConf
func (m *ClusterNotSupported) EnumerateNodeConf() (*osdconfig.NodesConfig, error) {
	return nil, nil
}

// EnumeratePairs
func (m *ClusterNotSupported) EnumeratePairs() (*api.ClusterPairsEnumerateResponse, error) {
	return nil, nil
}

// EraseAlert
func (m *ClusterNotSupported) EraseAlert(arg0 api.ResourceType, arg1 int64) error {
	return nil
}

// GetClusterConf
func (m *ClusterNotSupported) GetClusterConf() (*osdconfig.ClusterConfig, error) {
	return nil, nil
}

// GetData
func (m *ClusterNotSupported) GetData() (map[string]*api.Node, error) {
	return nil, nil
}

// GetGossipState
func (m *ClusterNotSupported) GetGossipState() *ClusterState {
	return nil
}

// GetNodeConf
func (m *ClusterNotSupported) GetNodeConf(arg0 string) (*osdconfig.NodeConfig, error) {
	return nil, nil
}

// GetNodeIdFromIp
func (m *ClusterNotSupported) GetNodeIdFromIp(arg0 string) (string, error) {
	return "", nil
}

// GetPair
func (m *ClusterNotSupported) GetPair(arg0 string) (*api.ClusterPairGetResponse, error) {
	return nil, nil
}

// GetPairToken
func (m *ClusterNotSupported) GetPairToken(arg0 bool) (*api.ClusterPairTokenGetResponse, error) {
	return nil, nil
}

// Inspect
func (m *ClusterNotSupported) Inspect(arg0 string) (api.Node, error) {
	return api.Node{}, nil
}

// NodeRemoveDone
func (m *ClusterNotSupported) NodeRemoveDone(arg0 string, arg1 error) {
	return
}

// NodeStatus
func (m *ClusterNotSupported) NodeStatus() (api.Status, error) {
	return api.Status_STATUS_NONE, nil
}

// ObjectStoreCreate
func (m *ClusterNotSupported) ObjectStoreCreate(arg0 string) (*api.ObjectstoreInfo, error) {
	return nil, nil
}

// ObjectStoreDelete
func (m *ClusterNotSupported) ObjectStoreDelete(arg0 string) error {
	return nil
}

// ObjectStoreInspect
func (m *ClusterNotSupported) ObjectStoreInspect(arg0 string) (*api.ObjectstoreInfo, error) {
	return nil, nil
}

// ObjectStoreUpdate
func (m *ClusterNotSupported) ObjectStoreUpdate(arg0 string, arg1 bool) error {
	return nil
}

// PeerStatus
func (m *ClusterNotSupported) PeerStatus(arg0 string) (map[string]api.Status, error) {
	return nil, nil
}

// ProcessPairRequest
func (m *ClusterNotSupported) ProcessPairRequest(arg0 *api.ClusterPairProcessRequest) (*api.ClusterPairProcessResponse, error) {
	return nil, nil
}

// Remove
func (m *ClusterNotSupported) Remove(arg0 []api.Node, arg1 bool) error {
	return nil
}

// SchedPolicyCreate
func (m *ClusterNotSupported) SchedPolicyCreate(arg0, arg1 string) error {
	return nil
}

// SchedPolicyDelete
func (m *ClusterNotSupported) SchedPolicyDelete(arg0 string) error {
	return nil
}

// SchedPolicyEnumerate
func (m *ClusterNotSupported) SchedPolicyEnumerate() ([]*schedpolicy.SchedPolicy, error) {
	return nil, nil
}

// SchedPolicyGet
func (m *ClusterNotSupported) SchedPolicyGet(arg0 string) (*schedpolicy.SchedPolicy, error) {
	return nil, nil
}

// SchedPolicyUpdate
func (m *ClusterNotSupported) SchedPolicyUpdate(arg0, arg1 string) error {
	return nil
}

// SecretCheckLogin
func (m *ClusterNotSupported) SecretCheckLogin() error {
	return nil
}

// SecretGet
func (m *ClusterNotSupported) SecretGet(arg0 string) (interface{}, error) {
	return nil, nil
}

// SecretGetDefaultSecretKey
func (m *ClusterNotSupported) SecretGetDefaultSecretKey() (interface{}, error) {
	return nil, nil
}

// SecretLogin
func (m *ClusterNotSupported) SecretLogin(arg0 string, arg1 map[string]string) error {
	return nil
}

// SecretSet
func (m *ClusterNotSupported) SecretSet(arg0 string, arg1 interface{}) error {
	return nil
}

// SecretSetDefaultSecretKey
func (m *ClusterNotSupported) SecretSetDefaultSecretKey(arg0 string, arg1 bool) error {
	return nil
}

// SetClusterConf
func (m *ClusterNotSupported) SetClusterConf(arg0 *osdconfig.ClusterConfig) error {
	return nil
}

// SetNodeConf
func (m *ClusterNotSupported) SetNodeConf(arg0 *osdconfig.NodeConfig) error {
	return nil
}

// SetSize
func (m *ClusterNotSupported) SetSize(arg0 int) error {
	return nil
}

// Shutdown
func (m *ClusterNotSupported) Shutdown() error {
	return nil
}

// Start
func (m *ClusterNotSupported) Start(arg0 int, arg1 bool, arg2 string) error {
	return nil
}

// StartWithConfiguration
func (m *ClusterNotSupported) StartWithConfiguration(arg0 int, arg1 bool, arg2 string, arg3 *ClusterServerConfiguration) error {
	return nil
}

// UpdateData
func (m *ClusterNotSupported) UpdateData(arg0 map[string]interface{}) error {
	return nil
}

// UpdateLabels
func (m *ClusterNotSupported) UpdateLabels(arg0 map[string]string) error {
	return nil
}
