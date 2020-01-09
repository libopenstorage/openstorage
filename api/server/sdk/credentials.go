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

	"github.com/golang/protobuf/jsonpb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CredentialServer is an implementation of the gRPC OpenStorageCredential interface
type CredentialServer struct {
	server serverAccessor
}

func (s *CredentialServer) driver(ctx context.Context) volume.VolumeDriver {
	return s.server.driver(ctx)
}

// Create method creates credentials
func (s *CredentialServer) Create(
	ctx context.Context,
	req *api.SdkCredentialCreateRequest,
) (*api.SdkCredentialCreateResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply a name")
	} else if aws := req.GetAwsCredential(); aws != nil {
		return s.awsCreate(ctx, req, aws)
	} else if azure := req.GetAzureCredential(); azure != nil {
		return s.azureCreate(ctx, req, azure)
	} else if google := req.GetGoogleCredential(); google != nil {
		return s.googleCreate(ctx, req, google)
	}
	return nil, status.Error(codes.InvalidArgument, "Unknown credential type")

}

func (s *CredentialServer) awsCreate(
	ctx context.Context,
	req *api.SdkCredentialCreateRequest,
	aws *api.SdkAwsCredentialRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if len(aws.GetAccessKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Access Key")
	}

	if len(aws.GetSecretKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Secret Key")
	}

	if len(aws.GetRegion()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Region Key")
	}

	if len(aws.GetEndpoint()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Endpoint Key")
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = req.GetEncryptionKey()
	params[api.OptCredBucket] = req.GetBucket()
	params[api.OptCredRegion] = aws.GetRegion()
	params[api.OptCredEndpoint] = aws.GetEndpoint()
	params[api.OptCredAccessKey] = aws.GetAccessKey()
	params[api.OptCredSecretKey] = aws.GetSecretKey()
	params[api.OptCredDisableSSL] = fmt.Sprintf("%v", aws.GetDisableSsl())
	params[api.OptCredDisablePathStyle] = fmt.Sprintf("%v", aws.GetDisablePathStyle())
	params[api.OptCredProxy] = fmt.Sprintf("%v", req.GetUseProxy())

	uuid, err := s.create(ctx, req, params)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create aws credentials: %v",
			err.Error())
	}

	err = validateAndDeleteIfInvalid(ctx, s, uuid)

	if err != nil {
		return nil, err
	}
	return &api.SdkCredentialCreateResponse{CredentialId: uuid}, nil
}

func (s *CredentialServer) azureCreate(
	ctx context.Context,
	req *api.SdkCredentialCreateRequest,
	azure *api.SdkAzureCredentialRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if len(azure.GetAccountKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Account Key")
	}

	if len(azure.GetAccountName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Account name")
	}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = req.GetEncryptionKey()
	params[api.OptCredBucket] = req.GetBucket()
	params[api.OptCredAzureAccountKey] = azure.GetAccountKey()
	params[api.OptCredAzureAccountName] = azure.GetAccountName()
	params[api.OptCredProxy] = fmt.Sprintf("%v", req.GetUseProxy())

	uuid, err := s.create(ctx, req, params)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create Azure credentials: %v",
			err.Error())
	}

	err = validateAndDeleteIfInvalid(ctx, s, uuid)

	if err != nil {
		return nil, err
	}
	return &api.SdkCredentialCreateResponse{CredentialId: uuid}, nil
}

func (s *CredentialServer) googleCreate(
	ctx context.Context,
	req *api.SdkCredentialCreateRequest,
	google *api.SdkGoogleCredentialRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if len(google.GetJsonKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply JSON Key")
	}

	if len(google.GetProjectId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Project ID")
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = req.GetEncryptionKey()
	params[api.OptCredBucket] = req.GetBucket()
	params[api.OptCredGoogleProjectID] = google.GetProjectId()
	params[api.OptCredGoogleJsonKey] = google.GetJsonKey()
	params[api.OptCredProxy] = fmt.Sprintf("%v", req.GetUseProxy())

	uuid, err := s.create(ctx, req, params)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create Google credentials: %v",
			err.Error())
	}

	err = validateAndDeleteIfInvalid(ctx, s, uuid)

	if err != nil {
		return nil, err
	}

	return &api.SdkCredentialCreateResponse{CredentialId: uuid}, nil
}

