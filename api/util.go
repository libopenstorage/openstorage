package api

import (
	"strconv"

	"github.com/pkg/errors"
)

const (
	credAws    = "aws"
	credAzure  = "azure"
	credGoogle = "google"
)

var (
	ErrUnsupportedCredType = errors.New("unsupported cred type")
)

func GetSdkCreds(params map[string]string) (isSdkCredentialCreateRequest_CredentialType, error) {
	switch params[OptCredType] {
	case credAws:
		disableSsl, err := strconv.ParseBool(params[OptCredDisableSSL])

		if err != nil {
			return nil, err
		}

		disablePathStyle, err := strconv.ParseBool(params[OptCredDisablePathStyle])

		if err != nil {
			return nil, err
		}

		return &SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &SdkAwsCredentialRequest{
				AccessKey:        params[OptCredAccessKey],
				SecretKey:        params[OptCredSecretKey],
				Endpoint:         params[OptCredEndpoint],
				Region:           params[OptCredRegion],
				DisableSsl:       disableSsl,
				DisablePathStyle: disablePathStyle,
			},
		}, nil
	case credAzure:
		return &SdkCredentialCreateRequest_AzureCredential{
			AzureCredential: &SdkAzureCredentialRequest{
				AccountName: params[OptCredAzureAccountName],
				AccountKey:  params[OptCredAzureAccountKey],
			},
		}, nil
	case credGoogle:
		return &SdkCredentialCreateRequest_GoogleCredential{
			GoogleCredential: &SdkGoogleCredentialRequest{
				ProjectId: params[OptCredGoogleProjectID],
				JsonKey:   params[OptCredGoogleJsonKey],
			},
		}, nil
	}

	return nil, errors.Wrapf(ErrUnsupportedCredType, params[OptCredType])
}
