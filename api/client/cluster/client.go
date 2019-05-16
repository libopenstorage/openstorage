package cluster

import (
	"errors"
	"strconv"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/cluster"
	sched "github.com/libopenstorage/openstorage/schedpolicy"
	"github.com/libopenstorage/openstorage/secrets"
)

const (
	clusterPath      = "/cluster"
	secretPath       = "/secrets"
	SchedPath        = "/schedpolicy"
	loggingurl       = "/loggingurl"
	managementurl    = "/managementurl"
	fluentdhost      = "/fluentdconfig"
	tunnelconfigurl  = "/tunnelconfig"
	PairPath         = "/pair"
	PairValidatePath = "/validate"
	PairTokenPath    = "/pairtoken"
)

type clusterClient struct {
	c *client.Client
}

func newClusterClient(c *client.Client) cluster.Cluster {
	return &clusterClient{c: c}
}

// String description of this driver.
func (c *clusterClient) Name() string {
	return "ClusterManager"
}

func (c *clusterClient) CreatePair(
	request *api.ClusterPairCreateRequest,
) (*api.ClusterPairCreateResponse, error) {
	resp := &api.ClusterPairCreateResponse{}

	path := clusterPath + PairPath
	response := c.c.Put().Resource(path).Body(request).Do()

	if response.Error() != nil {
		return nil, response.FormatError()
	}

	if err := response.Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clusterClient) ProcessPairRequest(
	request *api.ClusterPairProcessRequest,
) (*api.ClusterPairProcessResponse, error) {
	resp := &api.ClusterPairProcessResponse{}

	path := clusterPath + PairPath
	response := c.c.Post().Resource(path).Body(request).Do()
	if response.Error() != nil {
		return nil, response.FormatError()
	}

	if err := response.Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clusterClient) ValidatePair(
	id string,
) error {
	path := clusterPath + PairPath + PairValidatePath
	response := c.c.Put().Resource(path).Instance(id).Do()

	if response.Error() != nil {
		return response.FormatError()
	}
	return nil
}

func (c *clusterClient) DeletePair(
	pairId string,
) error {

	path := clusterPath + PairPath
	response := c.c.Delete().Resource(path).Instance(pairId).Do()

	if response.Error() != nil {
		return response.FormatError()
	}
	return nil
}

func (c *clusterClient) GetPair(
	id string,
) (*api.ClusterPairGetResponse, error) {
	resp := &api.ClusterPairGetResponse{}
	path := clusterPath + PairPath
	response := c.c.Get().Resource(path).Instance(id).Do()

	if response.Error() != nil {
		return nil, response.FormatError()
	}
	if err := response.Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clusterClient) RefreshPair(
	pairId string,
) error {

	path := clusterPath + PairPath
	response := c.c.Put().Resource(path).Instance(pairId).Do()

	if response.Error() != nil {
		return response.FormatError()
	}
	return nil
}

