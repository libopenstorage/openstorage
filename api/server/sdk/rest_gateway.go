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
	"mime"
	"net/http"

	"github.com/gobuffalo/packr"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
)

type sdkRestGateway struct {
	config     ServerConfig
	restPort   string
	grpcServer *sdkGrpcServer
	server     *http.Server
}

func newSdkRestGateway(config *ServerConfig, grpcServer *sdkGrpcServer) (*sdkRestGateway, error) {
	return &sdkRestGateway{
		config:     *config,
		restPort:   config.RestPort,
		grpcServer: grpcServer,
	}, nil
}

func (s *sdkRestGateway) Start() error {
	mux, err := s.restServerSetupHandlers()
	if err != nil {
		return err
	}

	// Create object here so that we can access its Close receiver.
	address := ":" + s.restPort
	s.server = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	ready := make(chan bool)
	go func() {
		ready <- true
		var err error
		if s.config.Tls != nil {
			err = s.server.ListenAndServeTLS(s.config.Tls.CertFile, s.config.Tls.KeyFile)
		} else {
			err = s.server.ListenAndServe()
		}

		if err == http.ErrServerClosed || err == nil {
			return
		} else {
			logrus.Fatalf("Unable to start SDK REST gRPC Gateway: %s\n",
				err.Error())
		}
	}()
	<-ready
	logrus.Infof("SDK gRPC REST Gateway started on port :%s", s.restPort)

	return nil
}

func (s *sdkRestGateway) Stop() {
	if err := s.server.Close(); err != nil {
		logrus.Fatalf("REST GW STOP error: %v", err)
	}
}

// restServerSetupHandlers sets up the handlers to the swagger ui and
// to the gRPC REST Gateway.
func (s *sdkRestGateway) restServerSetupHandlers() (*http.ServeMux, error) {

	// Create an HTTP server router
	mux := http.NewServeMux()

	// Swagger files using packr
	swaggerUIBox := packr.NewBox("./swagger-ui")
	swaggerJSONBox := packr.NewBox("./api")
	mime.AddExtensionType(".svg", "image/svg+xml")

	// Handler to return swagger.json
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write(swaggerJSONBox.Bytes("api.swagger.json"))
	})

	// Handler to access the swagger ui. The UI pulls the swagger
	// json file from /swagger.json
	// The link below MUST have th last '/'. It is really important.
	prefix := "/swagger-ui/"
	mux.Handle(prefix,
		http.StripPrefix(prefix, http.FileServer(swaggerUIBox)))

	// Create a router just for HTTP REST gRPC Server Gateway
	gmux := runtime.NewServeMux()

	// Connect to gRPC unix domain socket
	conn, err := grpcserver.Connect(
		s.grpcServer.Address(),
		[]grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to gRPC handler: %v", err)
	}

	// REST Gateway Handlers
	handlers := []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) (err error){
		api.RegisterOpenStorageClusterHandler,
		api.RegisterOpenStorageNodeHandler,
		api.RegisterOpenStorageVolumeHandler,
		api.RegisterOpenStorageObjectstoreHandler,
		api.RegisterOpenStorageCredentialsHandler,
		api.RegisterOpenStorageSchedulePolicyHandler,
		api.RegisterOpenStorageCloudBackupHandler,
		api.RegisterOpenStorageIdentityHandler,
		api.RegisterOpenStorageMountAttachHandler,
		api.RegisterOpenStorageAlertsHandler,
		api.RegisterOpenStorageClusterPairHandler,
		api.RegisterOpenStorageMigrateHandler,
		api.RegisterOpenStoragePolicyHandler,
	}

	// Register the REST Gateway handlers
	for _, handler := range handlers {
		err := handler(context.Background(), gmux, conn)
		if err != nil {
			return nil, err
		}
	}

	// Pass all other unhandled paths to the gRPC gateway
	mux.Handle("/", gmux)
	return mux, nil
}
