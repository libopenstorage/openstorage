package azure_test

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/azure"
	"github.com/libopenstorage/openstorage/pkg/storageops/test"
	"github.com/pborman/uuid"
)

const (
	newDiskSizeInGB = 10
	newDiskPrefix   = "openstorage-test"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.New())

func initAzure(t *testing.T) (storageops.Ops, map[string]interface{}) {
	driver, err := azure.NewEnvClient()
	if err != nil {
		t.Skipf("skipping Azure tests as environment is not set...\n")
	}

	size := int32(newDiskSizeInGB)
	name := diskName
	region := "eastus2"

	template := &compute.Disk{
		Name:     &name,
		Location: &region,
		DiskProperties: &compute.DiskProperties{
			DiskSizeGB:        &size,
			DiskIOPSReadWrite: to.Int64Ptr(1350),
			DiskMBpsReadWrite: to.Int32Ptr(550),
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