func (c *clusterClient) EnumeratePairs() (*api.ClusterPairsEnumerateResponse, error) {
	resp := &api.ClusterPairsEnumerateResponse{}
	path := clusterPath + PairPath
	response := c.c.Get().Resource(path).Do()

	if response.Error() != nil {
		return nil, response.FormatError()
	}
	if err := response.Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clusterClient) GetPairToken(
	resetToken bool,
) (*api.ClusterPairTokenGetResponse, error) {
	resp := &api.ClusterPairTokenGetResponse{}

	path := clusterPath + PairTokenPath
	response := c.c.Get().Resource(path).QueryOption("reset", strconv.FormatBool(resetToken)).Do()
	if response.Error() != nil {
		return nil, response.FormatError()
	}

	if err := response.Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Enumerate returns information about the cluster and its nodes
func (c *clusterClient) Enumerate() (api.Cluster, error) {
	clusterInfo := api.Cluster{}

	if err := c.c.Get().Resource(clusterPath + "/enumerate").Do().Unmarshal(&clusterInfo); err != nil {
		return clusterInfo, err
	}
	return clusterInfo, nil
}

func (c *clusterClient) SetSize(size int) error {
	resp := api.ClusterResponse{}

	request := c.c.Get().Resource(clusterPath + "/setsize")
	request.QueryOption("size", strconv.FormatInt(int64(size), 16))
	if err := request.Do().Unmarshal(&resp); err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}

func (c *clusterClient) Inspect(nodeID string) (api.Node, error) {
	var resp api.Node
	request := c.c.Get().Resource(clusterPath + "/inspect/" + nodeID)
	if err := request.Do().Unmarshal(&resp); err != nil {
		return api.Node{}, err
	}
	return resp, nil
}

func (c *clusterClient) AddEventListener(cluster.ClusterListener) error {
	return nil
}

func (c *clusterClient) UpdateData(nodeData map[string]interface{}) error {
	return nil
}

func (c *clusterClient) UpdateLabels(nodeLabels map[string]string) error {
	return nil
}

func (c *clusterClient) UpdateSchedulerNodeName(name string) error {
	return nil
}

func (c *clusterClient) GetData() (map[string]*api.Node, error) {
	return nil, nil
}

func (c *clusterClient) GetNodeIdFromIp(idIp string) (string, error) {
	var resp string
	request := c.c.Get().Resource(clusterPath + "/getnodeidfromip/" + idIp)
	if err := request.Do().Unmarshal(&resp); err != nil {
		return idIp, err
	}
	return resp, nil
}

func (c *clusterClient) NodeStatus() (api.Status, error) {
	var resp api.Status
	request := c.c.Get().Resource(clusterPath + "/nodestatus")
	if err := request.Do().Unmarshal(&resp); err != nil {
		return api.Status_STATUS_NONE, err
	}
	return resp, nil
}

func (c *clusterClient) PeerStatus(listenerName string) (map[string]api.Status, error) {
	var resp map[string]api.Status
	request := c.c.Get().Resource(clusterPath + "/peerstatus")
	request.QueryOption("name", listenerName)
	if err := request.Do().Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *clusterClient) Remove(nodes []api.Node, forceRemove bool) error {
	resp := api.ClusterResponse{}

	request := c.c.Delete().Resource(clusterPath + "/")

	for _, n := range nodes {
		request.QueryOption("id", n.Id)
	}
	request.QueryOption("forceRemove", strconv.FormatBool(forceRemove))

	if err := request.Do().Unmarshal(&resp); err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}

func (c *clusterClient) NodeRemoveDone(nodeID string, result error) {
}

func (c *clusterClient) Shutdown() error {
	return nil
}

func (c *clusterClient) Start(int, bool, string) error {
	return nil
}

func (c *clusterClient) Uuid() string {
	return ""
}

func (c *clusterClient) StartWithConfiguration(int, bool, string, *cluster.ClusterServerConfiguration) error {
	return nil
}

func (c *clusterClient) DisableUpdates() error {
	c.c.Put().Resource(clusterPath + "/disablegossip").Do()
	return nil
}

func (c *clusterClient) EnableUpdates() error {
	c.c.Put().Resource(clusterPath + "/enablegossip").Do()
	return nil
}

func (c *clusterClient) GetGossipState() *cluster.ClusterState {
	var status *cluster.ClusterState

	if err := c.c.Get().Resource(clusterPath + "/gossipstate").Do().Unmarshal(&status); err != nil {
		return nil
	}
	return status
}

func (c *clusterClient) EnumerateAlerts(ts, te time.Time, resource api.ResourceType) (*api.Alerts, error) {
	a := api.Alerts{}
	request := c.c.Get().Resource(clusterPath + "/alerts/" + strconv.FormatInt(int64(resource), 10))
	if !te.IsZero() {
		request.QueryOption("timestart", ts.Format(api.TimeLayout))
		request.QueryOption("timeend", te.Format(api.TimeLayout))
	}
	if err := request.Do().Unmarshal(&a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *clusterClient) EraseAlert(resource api.ResourceType, alertID int64) error {
	path := clusterPath + "/alerts/" + strconv.FormatInt(int64(resource), 10) + "/" + strconv.FormatInt(alertID, 10)
	request := c.c.Delete().Resource(path)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}

// SecretSetDefaultSecretKey sets the cluster wide secret key
func (c *clusterClient) SecretSetDefaultSecretKey(secretKey string, override bool) error {
	reqBody := &secrets.DefaultSecretKeyRequest{
		DefaultSecretKey: secretKey,
		Override:         override,
	}
	path := clusterPath + secretPath + "/defaultsecretkey"
	request := c.c.Put().Resource(path).Body(reqBody)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}

// SecretGetDefaultSecretKey returns cluster wide secret key's value
func (c *clusterClient) SecretGetDefaultSecretKey() (interface{}, error) {
	var defaultKeyResp interface{}
	path := clusterPath + secretPath + "/defaultsecretkey"
	request := c.c.Get().Resource(path)
	err := request.Do().Unmarshal(&defaultKeyResp)
	if err != nil {
		return defaultKeyResp, err
	}
	return defaultKeyResp, nil
}

// SecretSet the given value/data against the key
func (c *clusterClient) SecretSet(secretID string, secretValue interface{}) error {
	reqBody := &secrets.SetSecretRequest{
		SecretValue: secretValue,
	}
	path := clusterPath + secretPath
	request := c.c.Put().Resource(path).Body(reqBody)
	request.QueryOption(secrets.SecretKey, secretID)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}

// SecretGet retrieves the value/data for given key
func (c *clusterClient) SecretGet(secretID string) (interface{}, error) {
	var secResp interface{}
	path := clusterPath + secretPath
	request := c.c.Get().Resource(path)
	request.QueryOption(secrets.SecretKey, secretID)
	if err := request.Do().Unmarshal(&secResp); err != nil {
		return secResp, err
	}
	return secResp, nil
}

// SecretCheckLogin validates session with secret store
func (c *clusterClient) SecretCheckLogin() error {
	path := clusterPath + secretPath + "/verify"
	request := c.c.Get().Resource(path)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}

// SecretLogin create session with secret store
func (c *clusterClient) SecretLogin(secretType string, secretConfig map[string]string) error {
	reqBody := &secrets.SecretLoginRequest{
		SecretType:   secretType,
		SecretConfig: secretConfig,
	}
	path := clusterPath + secretPath + "/login"
	request := c.c.Post().Resource(path).Body(reqBody)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}

// SchedPolicyEnumerate enumerates all configured policies
func (c *clusterClient) SchedPolicyEnumerate() ([]*sched.SchedPolicy, error) {
	var schedPolicies []*sched.SchedPolicy
	req := c.c.Get().Resource(clusterPath + SchedPath)

	if err := req.Do().Unmarshal(&schedPolicies); err != nil {
		return nil, err
	}

	return schedPolicies, nil
}

// SchedPolicyCreate creates a policy with given name and schedule
func (c *clusterClient) SchedPolicyCreate(name, schedule string) error {
	request := &sched.SchedPolicy{
		Name:     name,
		Schedule: schedule,
	}

	req := c.c.Post().Resource(clusterPath + SchedPath).Body(request)
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

// SchedPolicyUpdate updates a policy with given name and schedule
func (c *clusterClient) SchedPolicyUpdate(name, schedule string) error {
	request := &sched.SchedPolicy{
		Name:     name,
		Schedule: schedule,
	}

	req := c.c.Put().Resource(clusterPath + SchedPath).Body(request)
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

// SchedPolicyDelete deletes a policy with given name
func (c *clusterClient) SchedPolicyDelete(name string) error {
	req := c.c.Delete().Resource(clusterPath + SchedPath).Instance(name)
	res := req.Do()

	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

// SchedPolicyGet returns schedule policy matching given name.
func (c *clusterClient) SchedPolicyGet(name string) (*sched.SchedPolicy, error) {
	policy := new(sched.SchedPolicy)
	if name == "" {
		return nil, errors.New("Missing policy name")
	}

	req := c.c.Get().Resource(clusterPath + SchedPath).Instance(name)

	if err := req.Do().Unmarshal(policy); err != nil {
		return nil, err
	}

	return policy, nil
}
