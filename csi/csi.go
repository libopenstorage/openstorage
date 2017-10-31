/*
CSI Interface for OSD
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
	"net"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.pedge.io/dlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type OsdCsiServerConfig struct {
	Listener net.Listener
}

type OsdCsiServer struct {
	listener net.Listener
	server   *grpc.Server
	wg       sync.WaitGroup
}

func NewOsdCsiServer(config *OsdCsiServerConfig) *OsdCsiServer {
	return &OsdCsiServer{
		listener: config.Listener,
	}
}

func (s *OsdCsiServer) Start() error {
	s.server = grpc.NewServer()

	csi.RegisterIdentityServer(s.server, s)
	reflection.Register(s.server)

	// Start listening for requests
	dlog.Infof("CSI Server ready on %s", s.Address())
	s.goServe()
	return nil
}

func (s *OsdCsiServer) Stop() {
	s.server.Stop()
	s.wg.Wait()
}

func (s *OsdCsiServer) Address() string {
	return s.listener.Addr().String()
}

func (s *OsdCsiServer) goServe() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		err := s.server.Serve(s.listener)
		if err != nil {
			dlog.Fatalf("ERROR: Unable to start gRPC server: %s\n", err.Error())
		}
	}()
}
