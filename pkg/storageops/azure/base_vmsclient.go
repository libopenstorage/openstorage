package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

type baseVMsClient struct {
	resourceGroupName string
	client            *compute.VirtualMachinesClient
}

func newBaseVMsClient(
	subscriptionID, resourceGroupName string,
	authorizer autorest.Authorizer,
) vmsClient {
	vmsClient := compute.NewVirtualMachinesClient(subscriptionID)
	vmsClient.Authorizer = authorizer
	vmsClient.PollingDelay = clientPollingDelay
	vmsClient.RetryAttempts = clientRetryAttempts
	vmsClient.AddToUserAgent(userAgentExtension)
	return &baseVMsClient{
		resourceGroupName: resourceGroupName,
		client:            &vmsClient,
	}
}

func (b *baseVMsClient) describe(
	instanceName string,
) (interface{}, error) {
	return b.describeInstance(instanceName)
}

func (b *baseVMsClient) getDataDisks(
	instanceName string,
) ([]compute.DataDisk, error) {
	vm, err := b.describeInstance(instanceName)
	if err != nil {
		return nil, fmt.Errorf("cannot get vm %v: %v", instanceName, err)
	}

	if vm.StorageProfile == nil || vm.StorageProfile.DataDisks == nil {
		return nil, fmt.Errorf("vm storage profile is invalid")
	}

	return *vm.StorageProfile.DataDisks, nil
}

func (b *baseVMsClient) updateDataDisks(
	instanceName string,
	dataDisks []compute.DataDisk,
) error {
	updatedVM := compute.VirtualMachineUpdate{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				DataDisks: &dataDisks,
			},
		},
	}

	ctx := context.Background()
	future, err := b.client.Update(
		ctx,
		b.resourceGroupName,
		instanceName,
		updatedVM,
	)
	if err != nil {
		return fmt.Errorf("cannot update vm %v: %v", instanceName, err)
	}

	err = future.WaitForCompletionRef(ctx, b.client.Client)
	if err != nil {
		return fmt.Errorf("cannot get the vm update future response: %v", err)
	}
	return nil
}

func (b *baseVMsClient) describeInstance(
	instanceName string,
) (compute.VirtualMachine, error) {
	return b.client.Get(
		context.Background(),
		b.resourceGroupName,
		instanceName,
		compute.InstanceView,
	)
}
