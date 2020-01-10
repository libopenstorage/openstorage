/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sdk

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/libopenstorage/openstorage/alerts"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/role"
	policy "github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// Default audig log location
	defaultAuditLog = "/var/log/openstorage-audit.log"
	// Default access log location
	defaultAccessLog = "/var/log/openstorage-access.log"
	// ContextDriverKey is the driver key passed in context's metadata
	ContextDriverKey = "driver"
	// DefaultDriverName is the default driver to be used
	DefaultDriverName = "default"
)

// TLSConfig points to the cert files needed for HTTPS
type TLSConfig struct {
	// CertFile is the path to the cert file
	CertFile string
	// KeyFile is the path to the key file
	KeyFile string
}

// SecurityConfig provides configuration for SDK auth
type SecurityConfig struct {
	// Role implementation
	Role role.RoleManager
	// Tls configuration
	Tls *TLSConfig
	// Authenticators per issuer. You can register multple authenticators
	// based on the "iss" string in the string. For example:
	// map[string]auth.Authenticator {
	//     "https://accounts.google.com": googleOidc,
	//     "openstorage-sdk-auth: selfSigned,
	// }
	Authenticators map[string]auth.Authenticator
	// PublicVolumeCreationDisabled controls whether or not we can create
	// public volumes with no ownership in an authenticated system.
	PublicVolumeCreationDisabled bool
}

// ServerConfig provides the configuration to the SDK server
type ServerConfig struct {
	// Net is the transport for gRPC: unix, tcp, etc.
	// For the gRPC Server. This value goes together with `Address`.
	Net string
	// Address is the port number or the unix domain socket path.
	// For the gRPC Server. This value goes together with `Net`.
	Address string
	// RestAdress is the port number. Example: 9110
	// For the gRPC REST Gateway.
	RestPort string
	// Unix domain socket for local communication. This socket
	// will be used by the REST Gateway to communicate with the gRPC server.
	// Only set for testing. Having a '%s' can be supported to use the
	// name of the driver as the driver name.
	Socket string
	// (optional) Location for audit log.
	// If not provided, it will go to /var/log/openstorage-audit.log
	AuditOutput io.Writer
	// (optional) Location of access log.
	// This is useful when authorization is not running.
	// If not provided, it will go to /var/log/openstorage-access.log
	AccessOutput io.Writer
	// (optional) The OpenStorage driver to use
	DriverName string
	// (optional) Cluster interface
	Cluster cluster.Cluster
	// AlertsFilterDeleter
	AlertsFilterDeleter alerts.FilterDeleter
	// StoragePolicy Manager
	StoragePolicy policy.PolicyManager
	// StoragePoolServer is the interface to manage storage pools in the cluster
	StoragePoolServer api.OpenStoragePoolServer
	// Security configuration
	Security *SecurityConfig
	// ServerExtensions allows you to extend the SDK gRPC server
	// with callback functions that are sequentially executed
	// at the end of Server.Start()
	//
	// To add your own service to the SDK gRPC server,
	// just append a function callback that registers it:
	//
	// s.config.ServerExtensions = append(s.config.ServerExtensions,
	// 		func(gs *grpc.Server) {
	//			api.RegisterCustomService(gs, customHandler)
	//		})
	GrpcServerExtensions []func(grpcServer *grpc.Server)

	// RestServerExtensions allows for extensions to be added
	// to the SDK Rest Gateway server.
	//
	// To add your own service to the SDK REST Server, simply add your handlers
	// to the RestSererExtensions slice. These handlers will be registered on the
	// REST Gateway http server.
	RestServerExtensions []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
}

// Server is an implementation of the gRPC SDK interface
type Server struct {
	config      ServerConfig
	netServer   *sdkGrpcServer
	udsServer   *sdkGrpcServer
	restGateway *sdkRestGateway

	accessLog *os.File
	auditLog  *os.File
}

type serverAccessor interface {
	alert() alerts.FilterDeleter
	cluster() cluster.Cluster
	driver(ctx context.Context) volume.VolumeDriver
}

