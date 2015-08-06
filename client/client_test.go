package client

import (
	"os"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/apiserver"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/drivers/btrfs"
	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

var (
	testPath = string("/tmp/openstorage_client_test")
)

func TestAll(t *testing.T) {
	err := os.MkdirAll(testPath, 0744)
	if err != nil {
		t.Fatalf("Failed to create test path: %v", err)
	}

	_, err = volume.New(btrfs.Name, volume.DriverParams{btrfs.RootParam: testPath})
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
	ctx.Filesystem = string("btrfs")
	test.Run(t, ctx)
}
