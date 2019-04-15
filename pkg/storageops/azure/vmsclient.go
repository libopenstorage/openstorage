package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

// vmsClient is an interface for azure vm client operations
type vmsClient interface {
	// describe returns the VM instance object
	describe(instanceID string) (interface{}, error)
	// getDataDisks returns a list of data disks attached to the given VM
	getDataDisks(instanceID string) ([]compute.DataDisk, error)
	// updateDataDisks update the data disks for the given VM
	updateDataDisks(instanceID string, dataDisks []compute.DataDisk) error
}

func NewVMsClient(
	scaleSetName string,
	subscriptionID, resourceGroupName string,
	authorizer autorest.Authorizer,
) vmsClient {
	if scaleSetName == "" {
		return newBaseVMsClient(
			subscriptionID,
			resourceGroupName,
			authorizer,
		)
	}
	return newScaleSetVMsClient(
		scaleSetName,
		subscriptionID,
		resourceGroupName,
		authorizer,
	)
}
