package sdk

import (
	"github.com/google/go-cloud/wire"
)

// ProviderSet is a collection of service providers.
// Such set is used by the wire command to generate initialization code.
var ProviderSet = wire.NewSet(
	NewServer,
	NewServerConfig,
	NewGrpcServer,
	NewAlertsServer,
	NewCloudBackupServer,
	NewClusterServer,
	NewCredentialServer,
	NewIdentityServer,
	NewNodeServer,
	NewObjectstoreServer,
	NewSchedulePolicyServer,
	NewVolumeServer, NewVolumeDriver, NewSpecHandler,
)
