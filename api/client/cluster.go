package client

import (
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
func (c *clusterClient) String() string {
	return "ClusterManager"
}

func (c *clusterClient) Enumerate() (api.Cluster, error) {
	var cluster api.Cluster

	err := c.c.Get().Resource(clusterPath + "/enumerate").Do().Unmarshal(&cluster)
	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (c *clusterClient) LocateNode(nodeID string) (api.Node, error) {
	return api.Node{}, nil
}

func (c *clusterClient) AddEventListener(cluster.ClusterListener) error {
	return nil
}

func (c *clusterClient) UpdateNodeData(dataKey string, value interface{}) {
}

func (c *clusterClient) GetClusterNodeData() map[string]*api.Node {
	return nil
}

func (c *clusterClient) Remove(nodes []api.Node) error {
	return nil
}

func (c *clusterClient) Shutdown(cluster bool, nodes []api.Node) error {
	return nil
}

func (c *clusterClient) Start() error {
	return nil
}

func (c *clusterClient) DisableGossipUpdates() {
	c.c.Post().Resource(clusterPath + "/disablegossip").Do()
}

func (c *clusterClient) EnableGossipUpdates() {
	c.c.Post().Resource(clusterPath + "/enablegossip").Do()
}