type logger struct {
	log *logrus.Entry
}

type sdkGrpcServer struct {
	*grpcserver.GrpcServer

	restPort string
	lock     sync.RWMutex
	name     string
	config   ServerConfig

	// Loggers
	log             *logrus.Entry
	auditLogOutput  io.Writer
	accessLogOutput io.Writer

	// Interface implementations
	clusterHandler cluster.Cluster
	driverHandlers map[string]volume.VolumeDriver
	alertHandler   alerts.FilterDeleter

	// gRPC Handlers
	clusterServer         *ClusterServer
	nodeServer            *NodeServer
	volumeServer          *VolumeServer
	objectstoreServer     *ObjectstoreServer
	schedulePolicyServer  *SchedulePolicyServer
	clusterPairServer     *ClusterPairServer
	cloudBackupServer     *CloudBackupServer
	credentialServer      *CredentialServer
	identityServer        *IdentityServer
	clusterDomainsServer  *ClusterDomainsServer
	roleServer            role.RoleManager
	alertsServer          api.OpenStorageAlertsServer
	policyServer          policy.PolicyManager
	storagePoolServer     api.OpenStoragePoolServer
	filesystemTrimServer  api.OpenStorageFilesystemTrimServer
	filesystemCheckServer api.OpenStorageFilesystemCheckServer
}

// Interface check
var _ grpcserver.Server = &sdkGrpcServer{}

// New creates a new SDK server
func New(config *ServerConfig) (*Server, error) {

	if config == nil {
		return nil, fmt.Errorf("Must provide configuration")
	}

	// If no security set, initialize the object as empty
	if config.Security == nil {
		config.Security = &SecurityConfig{}
	}

	// Check if the socket is provided to enable the REST gateway to communicate
	// to the unix domain socket
	if len(config.Socket) == 0 {
		return nil, fmt.Errorf("Must provide unix domain socket for SDK")
	}
	if len(config.RestPort) == 0 {
		return nil, fmt.Errorf("Must provide REST Gateway port for the SDK")
	}

	// Set default log locations
	var (
		accessLog, auditLog *os.File
		err                 error
	)
	if config.AuditOutput == nil {
		auditLog, err = openLog(defaultAuditLog)
		if err != nil {
			return nil, err
		}
		config.AuditOutput = auditLog
	}
	if config.AccessOutput == nil {
		accessLog, err := openLog(defaultAccessLog)
		if err != nil {
			return nil, err
		}
		config.AccessOutput = accessLog
	}

	// Create a gRPC server on the network
	netServer, err := newSdkGrpcServer(config)
	if err != nil {
		return nil, err
	}

	// Create a gRPC server on a unix domain socket
	udsConfig := *config
	udsConfig.Net = "unix"
	udsConfig.Address = config.Socket
	udsServer, err := newSdkGrpcServer(&udsConfig)
	if err != nil {
		return nil, err
	}

	// Create REST Gateway and connect it to the unix domain socket server
	restGateway, err := newSdkRestGateway(config, udsServer)
	if err != nil {
		return nil, err
	}

	return &Server{
		config:      *config,
		netServer:   netServer,
		udsServer:   udsServer,
		restGateway: restGateway,
		auditLog:    auditLog,
		accessLog:   accessLog,
	}, nil
}

