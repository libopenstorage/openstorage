package cluster

import (
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/objectstore"
)

const (
	ObjectStorePath = "/objectstore"
)

func (c *clusterClient) ObjectStoreInspect(objectStoreID string) (*api.ObjectstoreInfo, error) {
	objectstoreInfo := &api.ObjectstoreInfo{}
	request := c.c.Get().Resource(clusterPath + ObjectStorePath)
	request.QueryOption(objectstore.ObjectStoreID, objectStoreID)
	if err := request.Do().Unmarshal(objectstoreInfo); err != nil {
		return nil, err
	}
	return objectstoreInfo, nil
}

func (c *clusterClient) ObjectStoreCreate(volume string) (*api.ObjectstoreInfo, error) {
	objectstoreInfo := &api.ObjectstoreInfo{}
	req := c.c.Post().Resource(clusterPath + ObjectStorePath)
	// Since volume name can be case sensitive adding it as
	// query param instead of path variable,
	// current rest client converts /path vars into lowercase format automatically
	req.QueryOption(objectstore.VolumeName, volume)

	if err := req.Do().Unmarshal(objectstoreInfo); err != nil {
		return nil, err
	}
	return objectstoreInfo, nil
}

func (c *clusterClient) ObjectStoreUpdate(objectStoreID string, enable bool) error {
	req := c.c.Put().Resource(clusterPath + ObjectStorePath)
	req.QueryOption(objectstore.Enable, strconv.FormatBool(enable))
	req.QueryOption(objectstore.ObjectStoreID, objectStoreID)
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

func (c *clusterClient) ObjectStoreDelete(objectStoreID string) error {
	req := c.c.Delete().Resource(clusterPath + ObjectStorePath + "/delete")
	req.QueryOption(objectstore.ObjectStoreID, objectStoreID)
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}
