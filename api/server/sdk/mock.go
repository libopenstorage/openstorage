package sdk

import (
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/libopenstorage/openstorage/alerts"
	"github.com/libopenstorage/openstorage/alerts/mock"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const MockDriver = "mock"

// NewMockNet is a provider of mock net.
func NewMockNet() Net { return Net("tcp") }

// NewMockRestPort is a provider of mock rest port.
func NewMockRestPort() RestPort { return RestPort("") }

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

// NewMockCluster is a provider of mock cluster.
func NewMockCluster(cluster *mockcluster.MockCluster) cluster.Cluster {
	return cluster
}

// NewMockFilterDeleter is a provider of mock filter deleter.
func NewMockFilterDeleter(filterDelter *mockalerts.MockFilterDeleter) alerts.FilterDeleter {
	return filterDelter
}

// NewMockTestReporter is a provider for TestReporter
func NewMockTestReporter() gomock.TestReporter {
	return &utils.SafeGoroutineTester{}
}
