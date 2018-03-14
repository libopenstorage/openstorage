package cluster

import (
	"encoding/json"

	"github.com/libopenstorage/openstorage/osdconfig"
)

const (
	UriCluster   = "/config/cluster"
	UriNode      = "/config/node"
	UriEnumerate = "/config/enumerate"
)

// osdconfig.ConfigCaller interface compliance
func (c *clusterClient) GetClusterConf() (*osdconfig.ClusterConfig, error) {
	config := new(osdconfig.ClusterConfig)
	request := c.c.Get().Resource(clusterPath + UriCluster)
	if err := request.Do().Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *clusterClient) GetNodeConf(nodeID string) (*osdconfig.NodeConfig, error) {
	config := new(osdconfig.NodeConfig)
	request := c.c.Get().Resource(clusterPath + UriNode + "/" + nodeID)
	if err := request.Do().Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *clusterClient) EnumerateNodeConf() (*osdconfig.NodesConfig, error) {
	config := new(osdconfig.NodesConfig)
	request := c.c.Get().Resource(clusterPath + UriEnumerate)
	if err := request.Do().Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *clusterClient) SetClusterConf(config *osdconfig.ClusterConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	request := c.c.Post().Body(data).Resource(clusterPath + UriCluster)
	if err := request.Do().Error(); err != nil {
		return err
	}
	return nil
}

func (c *clusterClient) SetNodeConf(config *osdconfig.NodeConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	request := c.c.Post().Body(data).Resource(clusterPath + UriNode)
	if err := request.Do().Error(); err != nil {
		return err
	}
	return nil
}

func (c *clusterClient) DeleteNodeConf(nodeID string) error {
	request := c.c.Delete().Resource(clusterPath + UriNode + "/" + nodeID)
	if err := request.Do().Error(); err != nil {
		return err
	}
	return nil
}
