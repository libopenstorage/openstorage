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
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/correlation"
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
	clogger *logrus.Logger
)

const (
	connCleanupInterval = 15 * time.Minute
	connIdleConnLength  = 30 * time.Minute
)

func init() {
	clogger = correlation.NewPackageLogger(correlation.ComponentCSIDriver)
}

// OsdCsiServerConfig provides the configuration to the
// the gRPC CSI server created by NewOsdCsiServer()
type OsdCsiServerConfig struct {
	Net        string
	Address    string
	DriverName string
	Cluster    cluster.Cluster
	SdkUds     string
	SdkPort    string

	// Name to be reported back to the CO. If not provided,
	// the name will be in the format of <driver>.openstorage.org
	CsiDriverName string

	// EnableInlineVolumes decides whether or not we will allow
	// creation of inline volumes.
	EnableInlineVolumes bool
}

// TimedSDKConn represents a gRPC connection and the last time it was used
type TimedSDKConn struct {
	Conn      *grpc.ClientConn
	LastUsage time.Time
}

// OsdCsiServer is a OSD CSI compliant server which
// proxies CSI requests for a single specific driver
type OsdCsiServer struct {
	csi.ControllerServer
	csi.NodeServer
	csi.IdentityServer

	*grpcserver.GrpcServer
	specHandler          spec.SpecHandler
	driver               volume.VolumeDriver
	cluster              cluster.Cluster
	sdkUds               string
	sdkPort              string
	conn                 *grpc.ClientConn
	connMap              map[string]*TimedSDKConn
	nextCreateNodeNumber int
	mu                   sync.Mutex
	csiDriverName        string
	allowInlineVolumes   bool
	stopCleanupCh        chan bool
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

	// Add correlation interceptor
	correlationInterceptor := correlation.ContextInterceptor{
		Origin: correlation.ComponentCSIDriver,
	}
	opts := make([]grpc.ServerOption, 0)
	opts = append(opts, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			correlationInterceptor.ContextUnaryServerInterceptor,
		)))

	// Create server
	gServer, err := grpcserver.New(&grpcserver.GrpcServerConfig{
		Name:    "CSI 1.6",
		Net:     config.Net,
		Address: config.Address,
		Opts:    opts,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create CSI server: %v", err)
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
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get all nodes and sort them
	nodesResp, err := s.cluster.Enumerate()
	if err != nil {
		return nil, err
	}
	if len(nodesResp.Nodes) < 1 {
		return nil, errors.New("cluster nodes for remote connection not found")
	}
	sort.Slice(nodesResp.Nodes, func(i, j int) bool {
		return nodesResp.Nodes[i].Id < nodesResp.Nodes[j].Id
	})

	// Clean up connections for missing nodes
	s.cleanupMissingNodeConnections(ctx, nodesResp.Nodes)

	// Get target node info and set next round robbin node.
	// nextNode is always lastNode + 1 mod (numOfNodes), to loop back to zero
	var targetNodeNumber int
	if s.nextCreateNodeNumber != 0 {
		targetNodeNumber = s.nextCreateNodeNumber
	}
	targetNodeEndpoint := nodesResp.Nodes[targetNodeNumber].MgmtIp
	s.nextCreateNodeNumber = (targetNodeNumber + 1) % len(nodesResp.Nodes)

	// Get conn for this node, otherwise create new conn
	if len(s.connMap) == 0 {
		s.connMap = make(map[string]*TimedSDKConn)
	}
	if s.connMap[targetNodeEndpoint] == nil {
		var err error
		clogger.WithContext(ctx).Infof("Round-robin connecting to node %v - %s:%s", targetNodeNumber, targetNodeEndpoint, s.sdkPort)
		remoteConn, err := grpcserver.ConnectWithTimeout(
			fmt.Sprintf("%s:%s", targetNodeEndpoint, s.sdkPort),
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
			}, 10*time.Second)
		if err != nil {
			return nil, err
		}

		s.connMap[targetNodeEndpoint] = &TimedSDKConn{
			Conn: remoteConn,
		}
	}

	// Keep track of when this conn was last accessed
	clogger.WithContext(ctx).Infof("Using remote connection to SDK node %v - %s:%s", targetNodeNumber, targetNodeEndpoint, s.sdkPort)
	s.connMap[targetNodeEndpoint].LastUsage = time.Now()
	return s.connMap[targetNodeEndpoint].Conn, nil
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
	return s.GrpcServer.Start(func(grpcServer *grpc.Server) {
		go s.cleanupConnections()

		csi.RegisterIdentityServer(grpcServer, s)
		csi.RegisterControllerServer(grpcServer, s)
		csi.RegisterNodeServer(grpcServer, s)
	})
}

// Start is used to stop the server.
func (s *OsdCsiServer) Stop() {
	if s.stopCleanupCh != nil {
		close(s.stopCleanupCh)
	}
	s.GrpcServer.Stop()
}

func (s *OsdCsiServer) cleanupConnections() {
	s.stopCleanupCh = make(chan bool)
	ticker := time.NewTicker(connCleanupInterval)

	// Check every so often and delete/close connections
	for {
		select {
		case <-s.stopCleanupCh:
			ticker.Stop()
			return

		case _ = <-ticker.C:
			ctx := correlation.WithCorrelationContext(context.Background(), correlation.ComponentCSIDriver)

			// Anonymous function for using defer to unlock mutex
			func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				clogger.Infof("Cleaning up open gRPC connections for CSI distributed provisioning")

				// Clean all expired connections
				numConnsClosed := 0
				for ip, timedConn := range s.connMap {
					expiryTime := timedConn.LastUsage.Add(connIdleConnLength)

					// Connection has expired after 1hr of no usage.
					// Close connection and remove from connMap
					if expiryTime.Before(time.Now()) {
						clogger.Infof("SDK gRPC connection to %s is has expired after %v minutes of no usage. Closing this connection", ip, connIdleConnLength.Minutes())
						if err := timedConn.Conn.Close(); err != nil {
							clogger.Errorf("failed to close connection to %s: %v", ip, timedConn.Conn)
						}
						delete(s.connMap, ip)
						numConnsClosed++
					}
				}

				// Get all nodes and cleanup conns for missing/deprovisioned nodes
				nodesResp, err := s.cluster.Enumerate()
				if err != nil {
					clogger.Errorf("failed to get all nodes for connection cleanup: %v", err)
					return
				}
				if len(nodesResp.Nodes) < 1 {
					clogger.Errorf("no nodes available to cleanup: %v", err)
					return
				}
				s.cleanupMissingNodeConnections(ctx, nodesResp.Nodes)

				clogger.Infof("Cleaned up %v connections for CSI distributed provisioning. %v connections remaining", numConnsClosed, len(s.connMap))
			}()
		}
	}
}

func (s *OsdCsiServer) cleanupMissingNodeConnections(ctx context.Context, nodes []*api.Node) {
	nodesMap := make(map[string]bool)
	for _, node := range nodes {
		nodesMap[node.MgmtIp] = true
	}
	for ip, timedConn := range s.connMap {
		if ok := nodesMap[ip]; !ok {
			// If key in connmap is not in current nodes, close and remove it
			if err := timedConn.Conn.Close(); err != nil {
				clogger.WithContext(ctx).Errorf("failed to close conn to %s: %v", ip, err)
			}
			delete(s.connMap, ip)
		}
	}
}
