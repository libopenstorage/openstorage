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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
)

func TestSdkAWSCredentialCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateAWSRequest{
		Credential: &api.S3Credential{
			AccessKey: "dummy-access",
			SecretKey: "dummy-secret",
			Endpoint:  "dummy-endpoint",
			Region:    "dummy-region",
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredRegion] = req.GetCredential().GetRegion()
	params[api.OptCredEndpoint] = req.GetCredential().GetEndpoint()
	params[api.OptCredAccessKey] = req.GetCredential().GetAccessKey()
	params[api.OptCredSecretKey] = req.GetCredential().GetSecretKey()

	uuid := "good-uuid"
	s.MockDriver().
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	s.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create AWS Credentials
	_, err := c.CreateForAWS(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkAWSCredentialCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateAWSRequest{
		Credential: &api.S3Credential{
			AccessKey: "dummy-access",
			SecretKey: "dummy-secret",
			Endpoint:  "dummy-endpoint",
			Region:    "dummy-region",
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredRegion] = req.GetCredential().GetRegion()
	params[api.OptCredEndpoint] = req.GetCredential().GetEndpoint()
	params[api.OptCredAccessKey] = req.GetCredential().GetAccessKey()
	params[api.OptCredSecretKey] = req.GetCredential().GetSecretKey()

	uuid := "bad-uuid"
	s.MockDriver().
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	s.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(fmt.Errorf("Invalid credentials"))

	s.MockDriver().
		EXPECT().
		CredsDelete(uuid).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.CreateForAWS(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.PermissionDenied)
	assert.Contains(t, serverError.Message(), "Invalid credentials")
}

func TestSdkAWSCredentialCreateBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateAWSRequest{}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredRegion] = req.GetCredential().GetRegion()
	params[api.OptCredEndpoint] = req.GetCredential().GetEndpoint()
	params[api.OptCredAccessKey] = req.GetCredential().GetAccessKey()
	params[api.OptCredSecretKey] = req.GetCredential().GetSecretKey()

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.CreateForAWS(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Access Key")
}

func TestSdkAzureCredentialCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateAzureRequest{
		Credential: &api.AzureCredential{
			AccountKey:  "dummy-account-key",
			AccountName: "dummy-account-name",
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredAzureAccountKey] = req.GetCredential().GetAccountKey()
	params[api.OptCredAzureAccountName] = req.GetCredential().GetAccountName()

	uuid := "good-uuid"
	s.MockDriver().
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	s.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Azure Creds
	_, err := c.CreateForAzure(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkAzureCredentialCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateAzureRequest{
		Credential: &api.AzureCredential{
			AccountKey:  "dummy-account-key",
			AccountName: "dummy-account-name",
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredAzureAccountKey] = req.GetCredential().GetAccountKey()
	params[api.OptCredAzureAccountName] = req.GetCredential().GetAccountName()

	uuid := "bad-uuid"
	s.MockDriver().
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	s.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(fmt.Errorf("Invalid credentials"))

	s.MockDriver().
		EXPECT().
		CredsDelete(uuid).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.CreateForAzure(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.PermissionDenied)
	assert.Contains(t, serverError.Message(), "Invalid credentials")
}

func TestSdkAzureCredentialCreateBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateAzureRequest{}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredAzureAccountKey] = req.GetCredential().GetAccountKey()
	params[api.OptCredAzureAccountName] = req.GetCredential().GetAccountName()

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.CreateForAzure(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Account Key")
}
func TestSdkGoogleCredentialCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateGoogleRequest{
		Credential: &api.GoogleCredential{
			ProjectId: "dummy-project-id",
			JsonKey:   "dummy-json-key",
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredGoogleJsonKey] = req.GetCredential().GetJsonKey()
	params[api.OptCredGoogleProjectID] = req.GetCredential().GetProjectId()

	uuid := "good-uuid"
	s.MockDriver().
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	s.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Google Credentials
	_, err := c.CreateForGoogle(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkGoogleCredentialCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateGoogleRequest{
		Credential: &api.GoogleCredential{
			ProjectId: "dummy-project-id",
			JsonKey:   "dummy-json-key",
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredGoogleJsonKey] = req.GetCredential().GetJsonKey()
	params[api.OptCredGoogleProjectID] = req.GetCredential().GetProjectId()

	uuid := "bad-uuid"
	s.MockDriver().
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	s.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(fmt.Errorf("Invalid credentials"))

	s.MockDriver().
		EXPECT().
		CredsDelete(uuid).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.CreateForGoogle(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.PermissionDenied)
	assert.Contains(t, serverError.Message(), "Invalid credentials")
}

func TestSdkGoogleCredentialCreateBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateGoogleRequest{}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredGoogleJsonKey] = req.GetCredential().GetJsonKey()
	params[api.OptCredGoogleProjectID] = req.GetCredential().GetProjectId()

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.CreateForGoogle(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply JSON Key")
}

func TestSdkCredentialValidateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "good-uuid"

	req := &api.SdkCredentialValidateRequest{CredentialId: uuid}

	s.MockDriver().
		EXPECT().
		CredsValidate(req.GetCredentialId()).
		Return(nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Validate Created Credentials
	_, err := c.CredentialValidate(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCredentialValidateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "bad-uuid"

	req := &api.SdkCredentialValidateRequest{CredentialId: uuid}

	s.MockDriver().
		EXPECT().
		CredsValidate(req.GetCredentialId()).
		Return(fmt.Errorf("Failed to Validate Credentials"))

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Validate Created Credentials
	_, err := c.CredentialValidate(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.Internal)
	assert.Contains(t, serverError.Message(), "Failed to Validate Credentials")
}

func TestSdkCredentialValidateBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := ""

	req := &api.SdkCredentialValidateRequest{CredentialId: uuid}

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Validate Created Credentials
	_, err := c.CredentialValidate(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must provide credentials uuid")

}

func TestSdkCredentialEnumerateAWSSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateAWSRequest{CredentialId: "test"}

	enumS3test1 := map[string]interface{}{
		api.OptCredType:      "s3",
		api.OptCredAccessKey: "test-access",
		api.OptCredSecretKey: "test-secret",
		api.OptCredEndpoint:  "test-endpoint",
		api.OptCredRegion:    "test-region",
	}

	enumS3test2 := map[string]interface{}{
		api.OptCredType:      "s3",
		api.OptCredAccessKey: "test-access1",
		api.OptCredSecretKey: "test-secret1",
		api.OptCredEndpoint:  "test-endpoint1",
		api.OptCredRegion:    "test-region1",
	}

	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		"test-cred-uuid1": enumS3test1,
		"test-cred-uuid3": enumAzure,
		"test-cred-uui2":  enumS3test2,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Enumerate AWS credentials
	resp, err := c.EnumerateForAWS(context.Background(), req)
	assert.NoError(t, err)

	assert.Len(t, resp.GetCredential(), 2)
}

func TestSdkCredentialEnumerateAWSWithMultipleCredResponseSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateAWSRequest{CredentialId: "test"}

	enumS3 := map[string]interface{}{
		api.OptCredType:      "s3",
		api.OptCredAccessKey: "test-access",
		api.OptCredSecretKey: "test-secret",
		api.OptCredEndpoint:  "test-endpoint",
		api.OptCredRegion:    "test-region",
	}

	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}

	enumGoogle := map[string]interface{}{
		api.OptCredType:            "google",
		api.OptCredGoogleProjectID: "test-google-project-id",
	}

	enumerateData := map[string]interface{}{
		"test-s3-uuid1":     enumS3,
		"test-azure-uuid1":  enumAzure,
		"test-Google-uuid1": enumGoogle,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Enumerate AWS credentials
	resp, err := c.EnumerateForAWS(context.Background(), req)
	assert.NoError(t, err)
	assert.Len(t, resp.GetCredential(), 1)

	// Should Return only AWS creds
	assert.Equal(t, resp.GetCredential()[0].AccessKey, enumS3[api.OptCredAccessKey])
	assert.Equal(t, resp.GetCredential()[0].Endpoint, enumS3[api.OptCredEndpoint])
}

func TestSdkCredentialEnumerateAWSFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateAWSRequest{CredentialId: "test"}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(nil, fmt.Errorf("Failed to get credenntials details"))

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// EnumerateCredentials for AWS
	resp, err := c.EnumerateForAWS(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)

}

func TestSdkCredentialEnumerateAzureSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateAzureRequest{CredentialId: "test"}

	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		api.OptCredUUID: enumAzure,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Enumerate Azure Credentials
	resp, err := c.EnumerateForAzure(context.Background(), req)
	assert.NoError(t, err)

	assert.Equal(t, resp.GetCredential()[0].AccountName, enumAzure[api.OptCredAzureAccountName])
	assert.Equal(t, resp.GetCredential()[0].AccountKey, enumAzure[api.OptCredAzureAccountKey])

}

func TestSdkCredentialEnumerateAzureFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateAzureRequest{CredentialId: "test"}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(nil, fmt.Errorf("Failed to get credenntials details"))

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// EnumerateCredentials for AWS
	resp, err := c.EnumerateForAzure(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestSdkCredentialEnumerateGoogleSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateGoogleRequest{CredentialId: "test"}

	enumGoogle := map[string]interface{}{
		api.OptCredType:            "google",
		api.OptCredGoogleProjectID: "test-google-project-id",
	}
	enumerateData := map[string]interface{}{
		api.OptCredUUID: enumGoogle,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

		// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Enumerate Google credentials
	resp, err := c.EnumerateForGoogle(context.Background(), req)
	assert.NoError(t, err)

	assert.Equal(t, resp.GetCredential()[0].ProjectId, enumGoogle[api.OptCredGoogleProjectID])
}

func TestSdkCredentialEnumerateGoogleFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateGoogleRequest{CredentialId: "test"}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(nil, fmt.Errorf("Failed to get credenntials details"))

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// EnumerateCredentials for AWS
	resp, err := c.EnumerateForGoogle(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestSdkCredentialDeleteSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	cred_id := "myid"
	req := &api.SdkCredentialDeleteRequest{
		CredentialId: cred_id,
	}
	s.MockDriver().
		EXPECT().
		CredsDelete(cred_id).
		Return(nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Delete Credentials
	_, err := c.CredentialDelete(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCredentialDeleteBadArgument(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	cred_id := ""
	req := &api.SdkCredentialDeleteRequest{
		CredentialId: cred_id,
	}

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Delete Credentials
	_, err := c.CredentialDelete(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must provide credentials uuid")
}

func TestSdkCredentialDeleteFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	cred_id := "myid"
	req := &api.SdkCredentialDeleteRequest{
		CredentialId: cred_id,
	}
	s.MockDriver().
		EXPECT().
		CredsDelete(cred_id).
		Return(fmt.Errorf("Error deleting credentials"))

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Delete Credentials
	_, err := c.CredentialDelete(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error deleting credentials")
}