// Start all servers
func (s *Server) Start() error {
	if err := s.netServer.Start(); err != nil {
		return err
	} else if err := s.udsServer.Start(); err != nil {
		return err
	} else if err := s.restGateway.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() {
	s.netServer.Stop()
	s.udsServer.Stop()
	s.restGateway.Stop()

	if s.accessLog != nil {
		s.accessLog.Close()
	}
	if s.auditLog != nil {
		s.auditLog.Close()
	}
}

func (s *Server) Address() string {
	return s.netServer.Address()
}

func (s *Server) UdsAddress() string {
	return s.udsServer.Address()
}

// UseCluster will setup a new cluster object for the gRPC handlers
func (s *Server) UseCluster(c cluster.Cluster) {
	s.netServer.useCluster(c)
	s.udsServer.useCluster(c)
}

// UseVolumeDrivers will setup a new driver object for the gRPC handlers
func (s *Server) UseVolumeDrivers(d map[string]volume.VolumeDriver) {
	s.netServer.useVolumeDrivers(d)
	s.udsServer.useVolumeDrivers(d)
}

// UseAlert will setup a new alert object for the gRPC handlers
func (s *Server) UseAlert(a alerts.FilterDeleter) {
	s.netServer.useAlert(a)
	s.udsServer.useAlert(a)
}

// New creates a new SDK gRPC server
func newSdkGrpcServer(config *ServerConfig) (*sdkGrpcServer, error) {
	if nil == config {
		return nil, fmt.Errorf("Configuration must be provided")
	}

	// Create a log object for this server
	name := "SDK-" + config.Net
	log := logrus.WithFields(logrus.Fields{
		"name": name,
	})

	// Save the driver for future calls
	var (
		d   volume.VolumeDriver
		err error
	)
	if len(config.DriverName) != 0 {
		d, err = volumedrivers.Get(config.DriverName)
		if err != nil {
			return nil, fmt.Errorf("Unable to get driver %s info: %s", config.DriverName, err.Error())
		}
	}

	// Setup authentication
	for issuer, _ := range config.Security.Authenticators {
		log.Infof("Authentication enabled for issuer: %s", issuer)

		// Check the necessary security config options are set
		if config.Security.Role == nil {
			return nil, fmt.Errorf("Must supply role manager when authentication enabled")
		}
	}

	if config.StoragePolicy == nil {
		return nil, fmt.Errorf("Must supply storage policy server")
	}

	// Create gRPC server
	gServer, err := grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    name,
		Net:     config.Net,
		Address: config.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to setup %s server: %v", name, err)
	}

	s := &sdkGrpcServer{
		GrpcServer:      gServer,
		accessLogOutput: config.AccessOutput,
		auditLogOutput:  config.AuditOutput,
		config:          *config,
		name:            name,
		log:             log,
		clusterHandler:  config.Cluster,
		driverHandlers: map[string]volume.VolumeDriver{
			config.DriverName: d,
			DefaultDriverName: d,
		},
		alertHandler:      config.AlertsFilterDeleter,
		policyServer:      config.StoragePolicy,
		storagePoolServer: config.StoragePoolServer,
	}
	s.identityServer = &IdentityServer{
		server: s,
	}
	s.clusterServer = &ClusterServer{
		server: s,
	}
	s.nodeServer = &NodeServer{
		server: s,
	}
	s.volumeServer = &VolumeServer{
		server:      s,
		specHandler: spec.NewSpecHandler(),
	}
	s.objectstoreServer = &ObjectstoreServer{
		server: s,
	}
	s.schedulePolicyServer = &SchedulePolicyServer{
		server: s,
	}
	s.cloudBackupServer = &CloudBackupServer{
		server: s,
	}
	s.credentialServer = &CredentialServer{
		server: s,
	}
	s.alertsServer = &alertsServer{
		server: s,
	}
	s.clusterPairServer = &ClusterPairServer{
		server: s,
	}
	s.clusterDomainsServer = &ClusterDomainsServer{
		server: s,
	}
	s.filesystemTrimServer = &FilesystemTrimServer{
		server: s,
	}
	s.filesystemCheckServer = &FilesystemCheckServer{
		server: s,
	}

	s.roleServer = config.Security.Role
	s.policyServer = config.StoragePolicy
	s.storagePoolServer = config.StoragePoolServer
	return s, nil
}

