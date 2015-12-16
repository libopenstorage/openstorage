package client

import (
	"os"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/server"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/nfs"
	"github.com/libopenstorage/openstorage/volume/drivers/test"
)

var (
	testPath = string("/tmp/openstorage_client_test")
)

func makeRequest(t *testing.T) {
	c, err := NewDriverClient(nfs.Name)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	d := c.VolumeDriver()
	_, err = d.Inspect([]api.VolumeID{api.VolumeID("foo")})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
}

func TestAll(t *testing.T) {
	err := os.MkdirAll(testPath, 0744)
	if err != nil {
		t.Fatalf("Failed to create test path: %v", err)
	}

	_, err = volume.New(nfs.Name, volume.DriverParams{"path": testPath})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	server.StartServerAPI(nfs.Name, 9003, config.DriverAPIBase)
	time.Sleep(time.Second * 2)
	c, err := NewDriverClient(nfs.Name)
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	d := c.VolumeDriver()
	ctx := test.NewContext(d)
	ctx.Filesystem = string("btrfs")
	test.Run(t, ctx)
}

func TestConnections(t *testing.T) {
	for i := 0; i < 2000; i++ {
		makeRequest(t)
	}
}
