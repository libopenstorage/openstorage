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
	"time"

	"github.com/libopenstorage/openstorage/pkg/auth"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
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
	// Obtain token from metadata in the context
	token, err := grpc_auth.AuthFromMD(ctx, ContextMetadataTokenKey)
	if err != nil {
		return nil, err
	}

	// Determine issuer
	issuer, err := auth.TokenIssuer(token)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "Unable to obtain issuer from token: %v", err)
	}

	// Authenticate user
	if authenticator, ok := s.config.Security.Authenticators[issuer]; ok {
		var claims *auth.Claims
		claims, err = authenticator.AuthenticateToken(ctx, token)
		if err == nil {
			// Add authorization information back into the context so that other
			// functions can get access to this information
			ctx = auth.ContextSaveUserInfo(ctx, &auth.UserInfo{
				Username: authenticator.Username(claims),
				Claims:   *claims,
			})
			return ctx, nil
		} else {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
	} else {
		return nil, status.Errorf(codes.PermissionDenied, "%s is not a trusted issuer", issuer)
	}
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
	userinfo, err := auth.NewUserInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}
	claims := &userinfo.Claims

	// Setup auditor log
	log := logrus.New()
	log.Out = s.auditLogOutput
	logger := log.WithFields(logrus.Fields{
		"username": userinfo.Username,
		"subject":  claims.Subject,
		"name":     claims.Name,
		"email":    claims.Email,
		"roles":    claims.Roles,
		"groups":   claims.Groups,
		"method":   info.FullMethod,
	})

	// Authorize
	if err := s.roleServer.Verify(ctx, claims.Roles, info.FullMethod); err != nil {
		logger.Infof("Access denied")
		return nil, status.Errorf(
			codes.PermissionDenied,
			"Access to %s denied: %v",
			info.FullMethod, err)
	}

	logger.Info("Authorized")
	return handler(ctx, req)
}
