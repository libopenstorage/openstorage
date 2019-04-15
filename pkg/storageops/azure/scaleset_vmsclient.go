package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

type scaleSetVMsClient struct {
	scaleSetName      string
	resourceGroupName string
	client            *compute.VirtualMachineScaleSetVMsClient
}

func newScaleSetVMsClient(
	scaleSetName, subscriptionID, resourceGroupName string,
	authorizer autorest.Authorizer,
) vmsClient {
	vmsClient := compute.NewVirtualMachineScaleSetVMsClient(subscriptionID)
	vmsClient.Authorizer = authorizer
	vmsClient.PollingDelay = clientPollingDelay
	vmsClient.RetryAttempts = clientRetryAttempts
	vmsClient.AddToUserAgent(userAgentExtension)
	return &scaleSetVMsClient{
		scaleSetName:      scaleSetName,
		resourceGroupName: resourceGroupName,
		client:            &vmsClient,
	}
}

func (s *scaleSetVMsClient) describe(
	instanceID string,
) (interface{}, error) {
	return s.describeInstance(instanceID)
}

func (s *scaleSetVMsClient) getDataDisks(
	instanceID string,
) ([]compute.DataDisk, error) {
	vm, err := s.describeInstance(instanceID)
	if err != nil {
		return nil, fmt.Errorf("cannot get vm %v_%v: %v", s.scaleSetName, instanceID, err)
	}

	if vm.StorageProfile == nil || vm.StorageProfile.DataDisks == nil {
		return nil, fmt.Errorf("vm storage profile is invalid")
	}

	return *vm.StorageProfile.DataDisks, nil
}

func (s *scaleSetVMsClient) updateDataDisks(
	instanceID string,
	dataDisks []compute.DataDisk,
) error {
	vm, err := s.describeInstance(instanceID)
	if err != nil {
		return fmt.Errorf("cannot get vm %v_%v: %v", s.scaleSetName, instanceID, err)
	}

	vm.StorageProfile.DataDisks = &dataDisks

	ctx := context.Background()
	future, err := s.client.Update(
		ctx,
		s.resourceGroupName,
		s.scaleSetName,
		instanceID,
		vm,
	)
	if err != nil {
		return fmt.Errorf("cannot update vm %v_%v: %v", s.scaleSetName, instanceID, err)
	}

	err = future.WaitForCompletionRef(ctx, s.client.Client)
	if err != nil {
		return fmt.Errorf("cannot get the vm update future response: %v", err)
	}
	return nil
}

func (s *scaleSetVMsClient) describeInstance(
	instanceID string,
) (compute.VirtualMachineScaleSetVM, error) {
	return s.client.Get(
		context.Background(),
		s.resourceGroupName,
		s.scaleSetName,
		instanceID,
	)
}
