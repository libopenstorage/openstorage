package nfs

import (
	"testing"

	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

func TestAll(t *testing.T) {

	_, err := volume.New(Name, volume.DriverParams{"uri": "localhost:/nfs"})
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}
	ctx, err := test.NewContext(Name)
	if err != nil {
		t.Fatalf("Failed to initialize Driver: %v", err)
	}

	test.RunShort(t, ctx)
}
