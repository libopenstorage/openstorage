package aws

import (
	"testing"

	"github.com/libopenstorage/openstorage/drivers/test"
	"github.com/libopenstorage/openstorage/volume"
)

func TestAll(t *testing.T) {
	_, err := volume.New(Name, volume.DriverParams{})
	if err != nil {
		t.Logf("Failed to initialize Driver: %v", err)
	}
	d, err := volume.Get(Name)
	if err != nil {
		t.Fatalf("Failed to initialize Volume Driver: %v", err)
	}
	ctx := test.NewContext(d)
	ctx.Filesystem = "ext4"
	test.RunShort(t, ctx)
}
