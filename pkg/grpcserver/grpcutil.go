/*
Copyright 2017 The Kubernetes Authors.

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
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	DefaultConnectionTimeout = 1 * time.Minute
)

// GetTlsDialOptions returns the appropriate gRPC dial options to connect to a gRPC server over TLS.
// If caCertData is nil then it will use the CA from the host.
func GetTlsDialOptions(caCertData []byte) ([]grpc.DialOption, error) {
	// Read the provided CA cert from the user
	capool, err := x509.SystemCertPool()
	if err != nil || capool == nil {
		logrus.Warnf("cannot load system root certificates: %v", err)
		capool = x509.NewCertPool()
	}

	// If user provided CA cert, then append it to systemCertPool.
	if len(caCertData) != 0 {
		if !capool.AppendCertsFromPEM(caCertData) {
			return nil, fmt.Errorf("cannot parse CA certificate")
		}
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(capool, "")),
		grpc.WithUnaryInterceptor(correlation.ContextUnaryClientInterceptor),
	}
	return dialOptions, nil
}

// Connect to address by grpc
func Connect(address string, dialOptions []grpc.DialOption) (*grpc.ClientConn, error) {
	return ConnectWithTimeout(address, dialOptions, DefaultConnectionTimeout)
}

// ConnectWithTimeout to address by grpc with timeout
func ConnectWithTimeout(address string, dialOptions []grpc.DialOption, timeout time.Duration) (*grpc.ClientConn, error) {
	u, err := url.Parse(address)
	if err == nil {
		// Check if host just has an IP
		if u.Scheme == "unix" ||
			(!u.IsAbs() && net.ParseIP(address) == nil) {
			dialOptions = append(dialOptions,
				grpc.WithDialer(
					func(addr string, timeout time.Duration) (net.Conn, error) {
						return net.DialTimeout("unix", u.Path, timeout)
					}))
		}
	}

	dialOptions = append(dialOptions,
		grpc.WithBackoffMaxDelay(time.Second),
		grpc.WithBlock(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect gRPC server %s: %v", address, err)
	}

	return conn, nil
}

func AddMetadataToContext(ctx context.Context, k, v string) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = metadata.Join(md, metadata.New(map[string]string{
		k: v,
	}))
	return metadata.NewOutgoingContext(ctx, md)
}

func GetMetadataValueFromKey(ctx context.Context, k string) string {
	return metautils.ExtractIncoming(ctx).Get(k)
}

// GetMethodInformation returns the service and API of a gRPC fullmethod string.
// For example, if the full method is:
//   /openstorage.api.OpenStorage<service>/<method>
// Then, to extract the service and api we would call it as follows:
//   s, a := GetMethodInformation("openstorage.api.OpenStorage", info.FullMethod)
//      where info.FullMethod comes from the gRPC interceptor
func GetMethodInformation(constPath, fullmethod string) (service, api string) {
	parts := strings.Split(fullmethod, "/")

	if len(parts) > 1 {
		service = strings.TrimPrefix(strings.ToLower(parts[1]), strings.ToLower(constPath))
	}

	if len(parts) > 2 {
		api = strings.ToLower(parts[2])
	}

	return service, api
}
