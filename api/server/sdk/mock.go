package sdk

import (
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const MockDriver = "mock"

// NewMockNet is a provider of mock net.
func NewMockNet() Net { return Net("tcp") }

// NewMockAddress is a provider of mock address.
func NewMockAddress() Address { return Address("127.0.0.1:0") }

// NewMockDriver is a provider of MockDriver.
func NewMockDriver(volDriver *mockdriver.MockVolumeDriver) (Driver, error) {
	volumedrivers.Add(MockDriver,
		func(map[string]string) (volume.VolumeDriver, error) {
			return volDriver, nil
		},
	)

	// Register mock driver and return
	return Driver(MockDriver), volumedrivers.Register(MockDriver, nil)
}

// NewMockTestReporter is a provider for TestReporter
func NewMockTestReporter() gomock.TestReporter {
	return &utils.SafeGoroutineTester{}
}
