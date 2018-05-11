package cluster

import (
	"github.com/libopenstorage/openstorage/services"
)

// ServiceEnterMaintenanceMode takes node out of cluster
func (c *clusterClient) ServiceEnterMaintenanceMode(exitOut bool) error {
	req := c.c.Put().Resource(clusterPath + ServicePath + "/entermaintenance")
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

// ServiceExitMaintenanceMode puts node back into cluster
func (c *clusterClient) ServiceExitMaintenanceMode() error {
	req := c.c.Put().Resource(clusterPath + ServicePath + "/exitmaintenance")
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

// ServiceAddDrive adds specified drive
func (c *clusterClient) ServiceAddDrive(op, drive string, journal bool) (string, error) {
	request := &services.AddDrive{
		Operation: op,
		Drive:     drive,
		Journal:   journal,
	}
	var srvResp services.ServiceMessage
	req := c.c.Post().Resource(clusterPath + ServicePath + "/drive").Body(request)
	if err := req.Do().Unmarshal(&srvResp); err != nil {
		return srvResp.Status, err
	}

	return srvResp.Status, nil
}

// ServiceReplaceDrive replace specified old drive with new drive
func (c *clusterClient) ServiceReplaceDrive(op, source, target string) (string, error) {
	request := &services.ReplaceDrive{
		Operation: op,
		Source:    source,
		Target:    target,
	}
	var srvResp services.ServiceMessage
	req := c.c.Put().Resource(clusterPath + ServicePath + "/drive").Body(request)
	if err := req.Do().Unmarshal(&srvResp); err != nil {
		return srvResp.Status, err
	}

	return srvResp.Status, nil
}

// ServiceRebalancePool reblance the storage pool
func (c *clusterClient) ServiceRebalancePool(op string, poolID int) (string, error) {
	request := &services.RebalancePool{
		Operation: op,
		PoolID:    poolID,
	}
	var srvResp services.ServiceMessage
	req := c.c.Put().Resource(clusterPath + ServicePath + "/rebalancepool").Body(request)
	if err := req.Do().Unmarshal(&srvResp); err != nil {
		return srvResp.Status, err
	}

	return srvResp.Status, nil
}
