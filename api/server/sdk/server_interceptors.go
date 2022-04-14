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
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// Metedata context key where the token is found.
	// This key must be used by the caller as the key for the token in
	// the metedata of the context. The generated Rest Gateway also uses this
	// key as the location of the raw token coming from the standard REST
	// header: Authorization: bearer <adaf0sdfsd...token>
	ContextMetadataTokenKey = "bearer"
)

// This interceptor provides a way to lock out any unary calls while we adjust the server
func (s *sdkGrpcServer) rwlockUnaryIntercepter(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return handler(ctx, req)
}

// This interceptor provides a way to lock out any stream calls while we adjust the server
func (s *sdkGrpcServer) rwlockStreamIntercepter(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return handler(srv, stream)
}

// Authenticate user and add authorization information back in the context
func (s *sdkGrpcServer) auth(ctx context.Context) (context.Context, error) {
	var token string
	var err error

	// Audit log
	log := correlation.NewFunctionLogger(ctx)
	log.Out = s.auditLogOutput
	auditLogWarningf := func(c codes.Code, format string, a ...interface{}) error {
		log.WithContext(ctx).WithFields(logrus.Fields{
			"method": "Authentication",
			"code":   c.String(),
		}).Warningf(format, a...)
		return status.Errorf(c, format, a...)
	}

	// guest call attempted, add system.guest user
	if auth.IsGuest(ctx) {
		return auth.ContextSaveUserInfo(ctx, auth.NewGuestUser()), nil
	}

	// Obtain token from metadata in the context
	token, err = grpc_auth.AuthFromMD(ctx, ContextMetadataTokenKey)
	if err != nil {
		return nil, auditLogWarningf(codes.Unauthenticated, "Invalid or missing authentication token")
	}

	// Determine issuer
	issuer, err := auth.TokenIssuer(token)
	if err != nil {
		return nil, auditLogWarningf(codes.Unauthenticated, "Unable to obtain token issuer from authorization token")
	}

	// Authenticate user
	if authenticator, ok := s.config.Security.Authenticators[issuer]; ok {
		var claims *auth.Claims
		claims, err = authenticator.AuthenticateToken(ctx, token)
		if err == nil {
			// Add authorization information back into the context so that other
			// functions can get access to this information.
			// If this is in the context is how functions will know that security is enabled.
			ctx = auth.ContextSaveUserInfo(ctx, &auth.UserInfo{
				Username: authenticator.Username(claims),
				Claims:   *claims,
			})
			return ctx, nil
		} else {
			return nil, auditLogWarningf(codes.PermissionDenied, "Unable to authenticate token")
		}
	} else {
		return nil, auditLogWarningf(codes.Unauthenticated, "%s is not a trusted issuer", issuer)
	}
}

func (s *sdkGrpcServer) loggerInterceptor(ctx context.Context, handler func() error, fullMethod string) error {
	reqid := uuid.New()
	log := correlation.NewFunctionLogger(ctx)
	log.Out = s.accessLogOutput
	logger := log.WithContext(ctx).WithFields(logrus.Fields{
		"method": fullMethod,
		"reqid":  reqid,
	})

	logger.Info("Start")
	ts := time.Now()
	err := handler()
	duration := time.Now().Sub(ts)
	if err != nil {
		logger.WithFields(logrus.Fields{"duration": duration}).Infof("Failed: %v", err)
	} else {
		logger.WithFields(logrus.Fields{"duration": duration}).Info("Successful")
	}

	return err
}

func (s *sdkGrpcServer) loggerServerUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var i interface{}
	var err error

	err = s.loggerInterceptor(ctx, func() error {
		i, err = handler(ctx, req)
		return err
	}, info.FullMethod)

	return i, err
}

func (s *sdkGrpcServer) loggerServerStreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	var err error

	return s.loggerInterceptor(stream.Context(), func() error {
		err = handler(srv, stream)
		if err != nil {
			return err
		}
		return nil
	}, info.FullMethod)
}

func (s *sdkGrpcServer) authorizationInterceptor(
	ctx context.Context,
	handler func() error,
	fullMethod string,
) error {
	userinfo, ok := auth.NewUserInfoFromContext(ctx)
	if !ok {
		return status.Error(
			codes.Internal,
			"Unable to authorize user because token is missing from context")
	}
	claims := &userinfo.Claims

	// Get method and API
	reqService, reqAPI := grpcserver.GetMethodInformation(api.SdkRootPath, fullMethod)

	// Setup auditor log
	log := correlation.NewFunctionLogger(ctx)
	log.Out = s.auditLogOutput
	logger := log.WithFields(logrus.Fields{
		"username": userinfo.Username,
		"subject":  claims.Subject,
		"name":     claims.Name,
		"email":    claims.Email,
		"roles":    claims.Roles,
		"groups":   claims.Groups,
		"method":   fmt.Sprintf("%s.%s", reqService, reqAPI),
	}).WithContext(ctx)

	// Authorize
	if err := s.roleServer.Verify(ctx, claims.Roles, fullMethod); err != nil {
		logger.Warning("Access denied")
		if auth.IsGuest(ctx) {
			return status.Errorf(
				codes.PermissionDenied,
				"Access denied without authentication token")
		}

		return status.Errorf(
			codes.PermissionDenied,
			"Access to %s denied: %v",
			fullMethod, err)
	}

	// Check if we have been denied
	err := handler()
	if err != nil {
		if gErr, ok := status.FromError(err); ok {
			if gErr.Code() == codes.PermissionDenied {
				logger.Warningf("Access denied: %v", err)
				return err
			}
		}
	}

	// Log
	logger.Info("Authorized")

	return err
}

func (s *sdkGrpcServer) authorizationServerUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var i interface{}
	var err error

	err = s.authorizationInterceptor(ctx, func() error {
		i, err = handler(ctx, req)
		return err
	}, info.FullMethod)

	// Execute the command
	return i, err
}

func (s *sdkGrpcServer) authorizationServerStreamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	return s.authorizationInterceptor(stream.Context(), func() error {
		return handler(srv, stream)
	}, info.FullMethod)
}
