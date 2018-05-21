package cluster

import (
	"github.com/libopenstorage/openstorage/objectstore"
	"strconv"
)

const (
	ObjectStorePath = "/objectstore"
)

func (c *clusterClient) ObjectStoreInspect() (*objectstore.ObjectstoreInfo, error) {
	objectstoreInfo := &objectstore.ObjectstoreInfo{}
	request := c.c.Get().Resource(clusterPath + ObjectStorePath)
	if err := request.Do().Unmarshal(objectstoreInfo); err != nil {
		return nil, err
	}
	return objectstoreInfo, nil
}

func (c *clusterClient) ObjectStoreCreate(volume string) error {
	req := c.c.Post().Resource(clusterPath + ObjectStorePath)
	// Since volume name can be case sensitive adding it as
	// query param instead of path variable,
	// current rest client converts /path vars into lowercase format automatically
	req.QueryOption(objectstore.VolumeName, volume)
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

func (c *clusterClient) ObjectStoreUpdate(enable bool) error {
	req := c.c.Put().Resource(clusterPath + ObjectStorePath)
	req.QueryOption(objectstore.Enable, strconv.FormatBool(enable))
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}

func (c *clusterClient) ObjectStoreDelete() error {
	req := c.c.Delete().Resource(clusterPath + ObjectStorePath + "/delete")
	res := req.Do()
	if res.Error() != nil {
		return res.FormatError()
	}

	return nil
}
