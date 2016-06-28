package client

import (
	"errors"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
)

const (
	clusterPath = "/cluster"
)

type clusterClient struct {
	c *Client
}

func newClusterClient(c *Client) cluster.Cluster {
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

func (c *clusterClient) Remove(nodes []api.Node) error {
	resp := api.ClusterResponse{}

	request := c.c.Delete().Resource(clusterPath + "/")

	for _, n := range nodes {
		request.QueryOption("id", n.Id)
	}

	if err := request.Do().Unmarshal(&resp); err != nil {
		return err
	}

	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}

func (c *clusterClient) Shutdown() error {
	return nil
}

func (c *clusterClient) Start() error {
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

func (c *clusterClient) GetState() (*cluster.ClusterState, error) {
	var status *cluster.ClusterState

	if err := c.c.Get().Resource(clusterPath + "/status").Do().Unmarshal(&status); err != nil {
		return nil, err
	}
	return status, nil
}
