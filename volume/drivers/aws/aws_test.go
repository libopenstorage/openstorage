package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers/test"
	"github.com/stretchr/testify/require"
)

func testRemoveTags(t *testing.T, driver volume.VolumeDriver) {
	d := driver.(*Driver)
	// Create volume with labels
	sz := int64(1)
	voltype := opsworks.VolumeTypeIo1
	ec2Vol := &ec2.Volume{
		AvailabilityZone: &d.md.zone,
		VolumeType:       &voltype,
		Size:             &sz,
	}
	labelNames := []string{"label1", "label2"}
	labels := make(map[string]string)
	for _, name := range labelNames {
		labels[name] = name
	}
	vol, err := d.ops.Create(ec2Vol, labels)
	require.Nil(t, err, "Failed in CreateVolumeRequest :%v", err)
	defer d.ops.Delete(*vol.VolumeId)
	require.True(t, len(d.ops.Tags(vol)) == len(labelNames), "ApplyTags failed")
	require.Nil(t, d.ops.RemoveTags(vol, labels), "RemoveTags error")
	require.True(t, len(d.ops.Tags(vol)) == 0, "RemoveTags failed")
}

func TestAll(t *testing.T) {
	if _, err := credentials.NewEnvCredentials().Get(); err != nil {
		t.Skip("No AWS credentials, skipping AWS driver test: ", err)
	}
	driver, err := Init(map[string]string{})
	if err != nil {
		t.Fatalf("Failed to initialize Volume Driver: %v", err)
	}
	ctx := test.NewContext(driver)
	ctx.Filesystem = api.FSType_FS_TYPE_EXT4
	test.RunShort(t, ctx)
	testRemoveTags(t, driver)

}
