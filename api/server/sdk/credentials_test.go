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

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kubernetes-csi/csi-test/utils"

	"github.com/libopenstorage/openstorage/api"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/volume"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSdkAWSCredentialCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateRequest{
		Name:          "test",
		Bucket:        "mybucket",
		EncryptionKey: "key",
		CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialRequest{
				AccessKey:  "dummy-access",
				SecretKey:  "dummy-secret",
				Endpoint:   "dummy-endpoint",
				Region:     "dummy-region",
				DisableSsl: true,
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = req.GetEncryptionKey()
	params[api.OptCredBucket] = req.GetBucket()
	params[api.OptCredRegion] = req.GetAwsCredential().GetRegion()
	params[api.OptCredEndpoint] = req.GetAwsCredential().GetEndpoint()
	params[api.OptCredAccessKey] = req.GetAwsCredential().GetAccessKey()
	params[api.OptCredSecretKey] = req.GetAwsCredential().GetSecretKey()
	params[api.OptCredDisableSSL] = "true"

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
	_, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkAWSCredentialCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialRequest{
				AccessKey: "dummy-access",
				SecretKey: "dummy-secret",
				Endpoint:  "dummy-endpoint",
				Region:    "dummy-region",
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = ""
	params[api.OptCredBucket] = ""
	params[api.OptCredRegion] = req.GetAwsCredential().GetRegion()
	params[api.OptCredEndpoint] = req.GetAwsCredential().GetEndpoint()
	params[api.OptCredAccessKey] = req.GetAwsCredential().GetAccessKey()
	params[api.OptCredSecretKey] = req.GetAwsCredential().GetSecretKey()
	params[api.OptCredDisableSSL] = "false"

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
	_, err := c.Create(context.Background(), req)
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

	req := &api.SdkCredentialCreateRequest{}

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.Create(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "name")

	// Set AWS missing key
	req = &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialRequest{
				Endpoint: "dummy-endpoint",
				Region:   "dummy-region",
			},
		},
	}

	// Create Credentials
	_, err = c.Create(context.Background(), req)
	assert.Error(t, err)

	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Access Key")
}

func TestSdkAzureCredentialCreateSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_AzureCredential{
			AzureCredential: &api.SdkAzureCredentialRequest{
				AccountKey:  "dummy-account-key",
				AccountName: "dummy-account-name",
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = ""
	params[api.OptCredBucket] = ""
	params[api.OptCredAzureAccountKey] = req.GetAzureCredential().GetAccountKey()
	params[api.OptCredAzureAccountName] = req.GetAzureCredential().GetAccountName()

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
	_, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkAzureCredentialCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_AzureCredential{
			AzureCredential: &api.SdkAzureCredentialRequest{
				AccountKey:  "dummy-account-key",
				AccountName: "dummy-account-name",
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "azure"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = ""
	params[api.OptCredBucket] = ""
	params[api.OptCredAzureAccountKey] = req.GetAzureCredential().GetAccountKey()
	params[api.OptCredAzureAccountName] = req.GetAzureCredential().GetAccountName()

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
	_, err := c.Create(context.Background(), req)
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

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_AzureCredential{
			AzureCredential: &api.SdkAzureCredentialRequest{
				AccountName: "dummy-account-name",
			},
		},
	}

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.Create(context.Background(), req)
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

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_GoogleCredential{
			GoogleCredential: &api.SdkGoogleCredentialRequest{
				ProjectId: "dummy-project-id",
				JsonKey:   "dummy-json-key",
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = ""
	params[api.OptCredBucket] = ""
	params[api.OptCredGoogleJsonKey] = req.GetGoogleCredential().GetJsonKey()
	params[api.OptCredGoogleProjectID] = req.GetGoogleCredential().GetProjectId()

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
	_, err := c.Create(context.Background(), req)
	assert.NoError(t, err)
}
func TestSdkGoogleCredentialCreateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_GoogleCredential{
			GoogleCredential: &api.SdkGoogleCredentialRequest{
				ProjectId: "dummy-project-id",
				JsonKey:   "dummy-json-key",
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = ""
	params[api.OptCredBucket] = ""
	params[api.OptCredGoogleJsonKey] = req.GetGoogleCredential().GetJsonKey()
	params[api.OptCredGoogleProjectID] = req.GetGoogleCredential().GetProjectId()

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
	_, err := c.Create(context.Background(), req)
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

	req := &api.SdkCredentialCreateRequest{
		Name: "test",
		CredentialType: &api.SdkCredentialCreateRequest_GoogleCredential{
			GoogleCredential: &api.SdkGoogleCredentialRequest{
				ProjectId: "dummy-project-id",
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "google"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredGoogleJsonKey] = req.GetGoogleCredential().GetJsonKey()
	params[api.OptCredGoogleProjectID] = req.GetGoogleCredential().GetProjectId()

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Create Credentials
	_, err := c.Create(context.Background(), req)
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

	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		uuid: enumAzure,
	}
	s.MockDriver().
		EXPECT().
		CredsValidate(req.GetCredentialId()).
		Return(nil)
	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Validate Created Credentials
	_, err := c.Validate(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkCredentialValidateFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "bad-uuid"

	req := &api.SdkCredentialValidateRequest{CredentialId: uuid}
	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		uuid: enumAzure,
	}

	s.MockDriver().
		EXPECT().
		CredsValidate(req.GetCredentialId()).
		Return(fmt.Errorf("Failed to Validate Credentials"))
	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Validate Created Credentials
	_, err := c.Validate(context.Background(), req)
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
	_, err := c.Validate(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must provide credentials uuid")

}

func TestSdkCredentialEnumerate(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialEnumerateRequest{}

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
		"test-cred-uuid2": enumS3test2,
		"test-cred-uuid3": enumAzure,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Enumerate AWS credentials
	resp, err := c.Enumerate(context.Background(), req)
	assert.NoError(t, err)

	assert.Len(t, resp.GetCredentialIds(), 3)
	assert.Contains(t, resp.GetCredentialIds(), "test-cred-uuid1")
	assert.Contains(t, resp.GetCredentialIds(), "test-cred-uuid2")
	assert.Contains(t, resp.GetCredentialIds(), "test-cred-uuid3")
}

func TestSdkCredentialInspectIdNotFound(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialInspectRequest{
		CredentialId: "test",
	}

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

	// Inspect
	_, err := c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
}

func TestSdkCredentialInspectFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	req := &api.SdkCredentialInspectRequest{}

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Inpect
	_, err := c.Inspect(context.Background(), req)
	assert.Error(t, err)

	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "credential id")
}

func TestSdkAWSInspect(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "test"
	req := &api.SdkCredentialInspectRequest{
		CredentialId: uuid,
	}

	enumAws := map[string]interface{}{
		api.OptCredType:       "s3",
		api.OptCredName:       "test",
		api.OptCredBucket:     "mybucket",
		api.OptCredEncrKey:    "key",
		api.OptCredRegion:     "test-azure-account",
		api.OptCredEndpoint:   "test-azure-account",
		api.OptCredAccessKey:  "access",
		api.OptCredSecretKey:  "secret",
		api.OptCredDisableSSL: "false",
	}
	enumerateData := map[string]interface{}{
		uuid: enumAws,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Inspect
	resp, err := c.Inspect(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.GetAwsCredential())
	assert.Nil(t, resp.GetOwnership())
	assert.Equal(t, enumAws[api.OptCredName], resp.GetName())
	assert.Equal(t, enumAws[api.OptCredBucket], resp.GetBucket())
	assert.Equal(t, enumAws[api.OptCredRegion], resp.GetAwsCredential().GetRegion())
	assert.Equal(t, enumAws[api.OptCredEndpoint], resp.GetAwsCredential().GetEndpoint())
	assert.Equal(t, enumAws[api.OptCredAccessKey], resp.GetAwsCredential().GetAccessKey())
	assert.Equal(t, enumAws[api.OptCredDisableSSL] == "true", resp.GetAwsCredential().GetDisableSsl())
}

func TestSdkCredentialAzureInspect(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	uuid := "test"
	req := &api.SdkCredentialInspectRequest{
		CredentialId: uuid,
	}

	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredName:             "test",
		api.OptCredBucket:           "mybucket",
		api.OptCredEncrKey:          "key",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		uuid: enumAzure,
	}

	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Inspect
	resp, err := c.Inspect(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.GetAzureCredential())
	assert.Equal(t, resp.GetAzureCredential().GetAccountName(), enumAzure[api.OptCredAzureAccountName])
	assert.Equal(t, enumAzure[api.OptCredName], resp.GetName())
	assert.Equal(t, enumAzure[api.OptCredBucket], resp.GetBucket())
}

func TestSdkCredentialDeleteSuccess(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	cred_id := "myid"
	req := &api.SdkCredentialDeleteRequest{
		CredentialId: cred_id,
	}
	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		cred_id: enumAzure,
	}
	s.MockDriver().
		EXPECT().
		CredsDelete(cred_id).
		Return(nil)
	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Delete Credentials
	_, err := c.Delete(context.Background(), req)
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
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Must provide credentials uuid")
}

func TestSdkCredentialDeleteFailed(t *testing.T) {

	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	cred_id := "myid"
	enumAzure := map[string]interface{}{
		api.OptCredType:             "azure",
		api.OptCredAzureAccountName: "test-azure-account",
		api.OptCredAzureAccountKey:  "test-azure-account",
	}
	enumerateData := map[string]interface{}{
		cred_id: enumAzure,
	}

	req := &api.SdkCredentialDeleteRequest{
		CredentialId: cred_id,
	}
	s.MockDriver().
		EXPECT().
		CredsDelete(cred_id).
		Return(fmt.Errorf("Error deleting credentials"))
	s.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil)

	// Setup client
	c := api.NewOpenStorageCredentialsClient(s.Conn())

	// Delete Credentials
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error deleting credentials")
}

func TestSdkCredentialOwnership(t *testing.T) {
	// This test does not use the gRPC server
	mc := gomock.NewController(&utils.SafeGoroutineTester{})
	mv := mockdriver.NewMockVolumeDriver(mc)
	mcluster := mockcluster.NewMockCluster(mc)
	s := CredentialServer{
		server: &sdkGrpcServer{
			driverHandlers: map[string]volume.VolumeDriver{
				"mock":            mv,
				DefaultDriverName: mv,
			},
			clusterHandler: mcluster,
		},
	}

	// Create contexts
	user1 := "user1"
	ctxWithOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: user1,
	})

	req := &api.SdkCredentialCreateRequest{
		Name:          "test",
		Bucket:        "mybucket",
		EncryptionKey: "key",
		CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialRequest{
				AccessKey:  "dummy-access",
				SecretKey:  "dummy-secret",
				Endpoint:   "dummy-endpoint",
				Region:     "dummy-region",
				DisableSsl: true,
			},
		},
	}

	params := make(map[string]string)

	params[api.OptCredType] = "s3"
	params[api.OptCredName] = req.GetName()
	params[api.OptCredEncrKey] = req.GetEncryptionKey()
	params[api.OptCredBucket] = req.GetBucket()
	params[api.OptCredRegion] = req.GetAwsCredential().GetRegion()
	params[api.OptCredEndpoint] = req.GetAwsCredential().GetEndpoint()
	params[api.OptCredAccessKey] = req.GetAwsCredential().GetAccessKey()
	params[api.OptCredSecretKey] = req.GetAwsCredential().GetSecretKey()
	params[api.OptCredDisableSSL] = "true"

	// Create a marshalled ownership for the expect params
	ownership := &api.Ownership{
		Owner: user1,
		Acls: &api.Ownership_AccessControl{
			Collaborators: map[string]api.Ownership_AccessType{
				"collabread":  api.Ownership_Read,
				"collabadmin": api.Ownership_Admin,
			},
		},
	}
	m := jsonpb.Marshaler{OrigName: true}
	oStr, err := m.MarshalToString(ownership)
	params[api.OptCredOwnership] = oStr

	// add to request
	req.Ownership = ownership

	uuid := "good-uuid"
	mv.
		EXPECT().
		CredsCreate(params).
		Return(uuid, nil)

	mv.
		EXPECT().
		CredsValidate(uuid).
		Return(nil)

	_, err = s.Create(ctxWithOwner, req)
	assert.NoError(t, err)

	// Do an inspection with a different user
	user2 := "user2"
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: user2,
	})

	cred := make(map[string]interface{})
	for k, v := range params {
		cred[k] = v
	}
	creds := map[string]interface{}{
		uuid: cred,
	}

	mv.EXPECT().
		CredsEnumerate().
		Return(creds, nil)

	_, err = s.Inspect(ctxWithNotOwner, &api.SdkCredentialInspectRequest{
		CredentialId: uuid,
	})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.PermissionDenied, fmt.Sprintf("CODE: %v:%v", serverError.Code(), serverError.Message()))

	// Do an inspection with correct user
	mv.EXPECT().
		CredsEnumerate().
		Return(creds, nil)
	resp, err := s.Inspect(ctxWithOwner, &api.SdkCredentialInspectRequest{
		CredentialId: uuid,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.GetOwnership())
	assert.Equal(t, user1, resp.GetOwnership().GetOwner())

	// Delete the resource as a user
	ctxCollabRead := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "collabread",
	})
	ctxCollabAdmin := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "collabadmin",
	})

	// - test 1 cannot delete as a collaborator:READ
	mv.EXPECT().
		CredsEnumerate().
		Return(creds, nil)
	_, err = s.Delete(ctxCollabRead, &api.SdkCredentialDeleteRequest{
		CredentialId: uuid,
	})
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.PermissionDenied, fmt.Sprintf("CODE: %v:%v", serverError.Code(), serverError.Message()))

	// - test 1 cannot delete as a collaborator:READ
	mv.EXPECT().
		CredsEnumerate().
		Return(creds, nil)
	mv.EXPECT().CredsDelete(uuid).Return(nil)
	_, err = s.Delete(ctxCollabAdmin, &api.SdkCredentialDeleteRequest{
		CredentialId: uuid,
	})
	assert.NoError(t, err)
}

func TestSdkCredentialGetOwnershipFromCred(t *testing.T) {
	s := &CredentialServer{}

	o, err := s.getOwnershipFromCred(map[string]interface{}{
		api.OptCredOwnership: "",
	})
	assert.NoError(t, err)
	assert.Nil(t, o)

	o, err = s.getOwnershipFromCred(map[string]interface{}{
		"hello": "world",
	})
	assert.NoError(t, err)
	assert.Nil(t, o)

	ownership := &api.Ownership{
		Owner: "user1",
		Acls: &api.Ownership_AccessControl{
			Collaborators: map[string]api.Ownership_AccessType{
				"collabread":  api.Ownership_Read,
				"collabadmin": api.Ownership_Admin,
			},
		},
	}
	m := jsonpb.Marshaler{OrigName: true}
	oStr, err := m.MarshalToString(ownership)
	assert.NoError(t, err)
	o, err = s.getOwnershipFromCred(map[string]interface{}{
		api.OptCredOwnership: oStr,
	})
	assert.NoError(t, err)
	assert.NotNil(t, o)
}