// Validate validates a specified Credential.
func (s *CredentialServer) Validate(
	ctx context.Context,
	req *api.SdkCredentialValidateRequest,
) (*api.SdkCredentialValidateResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credentials uuid")
	}

	// Check ownership
	_, err := s.Inspect(ctx, &api.SdkCredentialInspectRequest{
		CredentialId: req.GetCredentialId(),
	})
	if err != nil {
		return nil, err
	}

	err = s.driver(ctx).CredsValidate(req.GetCredentialId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to validate credentials: %v",
			err.Error())
	}
	return &api.SdkCredentialValidateResponse{}, nil

}

// Delete deletes a specified credential
func (s *CredentialServer) Delete(
	ctx context.Context,
	req *api.SdkCredentialDeleteRequest,
) (*api.SdkCredentialDeleteResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credentials uuid")
	}

	// Check ownership
	resp, err := s.Inspect(ctx, &api.SdkCredentialInspectRequest{
		CredentialId: req.GetCredentialId(),
	})
	// This checks at least for READ access type to credential
	if err != nil {
		return nil, err
	}
	// This checks for admin access type to credential to be able to delete it
	if !resp.GetOwnership().IsPermittedByContext(ctx, api.Ownership_Admin) {
		return nil,
			status.Errorf(
				codes.PermissionDenied,
				"Only admin access type to credential is allowed to delete %v",
				req.GetCredentialId())
	}

	err = s.driver(ctx).CredsDelete(req.GetCredentialId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to  delete credentials: %v",
			err.Error())
	}

	return &api.SdkCredentialDeleteResponse{}, nil
}

// Enumerate returns a list credentials ids
func (s *CredentialServer) Enumerate(
	ctx context.Context,
	req *api.SdkCredentialEnumerateRequest,
) (*api.SdkCredentialEnumerateResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	credList, err := s.driver(ctx).CredsEnumerate()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to enumerate credentials AWS: %v",
			err.Error())
	}

	ids := make([]string, 0)
	for credId, cred := range credList {
		if s.isPermitted(ctx, api.Ownership_Read, cred) {
			ids = append(ids, credId)
		}
	}

	return &api.SdkCredentialEnumerateResponse{
		CredentialIds: ids,
	}, nil

}

// Inspect returns information about credential id
func (s *CredentialServer) Inspect(
	ctx context.Context,
	req *api.SdkCredentialInspectRequest,
) (*api.SdkCredentialInspectResponse, error) {
	if s.driver(ctx) == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide a credential id")
	}

	credList, err := s.driver(ctx).CredsEnumerate()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to enumerate credentials: %v",
			err.Error())
	}

	credUUID := ""
	for k, v := range credList {
		if k == req.GetCredentialId() {
			credUUID = k
			break
		}
		cred, ok := v.(map[string]interface{})
		if ok {
			name, _ := cred[api.OptCredName]
			if name == req.GetCredentialId() {
				credUUID = k
				break
			}
		}
	}
	if credUUID == "" {
		return nil, status.Errorf(codes.NotFound, "Credential id %s not found", req.GetCredentialId())
	}
	info, ok := credList[credUUID].(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to get credential id information")
	}

	// Check ownership
	if !s.isPermitted(ctx, api.Ownership_Read, info) {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied to %s", req.GetCredentialId())
	}

	credName, ok := info[api.OptCredName].(string)
	if !ok {
		// The code to support names may not be available
		credName = ""
	}
	bucket, ok := info[api.OptCredBucket].(string)
	if !ok {
		// The code to support bucket may not be available
		bucket = ""
	}

	// Get ownership
	ownership, err := s.getOwnershipFromCred(info)
	if err != nil {
		return nil, err
	}

	resp := &api.SdkCredentialInspectResponse{
		CredentialId: req.GetCredentialId(),
		Name:         credName,
		Bucket:       bucket,
		Ownership:    ownership,
	}
	useProxy, ok := info[api.OptCredProxy].(string)
	if !ok {
		// older format creds
		resp.UseProxy = false
	}
	resp.UseProxy = useProxy == "true"

	switch info[api.OptCredType] {
	case "s3":
		accessKey, ok := info[api.OptCredAccessKey].(string)
		if !ok {
			return nil, status.Error(codes.Internal, "Unable to parse accessKey")
		}
		endpoint, ok := info[api.OptCredEndpoint].(string)
		if !ok {
			return nil, status.Error(codes.Internal, "Unable to parse endpoint")
		}
		region, ok := info[api.OptCredRegion].(string)
		if !ok {
			return nil, status.Error(codes.Internal, "Unable to parse region")
		}
		disableSsl, ok := info[api.OptCredDisableSSL].(string)
		if !ok {
			return nil, status.Error(codes.Internal, "Unable to parse disabling ssl was requested")
		}
		disablePathStyle, ok := info[api.OptCredDisablePathStyle].(string)
		if !ok {
			// older format creds
			disablePathStyle = "false"
		}

		resp.CredentialType = &api.SdkCredentialInspectResponse_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialResponse{
				AccessKey:        accessKey,
				Endpoint:         endpoint,
				Region:           region,
				DisableSsl:       disableSsl == "true",
				DisablePathStyle: disablePathStyle == "true",
			},
		}
	case "azure":
		accountName, ok := info[api.OptCredAzureAccountName].(string)
		if !ok {
			return nil, status.Error(codes.Internal, "Unable to parse account name")
		}

		resp.CredentialType = &api.SdkCredentialInspectResponse_AzureCredential{
			AzureCredential: &api.SdkAzureCredentialResponse{
				AccountName: accountName,
			},
		}
	case "google":
		projectId, ok := info[api.OptCredGoogleProjectID].(string)
		if !ok {
			return nil, status.Error(codes.Internal, "Unable to parse project id")
		}
		resp.CredentialType = &api.SdkCredentialInspectResponse_GoogleCredential{
			GoogleCredential: &api.SdkGoogleCredentialResponse{
				ProjectId: projectId,
			},
		}
	default:
		return nil, status.Errorf(
			codes.Internal,
			"Received unknown credential type of %s",
			info[api.OptCredType])
	}

	return resp, nil
}

