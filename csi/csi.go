/*
Package csi is CSI driver interface for OSD
Copyright 2017 Portworx

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
package csi

import (
	"fmt"
	"os"
	"sync"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/csi/sched/k8s"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/libopenstorage/openstorage/pkg/loadbalancer"
	"github.com/libopenstorage/openstorage/pkg/options"
	"github.com/portworx/kvdb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	authsecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
)

var (
	clogger                *logrus.Logger
	csiSocketCheckInterval = 30 * time.Second
)

func init() {
	clogger = correlation.NewPackageLogger(correlation.ComponentCSIDriver)
}

// OsdCsiServerConfig provides the configuration to the
// the gRPC CSI server created by NewOsdCsiServer()
type OsdCsiServerConfig struct {
	Net                string
	Address            string
	DriverName         string
	Cluster            cluster.Cluster
	RoundRobinBalancer loadbalancer.Balancer
	CloudBackupClient  api.OpenStorageCloudBackupClient
	SdkUds             string
	SdkPort            string
	SchedulerName      string

	// Name to be reported back to the CO. If not provided,
	// the name will be in the format of <driver>.openstorage.org
	CsiDriverName string

	// EnableInlineVolumes decides whether or not we will allow
	// creation of inline volumes.
	EnableInlineVolumes bool
}

// OsdCsiServer is a OSD CSI compliant server which
// proxies CSI requests for a single specific driver
type OsdCsiServer struct {
	csi.ControllerServer
	csi.NodeServer
	csi.IdentityServer

	*grpcserver.GrpcServer
	specHandler        spec.SpecHandler
	driver             volume.VolumeDriver
	cluster            cluster.Cluster
	sdkUds             string
	sdkPort            string
	conn               *grpc.ClientConn
	mu                 sync.Mutex
	csiDriverName      string
	allowInlineVolumes bool
	roundRobinBalancer loadbalancer.Balancer
	cloudBackupClient  api.OpenStorageCloudBackupClient
	volumeClient       api.OpenStorageVolumeClient
	config             *OsdCsiServerConfig
	autoRecoverStopCh  chan struct{}
	stopCleanupCh      chan bool
}

// NewOsdCsiServer creates a gRPC CSI complient server on the
// specified port and transport.
func NewOsdCsiServer(config *OsdCsiServerConfig) (grpcserver.Server, error) {
	if nil == config {
		return nil, fmt.Errorf("Must supply configuration")
	}
	if len(config.SdkUds) == 0 {
		return nil, fmt.Errorf("SdkUds must be provided")
	}
	if len(config.DriverName) == 0 {
		return nil, fmt.Errorf("OSD Driver name must be provided")
	}
	if config.EnableInlineVolumes {
		clogger.Warnf("CSI ephemeral inline volumes are deprecated and will be disabled by default in the future")
	}
	// Save the driver for future calls
	d, err := volumedrivers.Get(config.DriverName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get driver %s info: %s", config.DriverName, err.Error())
	}

	gServer, err := createGrpcServer(config)
	if err != nil {
		return nil, err
	}

	return &OsdCsiServer{
		specHandler:        spec.NewSpecHandler(),
		GrpcServer:         gServer,
		driver:             d,
		cluster:            config.Cluster,
		sdkUds:             config.SdkUds,
		sdkPort:            config.SdkPort,
		csiDriverName:      config.CsiDriverName,
		allowInlineVolumes: config.EnableInlineVolumes,
		roundRobinBalancer: config.RoundRobinBalancer,
		config:             config,
		autoRecoverStopCh:  make(chan struct{}),
		cloudBackupClient:  config.CloudBackupClient,
	}, nil
}

func (s *OsdCsiServer) getConn() (*grpc.ClientConn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		var err error
		fmt.Println("Connecting to", s.sdkUds)
		s.conn, err = grpcserver.Connect(
			s.sdkUds,
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
			})
		if err != nil {
			return nil, fmt.Errorf("Failed to connect CSI to SDK uds %s: %v", s.sdkUds, err)
		}
	}

	return s.conn, nil
}

func (s *OsdCsiServer) getRemoteConn(ctx context.Context) (*grpc.ClientConn, error) {
	remoteConn, _, err := s.roundRobinBalancer.GetRemoteNodeConnection(ctx)
	return remoteConn, err
}

// driverGetVolume returns a volume for a given ID. This function skips
// PX security authentication and should be used only when a CSI request
// does not support secrets as a field
func (s *OsdCsiServer) driverGetVolume(ctx context.Context, id string) (*api.Volume, error) {
	vols, err := s.driver.Inspect([]string{id})
	if err != nil || len(vols) < 1 {
		if err == kvdb.ErrNotFound {
			clogger.WithContext(ctx).Infof("Volume %s cannot be found: %s", id, err.Error())
			return nil, status.Errorf(codes.NotFound, "Failed to find volume with id %s", id)
		} else if err != nil {
			return nil, status.Errorf(codes.NotFound, "Volume id %s not found: %s",
				id,
				err.Error())
		} else {
			clogger.WithContext(ctx).Infof("Volume %s cannot be found", id)
			return nil, status.Errorf(codes.NotFound, "Failed to find volume with id %s", id)
		}
	}
	vol := vols[0]

	return vol, nil
}

// Gets token from the secrets. In Kubernetes, the side car containers copy
// the contents of a K8S Secret map into the Secrets section of the CSI call.
// Also adds correlation ID to the outgoing context
func (s *OsdCsiServer) setupContext(ctx context.Context, csiSecrets map[string]string) context.Context {
	metadataMap := make(map[string]string)
	if token, ok := csiSecrets[authsecrets.SecretTokenKey]; ok {
		metadataMap["authorization"] = "bearer " + token
	}

	// Create and return metadata from map
	if len(metadataMap) > 0 {
		md := metadata.New(metadataMap)
		return metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}

// addEncryptionInfoToLabels adds the needed secret encryption
// fields to locator.VolumeLabels.
func (s *OsdCsiServer) addEncryptionInfoToLabels(labels, csiSecrets map[string]string) map[string]string {
	if len(csiSecrets) == 0 {
		return labels
	}

	if s, exists := csiSecrets[options.OptionsSecret]; exists {
		labels[options.OptionsSecret] = s

		if context, exists := csiSecrets[options.OptionsSecretContext]; exists {
			labels[options.OptionsSecretContext] = context
		}

		if secretKey, exists := csiSecrets[options.OptionsSecretKey]; exists {
			labels[options.OptionsSecretKey] = secretKey
		}
	}

	return labels
}

// Start is used to start the server.
// It will return an error if the server is already running.
func (s *OsdCsiServer) Start() error {
	if err := s.GrpcServer.Start(func(grpcServer *grpc.Server) {
		csi.RegisterIdentityServer(grpcServer, s)
		csi.RegisterControllerServer(grpcServer, s)
		csi.RegisterNodeServer(grpcServer, s)
	}); err != nil {
		return err
	}

	if s.config.Net == "unix" {
		go func() {
			err := autoSocketRecover(s, s.autoRecoverStopCh)
			if err != nil {
				logrus.Errorf("failed to start CSI driver socket auto-recover watcher: %v", err)
			}
		}()
	}

	return nil
}

// Start is used to stop the server.
func (s *OsdCsiServer) Stop() {
	close(s.autoRecoverStopCh)
	s.GrpcServer.Stop()
}

func createGrpcServer(config *OsdCsiServerConfig) (*grpcserver.GrpcServer, error) {
	// create correlation interceptor
	var unaryInterceptors []grpc.UnaryServerInterceptor
	correlationInterceptor := correlation.ContextInterceptor{
		Origin: correlation.ComponentCSIDriver,
	}
	opts := make([]grpc.ServerOption, 0)
	unaryInterceptors = append(unaryInterceptors, correlationInterceptor.ContextUnaryServerInterceptor)

	// create scheduler interceptor
	switch config.SchedulerName {
	case "kubernetes":
		logrus.Infof("CSI K8s filter being added for %s scheduler", config.SchedulerName)
		ki := k8s.NewInterceptor()
		unaryInterceptors = append(unaryInterceptors, ki.SchedUnaryInterceptor)

	default:
		logrus.Infof("No CSI filter being added for %s scheduler", config.SchedulerName)
	}

	// Add interceptors
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)))

	// Create server
	gServer, err := grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "CSI 1.7",
		Net:     config.Net,
		Address: config.Address,
		Opts:    opts,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create CSI server: %v", err)
	}

	return gServer, nil
}

func autoSocketRecover(s *OsdCsiServer, stopCh chan struct{}) error {
	socketPath := s.Address()
	ticker := time.NewTicker(csiSocketCheckInterval)

	// Start checking for CSI socket delete
	for {
		select {
		case <-stopCh:
			return nil
		case <-ticker.C:
		}

		// Check if socket deleted
		_, err := os.Stat(socketPath)
		if err == nil {
			continue
		}

		logrus.Infof("Detected CSI socket deleted at path %s. Stopping CSI gRPC server", socketPath)
		s.GrpcServer.Stop()

		// Re-create gRPC server
		gServer, err := createGrpcServer(s.config)
		if err != nil {
			logrus.Errorf("failed to re-create gRPC server: %v. Retrying in %s...", err, csiSocketCheckInterval)
			continue
		}
		s.GrpcServer = gServer

		// Start server
		logrus.Infof("Restarting CSI gRPC server at %s", socketPath)
		if err := s.Start(); err != nil {
			logrus.Errorf("CSI server failed to auto-recover after socket deletion: %v. Retrying in %s...", err, csiSocketCheckInterval)
			continue
		}

		// Exit for next process to start
		return nil
	}
}

// adjustFinalErrors adjusts certain gRPC status to make CSI callers
// (csi-provisioner, kubelet, etc) retry instead of being marked as a failure.
// See https://github.com/kubernetes/kubernetes/blob/64ed9145452d2d1d324d2437566f1ea1ce76f226/pkg/volume/csi/csi_client.go#L718-L724
func adjustFinalErrors(err error) error {
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch grpcStatus.Code() {
	case codes.Internal:
		// gRPC: For cases where the CSI Driver cannot talk to the SDK, return unavailable.
		// this can occur when the SDK server is starting or offline intermittently.
		return status.New(codes.Unavailable, grpcStatus.Message()).Err()
	}

	return err
}
