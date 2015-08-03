package client

import (
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/apiserver"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/drivers/nfs"
	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

func TestAll(t *testing.T) {

	_, err := volume.New(nfs.Name, volume.DriverParams{"uri": "localhost:/nfs"})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	apiserver.StartDriverAPI(nfs.Name, 9003, config.DriverAPIBase)
	time.Sleep(time.Second * 2)
	c, err := NewDriverClient(nfs.Name)
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	d := c.VolumeDriver()
	ctx := test.NewContext(d)
	test.RunShort(t, ctx)
}
