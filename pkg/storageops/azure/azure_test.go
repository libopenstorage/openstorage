package azure_test

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-12-01/compute"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/azure"
	"github.com/libopenstorage/openstorage/pkg/storageops/test"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

const (
	newDiskSizeInGB = 10
	newDiskPrefix   = "openstorage-test"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.New())

func initAzure(t *testing.T) (storageops.Ops, map[string]interface{}) {
	driver, err := azure.NewEnvClient()
	require.NoError(t, err, "failed to instantiate storage ops driver")

	size := int32(newDiskSizeInGB)
	name := diskName
	region := "eastus2"

	template := &compute.Disk{
		Name:     &name,
		Location: &region,
		DiskProperties: &compute.DiskProperties{
			DiskSizeGB: &size,
		},
		Sku: &compute.DiskSku{
			Name: compute.PremiumLRS,
		},
	}

	return driver, map[string]interface{}{
		diskName: template,
	}
}

func TestAll(t *testing.T) {
	drivers := make(map[string]storageops.Ops)
	diskTemplates := make(map[string]map[string]interface{})

	d, disks := initAzure(t)
	drivers[d.Name()] = d
	diskTemplates[d.Name()] = disks
	test.RunTest(drivers, diskTemplates, t)
}
