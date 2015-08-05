package client

import (
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/apiserver"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/drivers/btrfs"
	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

func TestAll(t *testing.T) {

	_, err := volume.New(btrfs.Name, volume.DriverParams{btrfs.RootParam: "/tmp/btrfs_test"})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	apiserver.StartDriverAPI(btrfs.Name, 9003, config.DriverAPIBase)
	time.Sleep(time.Second * 2)
	c, err := NewDriverClient(btrfs.Name)
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	d := c.VolumeDriver()
	ctx := test.NewContext(d)
	ctx.Filesystem = string(api.FsBtrfs)
	test.Run(t, ctx)
}
