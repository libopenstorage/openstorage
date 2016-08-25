package testing

import (
	"os"
	"testing"
	"time"

	"go.pedge.io/dlog"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/api/server"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/libopenstorage/openstorage/volume/drivers/nfs"
	"github.com/libopenstorage/openstorage/volume/drivers/test"
)

var (
	testPath = string("/tmp/openstorage_client_test")
)

func init() {
	dlog.SetLevel(dlog.LevelDebug)
}

func makeRequest(t *testing.T) {
	versions, err := client.GetSupportedDriverVersions(nfs.Name, "")
	if err != nil {
		t.Fatalf("Failed to obtain supported versions. Err: %v", err)
	}
	if len(versions) == 0 {
		t.Fatalf("Versions array is empty")
	}
	c, err := client.NewDriverClient(nfs.Name, versions[0])
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	d := c.VolumeDriver()
	_, err = d.Inspect([]string{"foo"})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
}

func TestAll(t *testing.T) {
	err := os.MkdirAll(testPath, 0744)
	if err != nil {
		t.Fatalf("Failed to create test path: %v", err)
	}

	err = volumedrivers.Register(nfs.Name, map[string]string{"path": testPath})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}

	server.StartPluginAPI(
		nfs.Name,
		config.DriverAPIBase,
		config.PluginAPIBase,
		0,
		0,
	)
	time.Sleep(time.Second * 2)
	versions, err := client.GetSupportedDriverVersions(nfs.Name, "")
	if err != nil {
		t.Fatalf("Failed to obtain supported versions. Err: %v", err)
	}
	if len(versions) == 0 {
		t.Fatalf("Versions array is empty")
	}
	c, err := client.NewDriverClient(nfs.Name, versions[0])
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	d := c.VolumeDriver()
	ctx := test.NewContext(d)
	ctx.Filesystem = api.FSType_FS_TYPE_BTRFS
	test.Run(t, ctx)
}

func TestConnections(t *testing.T) {
	for i := 0; i < 2000; i++ {
		makeRequest(t)
	}
}
