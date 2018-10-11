package sdk

import (
	"github.com/golang/mock/gomock"
	"github.com/google/go-cloud/wire"
	"github.com/libopenstorage/openstorage/alerts/mock"
	"github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/volume/drivers/mock"
)

// ProviderSet is a collection of service providers.
// Such set is used by the wire command to generate initialization code.
var ProviderSet = wire.NewSet(
	NewServerProvider,
	NewServerConfig,
	NewGrpcServer,
	ServiceProviderSet,
)

// MockProviderSet is a collection of service providers.
// Such set is used by the wire command to generate initialization code.
var MockProviderSet = wire.NewSet(
	NewServerProvider,
	NewServerConfig,
	NewGrpcServer,
	ServiceProviderSet,
	gomock.NewController,
	mockcluster.NewMockCluster,
	mockdriver.NewMockVolumeDriver,
	mockalerts.NewMockFilterDeleter,
	NewMockNet,
	NewMockAddress,
	NewMockDriver,
	NewMockTestReporter,
)

// ServiceProviderSet is a set of providers for various server subsystems.
var ServiceProviderSet = wire.NewSet(
	NewAlertsServer,
	NewCloudBackupServer,
	NewClusterServer,
	NewCredentialServer,
	NewIdentityServer,
	NewNodeServer,
	NewObjectstoreServer,
	NewSchedulePolicyServer,
	VolumeProviderSet,
)

// VolumeProviderSet is a set of providers for various volume subsystems.
var VolumeProviderSet = wire.NewSet(
	NewVolumeServer,
	NewVolumeDriver,
	NewSpecHandler,
)
