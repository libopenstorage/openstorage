package api

import (
	"strings"
	"testing"
)

func TestGetSdkCredsAws(t *testing.T) {
	testCases := []struct {
		description    string
		params         map[string]string
		expectedErrMsg string
	}{
		{
			description: "Test AWS Success",
			params: map[string]string{
				OptCredAccessKey:        "access_key",
				OptCredSecretKey:        "secret_key",
				OptCredEndpoint:         "endpoint",
				OptCredDisablePathStyle: "true",
				OptCredDisableSSL:       "true",
				OptCredRegion:           "us-west1",
				OptCredType:             credAws,
			},
		},
		{
			description: "Test AWS Fail",
			params: map[string]string{
				OptCredAccessKey:        "access_key",
				OptCredSecretKey:        "secret_key",
				OptCredEndpoint:         "endpoint",
				OptCredDisablePathStyle: "malformed bool",
				OptCredDisableSSL:       "true",
				OptCredRegion:           "us-west1",
				OptCredType:             credAws,
			},
			expectedErrMsg: "parsing",
		},
		{
			description: "Unknown provider",
			params: map[string]string{
				OptCredType: "unknown",
			},
			expectedErrMsg: "unsupported cred type",
		},
	}

	for _, testCase := range testCases {
		obj, err := GetSdkCreds(testCase.params)

		if len(testCase.expectedErrMsg) != 0 && err == nil {
			t.Errorf("Error must not be nil")
		}

		if len(testCase.expectedErrMsg) == 0 && err != nil {
			t.Errorf("Unexpected error %v", err)
		}

		if err != nil && !strings.Contains(err.Error(), testCase.expectedErrMsg) {
			t.Errorf("Error message %s does  not contain expected message %s",
				err.Error(), testCase.expectedErrMsg)
		}

		if obj != nil {
			_, ok := obj.(*SdkCredentialCreateRequest_AwsCredential)

			if !ok {
				t.Errorf("Cannot convert %v to SdkCredentialCreateRequest_AwsCredential", obj)
				continue
			}
		}
	}
}

func TestGetSdkCredsAzure(t *testing.T) {
	params := map[string]string{
		OptCredAzureAccountKey:  "account_key",
		OptCredAzureAccountName: "account_name",
		OptCredType:             credAzure,
	}

	obj, err := GetSdkCreds(params)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if obj != nil {
		creds, ok := obj.(*SdkCredentialCreateRequest_AzureCredential)

		if !ok {
			t.Errorf("Cannot convert %v to SdkCredentialCreateRequest_AzureCredential", obj)
		}

		if creds.AzureCredential.AccountKey != params[OptCredAzureAccountKey] {
			t.Errorf("wront account key value exppected %s actual %s",
				params[OptCredAzureAccountKey], creds.AzureCredential.AccountKey)
		}

		if creds.AzureCredential.AccountName != params[OptCredAzureAccountName] {
			t.Errorf("wront account name value exppected %s actual %s",
				params[OptCredAzureAccountName], creds.AzureCredential.AccountName)
		}
	}
}

func TestGetSdkCredsGoogle(t *testing.T) {
	params := map[string]string{
		OptCredGoogleJsonKey:   "{}",
		OptCredGoogleProjectID: "project_id",
		OptCredType:            credGoogle,
	}

	obj, err := GetSdkCreds(params)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
		return
	}

	if obj != nil {
		creds, ok := obj.(*SdkCredentialCreateRequest_GoogleCredential)

		if !ok {
			t.Errorf("Cannot convert %v to SdkCredentialCreateRequest_AzureCredential", obj)
		}

		if creds.GoogleCredential.ProjectId != params[OptCredGoogleProjectID] {
			t.Errorf("wront project_id value exppected %s actual %s",
				params[OptCredGoogleProjectID], creds.GoogleCredential.ProjectId)
		}

		if creds.GoogleCredential.JsonKey != params[OptCredGoogleJsonKey] {
			t.Errorf("wront json key value exppected %s actual %s",
				params[OptCredGoogleJsonKey], creds.GoogleCredential.JsonKey)
		}
	}
}
