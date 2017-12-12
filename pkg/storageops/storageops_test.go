package storageops_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/aws"
	"github.com/libopenstorage/openstorage/pkg/storageops/gce"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const (
	newDiskPrefix      = "openstorage-test"
	newDiskDescription = "Disk created by Openstorage tests"
	newDiskSizeInGB    = 10
)

var drivers map[string]storageops.Ops
var diskTemplates map[string]map[string]interface{}
var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.NewV4())

func initGCE(t *testing.T) (storageops.Ops, map[string]interface{}) {
	driver, err := gce.NewClient()
	require.NoError(t, err, "failed to instantiate storage ops driver")

	template := &compute.Disk{
		Description: newDiskDescription,
		Name:        diskName,
		SizeGb:      newDiskSizeInGB,
	}

	return driver, map[string]interface{}{
		diskName: template,
	}
}

func initDrivers(t *testing.T) {
	drivers = make(map[string]storageops.Ops)
	diskTemplates = make(map[string]map[string]interface{})

	// GCE
	if gce.IsDevMode() {
		d, disks := initGCE(t)
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = disks
	} else {
		fmt.Printf("skipping GCE tests as environment is not set...\n")
	}

	// AWS
	if d, err := aws.NewEnvClient(); err != aws.ErrAWSEnvNotAvailable {
		volType := opsworks.VolumeTypeGp2
		volSize := int64(newDiskSizeInGB)
		zone := os.Getenv("AWS_ZONE")
		ebsVol := &ec2.Volume{
			AvailabilityZone: &zone,
			VolumeType:       &volType,
			Size:             &volSize,
		}
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = map[string]interface{}{
			diskName: ebsVol,
		}
	} else {
		fmt.Printf("skipping AWS tests as environment is not set...\n")
	}
}

func TestAll(t *testing.T) {
	initDrivers(t)

	for _, d := range drivers {
		doAll(t, d)
	}

}

func doAll(t *testing.T, driver storageops.Ops) {
	name(t, driver)

	for _, template := range diskTemplates[driver.Name()] {
		d := create(t, driver, template)
		tags(t, driver, d)

		diskName := id(t, driver, d)
		enumerate(t, driver, diskName)
		inspect(t, driver, diskName)
		attach(t, driver, diskName)
		devicePath(t, driver, d)
		teardown(t, driver, diskName)
	}
}

func name(t *testing.T, driver storageops.Ops) {
	name := driver.Name()
	require.NotEmpty(t, name, "driver returned empty name")
}

func create(t *testing.T, driver storageops.Ops, template interface{}) interface{} {
	d, err := driver.Create(template, nil)
	require.NoError(t, err, "failed to create disk")
	require.NotNil(t, d, "got nil disk from create api")

	return d
}

func id(t *testing.T, driver storageops.Ops, disk interface{}) string {
	id := driver.GetDeviceID(disk)
	require.NotEmpty(t, id, "got empty disk name/ID")
	return id
}

func tags(t *testing.T, driver storageops.Ops, disk interface{}) {
	labels := map[string]string{
		"source": "openstorage-test",
		"foo":    "bar",
	}

	err := driver.ApplyTags(disk, labels)
	require.NoError(t, err, "failed to apply tags to disk")

	tags, err := driver.Tags(disk)
	require.NoError(t, err, "failed to get tags for disk")
	require.Len(t, tags, 2, "invalid number of labels found on disk")

	labelsToRemove := map[string]string{"foo": "bar"}
	err = driver.RemoveTags(disk, labelsToRemove)
	require.NoError(t, err, "failed to remove tags from disk")

	tags, err = driver.Tags(disk)
	require.NoError(t, err, "failed to get tags for disk")
	require.Len(t, tags, 1, "invalid number of labels found on disk")
}

func enumerate(t *testing.T, driver storageops.Ops, diskName string) {
	disks, err := driver.Enumerate([]*string{&diskName}, nil, storageops.SetIdentifierNone)
	require.NoError(t, err, "failed to create disk")
	require.Len(t, disks, 1, "inspect returned invalid length")
}

func inspect(t *testing.T, driver storageops.Ops, diskName string) {
	disks, err := driver.Inspect([]*string{&diskName})
	require.NoError(t, err, "failed to create disk")
	require.Len(t, disks, 1, fmt.Sprintf("inspect returned invalid length: %d", len(disks)))
}

func attach(t *testing.T, driver storageops.Ops, diskName string) {
	devPath, err := driver.Attach(diskName)
	require.NoError(t, err, "disk attach returned error")
	require.NotEmpty(t, devPath, "disk attach returned empty devicePath")

	mappings, err := driver.DeviceMappings()
	require.NoError(t, err, "get device mappings returned error")
	require.NotEmpty(t, mappings, "received empty device mappings")
}

func devicePath(t *testing.T, driver storageops.Ops, d interface{}) {
	devPath, err := driver.DevicePath(d)
	require.NoError(t, err, "get device path returned error")
	require.NotEmpty(t, devPath, "received empty devicePath")
}

func teardown(t *testing.T, driver storageops.Ops, diskName string) {
	err := driver.Detach(diskName)
	require.NoError(t, err, "disk detach returned error")

	time.Sleep(3 * time.Second)

	err = driver.Delete(diskName)
	require.NoError(t, err, "failed to delete disk")
}
