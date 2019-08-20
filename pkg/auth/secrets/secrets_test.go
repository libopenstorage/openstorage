package secrets

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAuth(t *testing.T) {
	a, err := NewAuth(TypeK8s, nil)
	require.EqualError(t, ErrSecretsNotInitialized, err.Error(), "Expected an error on NewAuth")
	require.Nil(t, a, "Expected auth object to be nil")

	secretsInstance, _ := getSecretsMock(t)
	a, err = NewAuth(111, secretsInstance)
	require.Error(t, err, "Expected an error on NewAuth")
	require.Nil(t, a, "Expected auth object to be nil")
}

func TestGetToken(t *testing.T) {
	tt := []struct {
		testName string

		secretType      AuthTokenProviders
		secretName      string
		secretNamespace string
		token           string
		publicData      bool
		customData      bool

		expectError   bool
		expectedError string
		expectedToken string
	}{
		{
			testName: "k8s get token",

			secretType:      TypeK8s,
			secretName:      "secret-name-k8s",
			secretNamespace: "ns",
			token:           "auth-token",

			expectError: false,
		},
		{
			testName: "k8s empty token",

			secretType:      TypeK8s,
			secretName:      "secret-name-k8s-empty",
			secretNamespace: "ns",
			token:           "",

			expectError:   true,
			expectedError: ErrAuthTokenNotFound.Error(),
		},
		{
			testName: "dcos get token",

			secretType:      TypeDCOS,
			secretName:      "secret-name-dcos",
			secretNamespace: "ns",
			token:           "auth-token",

			expectError:   false,
			expectedToken: "ns/secret-name-dcos",
		},
		{
			testName: "dcos empty token",

			secretType:      TypeDCOS,
			secretName:      "secret-name-dcos-empty",
			secretNamespace: "ns",
			token:           "",

			expectError:   true,
			expectedError: ErrAuthTokenNotFound.Error(),
		},
		{
			testName: "dcos empty namespace",

			secretType:      TypeDCOS,
			secretName:      "secret-name-dcos-no-ns",
			secretNamespace: "",
			token:           "abcd",

			expectError: false,
		},
		{
			testName: "vault get token",

			secretType: TypeVault,
			secretName: "secret-name-vault",
			token:      "auth-token",

			expectError: false,
		},
		{
			testName: "docker get token",

			secretType: TypeDocker,
			secretName: "secret-name-docker",
			token:      "auth-token",

			expectError: false,
		},
		{
			testName: "kvdb get token",

			secretType: TypeKVDB,
			secretName: "secret-name-kvdb",
			token:      "auth-token",

			expectError: false,
		},
	}

	for _, tc := range tt {
		secretContext := make(map[string]string)
		if tc.secretNamespace != "" {
			secretContext[SecretNamespaceKey] = tc.secretNamespace
		}
		if tc.publicData {
			secretContext[secrets.PublicSecretData] = "true"
		}
		if tc.customData {
			secretContext[secrets.CustomSecretData] = "true"
		}
		s, mockSecret := getSecretsMock(t)
		a, err := NewAuth(tc.secretType, s)
		require.NoError(t, err, "Expected no error on auth")
		expectedTokenResponse := make(map[string]interface{})
		if !tc.expectError {
			switch tc.secretType {
			case TypeDCOS:
				key := tc.secretName
				if tc.secretNamespace != "" {
					key = tc.secretNamespace + "/" + key
				}
				expectedTokenResponse[key] = tc.expectedToken
			case TypeK8s:
				expectedTokenResponse[SecretTokenKey] = tc.expectedToken
			default:
				expectedTokenResponse[SecretNameKey] = tc.expectedToken
			}
		}
		mockSecret.EXPECT().
			GetSecret(
				gomock.Any(),
				gomock.Any(),
			).
			Return(expectedTokenResponse, nil).
			Times(1)

		req := &api.TokenSecretContext{
			SecretName:      tc.secretName,
			SecretNamespace: tc.secretNamespace,
		}
		gotToken, err := a.GetToken(req)
		if tc.expectError {
			if err == nil {
				t.Errorf("[%s]: Expected error on GetToken, but got nil", tc.testName)
			}
			if err != nil && err.Error() != tc.expectedError {
				t.Errorf("[%s]: Expected error '%s' on GetToken, but got '%s'", tc.testName, err.Error(), tc.expectedError)
			}
		} else {
			if err != nil {
				t.Errorf("[%s]: Expected no error on GetToken, but got '%s'", tc.testName, err.Error())
			}
			if gotToken != tc.expectedToken {
				t.Errorf("[%s]: Expected token '%s' on GetToken, but got '%s'", tc.testName, tc.expectedToken, gotToken)
			}
		}
	}
}

func getSecretsMock(t *testing.T) (secrets.Secrets, *mock.MockSecrets) {
	mockCtrl := gomock.NewController(t)
	mockSecret := mock.NewMockSecrets(mockCtrl)
	return mockSecret, mockSecret
}
