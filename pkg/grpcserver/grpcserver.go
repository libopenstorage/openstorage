/*
Package grpcserver is a generic gRPC server manager
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
package grpcserver

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcServerConfig provides the configuration to the
// the gRPC server created by NewGrpcServer()
type GrpcServerConfig struct {
	Name    string
	Net     string
	Address string
}

// GrpcServer is a server manager for gRPC implementations
type GrpcServer struct {
	name     string
	listener net.Listener
	server   *grpc.Server
	wg       sync.WaitGroup
	running  bool
	lock     sync.Mutex
}

// New creates a gRPC server on the specified port and transport.
func New(config *GrpcServerConfig) (*GrpcServer, error) {
	if nil == config {
		return nil, fmt.Errorf("Configuration must be provided")
	}
	if len(config.Name) == 0 {
		return nil, fmt.Errorf("Name of server must be provided")
	}
	if len(config.Address) == 0 {
		return nil, fmt.Errorf("Address must be provided")
	}
	if len(config.Net) == 0 {
		return nil, fmt.Errorf("Net must be provided")
	}

	l, err := net.Listen(config.Net, config.Address)
	if err != nil {
		return nil, fmt.Errorf("Unable to setup server: %s", err.Error())
	}

	return &GrpcServer{
		name:     config.Name,
		listener: l,
	}, nil
}

// Start is used to start the server.
// It will return an error if the server is already runnig.
func (s *GrpcServer) Start(register func(grpcServer *grpc.Server)) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.running {
		return fmt.Errorf("Server already running")
	}

	s.server = grpc.NewServer()
	register(s.server)

	// Start listening for requests
	s.startGrpcService()
	return nil
}

// StartWithServer is used to start the server.
// It will return an error if the server is already runnig.
func (s *GrpcServer) StartWithServer(server func() *grpc.Server) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if server == nil {
		return fmt.Errorf("Server function has not been defined")
	}

	if s.running {
		return fmt.Errorf("Server already running")
	}

	s.server = server()

	// Start listening for requests
	s.startGrpcService()
	return nil
}

// Lock must have been taken
func (s *GrpcServer) startGrpcService() {
	// Start listening for requests
	reflection.Register(s.server)
	logrus.Infof("%s gRPC Server ready on %s", s.name, s.Address())
	waitForServer := make(chan bool)
	s.goServe(waitForServer)
	<-waitForServer
	s.running = true
}

// Stop is used to stop the gRPC server.
// It can be called multiple times. It does nothing if the server
// has already been stopped.
func (s *GrpcServer) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.running {
		return
	}

	s.server.Stop()
	s.wg.Wait()
	s.running = false
}

// Address returns the address of the server which can be
// used by clients to connect.
func (s *GrpcServer) Address() string {
	return s.listener.Addr().String()
}

// IsRunning returns true if the server is currently running
func (s *GrpcServer) IsRunning() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.running
}

func (s *GrpcServer) goServe(started chan<- bool) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		started <- true
		err := s.server.Serve(s.listener)
		if err != nil {
			logrus.Fatalf("ERROR: Unable to start %s gRPC server: %s\n",
				s.name,
				err.Error())
		}
	}()
}