func (s *CredentialServer) create(
	ctx context.Context,
	req *api.SdkCredentialCreateRequest,
	params map[string]string) (string, error) {
	if params == nil || req == nil {
		return "", fmt.Errorf("params and/or request is nil and cannot create credentials")
	}

	// Add user as owner
	ownership := api.OwnershipSetUsernameFromContext(ctx, req.GetOwnership())
	if ownership != nil {
		// Encode ownership in params
		m := jsonpb.Marshaler{OrigName: true}
		ownershipString, err := m.MarshalToString(ownership)
		if err != nil {
			return "", fmt.Errorf("failed to marshal ownership: %v", err)
		}
		params[api.OptCredOwnership] = ownershipString
	}

	return s.driver(ctx).CredsCreate(params)
}

func (s *CredentialServer) getOwnershipFromCred(cred interface{}) (*api.Ownership, error) {
	info, ok := cred.(map[string]interface{})
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to get credential id information")
	}

	// Get ownership
	var ownership *api.Ownership
	ownershipString, ok := info[api.OptCredOwnership].(string)
	if ok {
		if len(ownershipString) == 0 {
			return nil, nil
		}
		ownership = &api.Ownership{}
		err := jsonpb.UnmarshalString(ownershipString, ownership)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Failed to retreive ownership from credential object: %v", err)
		}
	}
	return ownership, nil
}

func (s CredentialServer) isPermitted(
	ctx context.Context,
	accessType api.Ownership_AccessType,
	cred interface{},
) bool {
	ownership, err := s.getOwnershipFromCred(cred)
	if err != nil {
		return false
	}

	// If ownership is missing then it is also public
	if ownership == nil || ownership.IsPublic(accessType) {
		return true
	}

	if userinfo, ok := auth.NewUserInfoFromContext(ctx); ok {
		return ownership.IsPermitted(userinfo, accessType)
	}

	// Auth is not enabled if there is no user context
	return true
}

func validateAndDeleteIfInvalid(ctx context.Context, s *CredentialServer, uuid string) error {
	// Validate if the credentials provided were correct or not
	req := &api.SdkCredentialValidateRequest{CredentialId: uuid}

	validateErr := s.driver(ctx).CredsValidate(req.GetCredentialId())

	if validateErr != nil {
		deleteCred := &api.SdkCredentialDeleteRequest{CredentialId: uuid}
		err := s.driver(ctx).CredsDelete(deleteCred.GetCredentialId())

		if err != nil {
			return status.Errorf(
				codes.Internal,
				"failed to delete invalid Google credentials: %v",
				err.Error())
		}

		return status.Errorf(
			codes.PermissionDenied,
			"credentials could not be validated: %v",
			validateErr.Error())
	}

	return nil
}
