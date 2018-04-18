package cluster

import (
	"errors"
	"strconv"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/cluster"
)

const (
	clusterPath     = "/cluster"
	loggingurl      = "/loggingurl"
	managementurl   = "/managementurl"
	fluentdhost     = "/fluentdconfig"
	tunnelconfigurl = "/tunnelconfig"
)

// Interface check
var _ cluster.ClusterClient = &ClusterClient{}

// ClusterClient a REST of implementation of cluster.ClusterClient
type ClusterClient struct {
	c *client.Client
}

// New returns a golang client to an OpenStorage REST server
func New(c *client.Client) *ClusterClient {
	return &ClusterClient{c: c}
}

// Enumerate returns information about the cluster and its nodes
func (c *ClusterClient) Enumerate() (api.Cluster, error) {
	clusterInfo := api.Cluster{}

	if err := c.c.Get().Resource(clusterPath + "/enumerate").Do().Unmarshal(&clusterInfo); err != nil {
		return clusterInfo, err
	}
	return clusterInfo, nil
}

// SetSize sets the maximum number of nodes in a cluster.
func (c *ClusterClient) SetSize(size int) error {
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

// Inspect the node given a UUID
func (c *ClusterClient) Inspect(nodeID string) (api.Node, error) {
	var resp api.Node
	request := c.c.Get().Resource(clusterPath + "/inspect/" + nodeID)
	if err := request.Do().Unmarshal(&resp); err != nil {
		return api.Node{}, err
	}
	return resp, nil
}

// GetNodeIdFromIp returns a Node Id given an IP.
func (c *ClusterClient) GetNodeIdFromIp(idIp string) (string, error) {
	var resp string
	request := c.c.Get().Resource(clusterPath + "/getnodeidfromip/" + idIp)
	if err := request.Do().Unmarshal(&resp); err != nil {
		return idIp, err
	}
	return resp, nil
}

// NodeStatus returns the status of THIS node as seen by the Cluster Provider
// for a given listener. If listenerName is empty it returns the status of
// THIS node maintained by the Cluster Provider.
// At any time the status of the Cluster Provider takes precedence over
// the status of listener. Precedence is determined by the severity of the status.
func (c *ClusterClient) NodeStatus() (api.Status, error) {
	var resp api.Status
	request := c.c.Get().Resource(clusterPath + "/nodestatus")
	if err := request.Do().Unmarshal(&resp); err != nil {
		return api.Status_STATUS_NONE, err
	}
	return resp, nil
}

// PeerStatus returns the statuses of all peer nodes as seen by the
// Cluster Provider for a given listener. If listenerName is empty is returns the
// statuses of all peer nodes as maintained by the ClusterProvider (gossip)
func (c *ClusterClient) PeerStatus(listenerName string) (map[string]api.Status, error) {
	var resp map[string]api.Status
	request := c.c.Get().Resource(clusterPath + "/peerstatus")
	request.QueryOption("name", listenerName)
	if err := request.Do().Unmarshal(&resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Remove node(s) from the cluster permanently.
func (c *ClusterClient) Remove(nodes []api.Node, forceRemove bool) error {
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

// DisableUpdates disables cluster data updates to be sent to listeners
func (c *ClusterClient) DisableUpdates() error {
	c.c.Put().Resource(clusterPath + "/disablegossip").Do()
	return nil
}

// EnableUpdates cluster data updates to be sent to listeners
func (c *ClusterClient) EnableUpdates() error {
	c.c.Put().Resource(clusterPath + "/enablegossip").Do()
	return nil
}

// GetGossipState returns the state of nodes according to gossip
func (c *ClusterClient) GetGossipState() *cluster.ClusterState {
	var status *cluster.ClusterState

	if err := c.c.Get().Resource(clusterPath + "/gossipstate").Do().Unmarshal(&status); err != nil {
		return nil
	}
	return status
}

// EnumerateAlerts enumerates alerts on this cluster for the given resource within a specific time range.
func (c *ClusterClient) EnumerateAlerts(ts, te time.Time, resource api.ResourceType) (*api.Alerts, error) {
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

// ClearAlert clears an alert for the given resource
func (c *ClusterClient) ClearAlert(resource api.ResourceType, alertID int64) error {
	path := clusterPath + "/alerts/" + strconv.FormatInt(int64(resource), 10) + "/" + strconv.FormatInt(alertID, 10)
	request := c.c.Put().Resource(path)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}

// EraseAlert erases an alert for the given resource
func (c *ClusterClient) EraseAlert(resource api.ResourceType, alertID int64) error {
	path := clusterPath + "/alerts/" + strconv.FormatInt(int64(resource), 10) + "/" + strconv.FormatInt(alertID, 10)
	request := c.c.Delete().Resource(path)
	resp := request.Do()
	if resp.Error() != nil {
		return resp.FormatError()
	}
	return nil
}
