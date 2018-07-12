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

// Server is an interface to a gRPC server which provides an implementation
// of an exported gRPC interface
type Server interface {
	// Start the server. If called on a running server it will return an error.
	Start() error

	// Stop the server. If called on a stopped server it will have no effect
	Stop()

	// IsRunning tell the caller if the server is currently running
	IsRunning() bool

	// Address returns the address used by clients to connect to the server
	Address() string
}
