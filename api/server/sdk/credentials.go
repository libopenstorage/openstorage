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

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CredentialServer is an implementation of the gRPC OpenStorageCredential interface
type CredentialServer struct {
	driver volume.VolumeDriver
}

// Create method creates credentials
func (s *CredentialServer) Create(
	ctx context.Context,
	req *api.SdkCredentialCreateRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if aws := req.GetAwsCredential(); aws != nil {
		return s.awsCreate(ctx, aws)
	} else if azure := req.GetAzureCredential(); azure != nil {
		return s.azureCreate(ctx, azure)
	} else if google := req.GetGoogleCredential(); google != nil {
		return s.googleCreate(ctx, google)
	}
	return nil, status.Error(codes.InvalidArgument, "Unknown credential type")

}

func (s *CredentialServer) awsCreate(
	ctx context.Context,
	req *api.SdkAwsCredentialRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if len(req.GetAccessKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Access Key")
	}

	if len(req.GetSecretKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Secret Key")
	}

	if len(req.GetRegion()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Region Key")
	}

	if len(req.GetEndpoint()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Endpoint Key")
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredRegion] = req.GetRegion()
	params[api.OptCredEndpoint] = req.GetEndpoint()
	params[api.OptCredAccessKey] = req.GetAccessKey()
	params[api.OptCredSecretKey] = req.GetSecretKey()

	uuid, err := s.driver.CredsCreate(params)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create aws credentials: %v",
			err.Error())
	}

	err = validateAndDeleteIfInvalid(s, uuid)

	if err != nil {
		return nil, err
	}
	return &api.SdkCredentialCreateResponse{CredentialId: uuid}, nil
}

func (s *CredentialServer) azureCreate(
	ctx context.Context,
	req *api.SdkAzureCredentialRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if len(req.GetAccountKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Account Key")
	}

	if len(req.GetAccountName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Account name")
	}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredAzureAccountKey] = req.GetAccountKey()
	params[api.OptCredAzureAccountName] = req.GetAccountName()

	uuid, err := s.driver.CredsCreate(params)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create Azure credentials: %v",
			err.Error())
	}

	err = validateAndDeleteIfInvalid(s, uuid)

	if err != nil {
		return nil, err
	}
	return &api.SdkCredentialCreateResponse{CredentialId: uuid}, nil
}

func (s *CredentialServer) googleCreate(
	ctx context.Context,
	req *api.SdkGoogleCredentialRequest,
) (*api.SdkCredentialCreateResponse, error) {

	if len(req.GetJsonKey()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply JSON Key")
	}

	if len(req.GetProjectId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must supply Project ID")
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredGoogleProjectID] = req.GetProjectId()
	params[api.OptCredGoogleJsonKey] = req.GetJsonKey()

	uuid, err := s.driver.CredsCreate(params)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to create Google credentials: %v",
			err.Error())
	}

	err = validateAndDeleteIfInvalid(s, uuid)

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

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credentials uuid")
	}

	validateReq := &api.SdkCredentialValidateRequest{CredentialId: req.GetCredentialId()}

	err := s.driver.CredsValidate(validateReq.GetCredentialId())

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

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide credentials uuid")
	}

	err := s.driver.CredsDelete(req.GetCredentialId())
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

	credList, err := s.driver.CredsEnumerate()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to enumerate credentials AWS: %v",
			err.Error())
	}

	ids := make([]string, len(credList))
	i := 0
	for id := range credList {
		ids[i] = id
		i++
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

	if len(req.GetCredentialId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Must provide a credential id")
	}

	credList, err := s.driver.CredsEnumerate()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"Unable to enumerate credentials: %v",
			err.Error())
	}

	val, ok := credList[req.GetCredentialId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "Credential id %s not found", req.GetCredentialId())
	}
	info, ok := val.(map[string]string)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to get credential id information")
	}

	switch info[api.OptCredType] {
	case "s3":
		return &api.SdkCredentialInspectResponse{
			CredentialType: &api.SdkCredentialInspectResponse_AwsCredential{
				AwsCredential: &api.SdkAwsCredentialResponse{
					CredentialId: req.GetCredentialId(),
					AccessKey:    info[api.OptCredAccessKey],
					Endpoint:     info[api.OptCredEndpoint],
					Region:       info[api.OptCredRegion],
				},
			},
		}, nil
	case "azure":
		return &api.SdkCredentialInspectResponse{
			CredentialType: &api.SdkCredentialInspectResponse_AzureCredential{
				AzureCredential: &api.SdkAzureCredentialResponse{
					CredentialId: req.GetCredentialId(),
					AccountName:  info[api.OptCredAzureAccountName],
				},
			},
		}, nil
	case "google":
		return &api.SdkCredentialInspectResponse{
			CredentialType: &api.SdkCredentialInspectResponse_GoogleCredential{
				GoogleCredential: &api.SdkGoogleCredentialResponse{
					CredentialId: req.GetCredentialId(),
					ProjectId:    info[api.OptCredGoogleProjectID],
				},
			},
		}, nil
	default:
		return nil, status.Errorf(
			codes.Internal,
			"Received unknown credential type of %s",
			info[api.OptCredType])
	}

}

func validateAndDeleteIfInvalid(s *CredentialServer, uuid string) error {
	// Validate if the credentials provided were correct or not
	req := &api.SdkCredentialValidateRequest{CredentialId: uuid}

	validateErr := s.driver.CredsValidate(req.GetCredentialId())

	if validateErr != nil {
		deleteCred := &api.SdkCredentialDeleteRequest{CredentialId: uuid}
		err := s.driver.CredsDelete(deleteCred.GetCredentialId())

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
