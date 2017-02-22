package cluster

import (
	"errors"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/cluster"
)

const (
	clusterPath = "/cluster"
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

func (c *clusterClient) Enumerate() (api.Cluster, error) {
	cluster := api.Cluster{}

	if err := c.c.Get().Resource(clusterPath + "/enumerate").Do().Unmarshal(&cluster); err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (c *clusterClient) SetSize(size int) error {
	resp := api.ClusterResponse{}

	request := c.c.Get().Resource(clusterPath + "/setsize")
	request.QueryOption("size", strconv.FormatInt(int64(size), 16))
	request.Do()
	if err := request.Do().Unmarshal(&resp); err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}

func (c *clusterClient) Inspect(nodeID string) (api.Node, error) {
	return api.Node{}, nil
}

func (c *clusterClient) AddEventListener(cluster.ClusterListener) error {
	return nil
}

func (c *clusterClient) UpdateData(dataKey string, value interface{}) error {
	return nil
}

func (c *clusterClient) GetData() (map[string]*api.Node, error) {
	return nil, nil
}

func (c *clusterClient) NodeStatus(listenerName string) (api.Status, error) {
	var resp api.Status
	request := c.c.Get().Resource(clusterPath + "/status")
	request.QueryOption("name", listenerName)
	request.Do()
	if err := request.Do().Unmarshal(&resp); err != nil {
		return api.Status_STATUS_NONE, err
	}
	return resp, nil
}

func (c *clusterClient) PeerStatus(listenerName string) (map[string]api.Status, error) {
	var resp map[string]api.Status
	request := c.c.Get().Resource(clusterPath + "/peerstatus")
	request.QueryOption("name", listenerName)
	request.Do()
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

func (c *clusterClient) Start(int, bool) error {
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