// Start is used to start the server.
// It will return an error if the server is already running.
func (s *sdkGrpcServer) Start() error {

	// Setup https if certs have been provided
	opts := make([]grpc.ServerOption, 0)
	if s.config.Net != "unix" && s.config.Security.Tls != nil {
		creds, err := credentials.NewServerTLSFromFile(
			s.config.Security.Tls.CertFile,
			s.config.Security.Tls.KeyFile)
		if err != nil {
			return fmt.Errorf("Failed to create credentials from cert files: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		s.log.Info("SDK TLS enabled")
	} else {
		s.log.Info("SDK TLS disabled")
	}

	// Setup authentication and authorization using interceptors if auth is enabled
	if len(s.config.Security.Authenticators) != 0 {
		opts = append(opts, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				s.rwlockIntercepter,
				grpc_auth.UnaryServerInterceptor(s.auth),
				s.authorizationServerInterceptor,
				s.loggerServerInterceptor,
			)))
	} else {
		opts = append(opts, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				s.rwlockIntercepter,
				s.loggerServerInterceptor,
			)))
	}

	// Start the gRPC Server
	err := s.GrpcServer.StartWithServer(func() *grpc.Server {
		grpcServer := grpc.NewServer(opts...)

		api.RegisterOpenStorageClusterServer(grpcServer, s.clusterServer)
		api.RegisterOpenStorageNodeServer(grpcServer, s.nodeServer)
		api.RegisterOpenStorageObjectstoreServer(grpcServer, s.objectstoreServer)
		api.RegisterOpenStorageSchedulePolicyServer(grpcServer, s.schedulePolicyServer)
		api.RegisterOpenStorageIdentityServer(grpcServer, s.identityServer)
		api.RegisterOpenStorageVolumeServer(grpcServer, s.volumeServer)
		api.RegisterOpenStorageMigrateServer(grpcServer, s.volumeServer)
		api.RegisterOpenStorageCredentialsServer(grpcServer, s.credentialServer)
		api.RegisterOpenStorageCloudBackupServer(grpcServer, s.cloudBackupServer)
		api.RegisterOpenStorageMountAttachServer(grpcServer, s.volumeServer)
		api.RegisterOpenStorageAlertsServer(grpcServer, s.alertsServer)
		api.RegisterOpenStorageClusterPairServer(grpcServer, s.clusterPairServer)
		api.RegisterOpenStoragePolicyServer(grpcServer, s.policyServer)
		api.RegisterOpenStorageClusterDomainsServer(grpcServer, s.clusterDomainsServer)
		api.RegisterOpenStorageFilesystemTrimServer(grpcServer, s.filesystemTrimServer)
		api.RegisterOpenStorageFilesystemCheckServer(grpcServer, s.filesystemCheckServer)
		if s.storagePoolServer != nil {
			api.RegisterOpenStoragePoolServer(grpcServer, s.storagePoolServer)
		}

		if s.config.Security.Role != nil {
			api.RegisterOpenStorageRoleServer(grpcServer, s.roleServer)
		}

		s.registerServerExtensions(grpcServer)

		return grpcServer
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *sdkGrpcServer) registerServerExtensions(grpcServer *grpc.Server) {
	for _, ext := range s.config.GrpcServerExtensions {
		ext(grpcServer)
	}
}

func (s *sdkGrpcServer) useCluster(c cluster.Cluster) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.clusterHandler = c
}

func (s *sdkGrpcServer) useVolumeDrivers(d map[string]volume.VolumeDriver) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.driverHandlers = d
}

func (s *sdkGrpcServer) useAlert(a alerts.FilterDeleter) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.alertHandler = a
}

// Accessors
func (s *sdkGrpcServer) driver(ctx context.Context) volume.VolumeDriver {
	driverName := grpcserver.GetMetadataValueFromKey(ctx, ContextDriverKey)
	if handler, ok := s.driverHandlers[driverName]; ok {
		return handler
	} else {
		return s.driverHandlers[DefaultDriverName]
	}
}

func (s *sdkGrpcServer) cluster() cluster.Cluster {
	return s.clusterHandler
}

func (s *sdkGrpcServer) alert() alerts.FilterDeleter {
	return s.alertHandler
}
