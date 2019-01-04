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
	"encoding/json"
	"time"

	sdk_auth "github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Keys to store data in gRPC context. Use these keys to retrieve
// the data from the gRPC context
type InterceptorContextkey string

const (
	// Key to store in the token claims in gRPC context
	InterceptorContextTokenKey InterceptorContextkey = "tokenclaims"

	// Metedata context key where the token is found.
	// This key must be used by the caller as the key for the token in
	// the metedata of the context. The generated Rest Gateway also uses this
	// key as the location of the raw token coming from the standard REST
	// header: Authorization: bearer <adaf0sdfsd...token>
	ContextMetadataTokenKey = "bearer"
)

// This interceptor provides a way to lock out any calls while we adjust the server
func (s *sdkGrpcServer) rwlockIntercepter(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return handler(ctx, req)
}

// Authenticate user and add authorization information back in the context
func (s *sdkGrpcServer) auth(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, ContextMetadataTokenKey)
	if err != nil {
		return nil, err
	}

	// Authenticate user
	claims, err := s.authenticator.AuthenticateToken(token)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Add authorization information back into the context so that other
	// functions can get access to this information
	ctx = context.WithValue(ctx, InterceptorContextTokenKey, claims)

	return ctx, nil
}

func (s *sdkGrpcServer) loggerServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	reqid := uuid.New()
	log := logrus.New()
	log.Out = s.accessLogOutput
	logger := log.WithFields(logrus.Fields{
		"method": info.FullMethod,
		"reqid":  reqid,
	})

	logger.Info("Start")
	ts := time.Now()
	i, err := handler(ctx, req)
	duration := time.Now().Sub(ts)
	if err != nil {
		logger.WithFields(logrus.Fields{"duration": duration}).Infof("Failed: %v", err)
	} else {
		logger.WithFields(logrus.Fields{"duration": duration}).Info("Successful")
	}

	return i, err
}

func (s *sdkGrpcServer) authorizationServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	claims, ok := ctx.Value(InterceptorContextTokenKey).(*sdk_auth.Claims)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Authorization called without token")
	}

	// Setup auditor log
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		logrus.Warningf("Unable to unmarshal claims: %v", err)
	}
	log := logrus.New()
	log.Out = s.auditLogOutput
	logger := log.WithFields(logrus.Fields{
		"name":   claims.Name,
		"email":  claims.Email,
		"role":   claims.Role,
		"claims": string(claimsJSON),
		"method": info.FullMethod,
	})

	// Authorize
	if err := s.roleServer.Verify(ctx, claims.Role, info.FullMethod); err != nil {
		logger.Infof("Access denied")
		return nil, status.Errorf(
			codes.PermissionDenied,
			"Access to %s denied: %v",
			info.FullMethod, err)
	}

	logger.Info("Authorized")
	return handler(ctx, req)
}
