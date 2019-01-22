package secrets

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/k8s"
	"github.com/libopenstorage/secrets/mock"
	"github.com/stretchr/testify/require"
)

func TestNewAuthOnError(t *testing.T) {
	a, err := NewAuth(TypeK8s, nil)
	require.EqualError(t, ErrSecretsNotInitialized, err.Error(), "Expected an error on NewAuth")
	require.Nil(t, a, "Expected auth object to be nil")

	secretsInstance, _ := getSecretsMock(t)
	a, err = NewAuth(5, secretsInstance)
	require.Error(t, err, "Expected an error on NewAuth")
	require.Nil(t, a, "Expected auth object to be nil")
}

func TestK8sGetToken(t *testing.T) {
	s, mockSecret := getSecretsMock(t)
	a, err := NewAuth(TypeK8s, s)
	require.NoError(t, err, "Expected no error on auth")
	name := "secret-name"
	namespace := "ns"
	token := "auth-token"
	mockSecret.EXPECT().
		GetSecret(
			name,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(map[string]interface{}{SecretTokenKey: token}, nil).
		Times(1)
	gotToken, err := a.GetToken(name, namespace)
	require.NoError(t, err, "Unexpected error on GetToken")
	require.Equal(t, gotToken, token, "Unexpected token returned")
}

func TestK8sGetTokenWithEmptyTokent(t *testing.T) {
	s, mockSecret := getSecretsMock(t)
	a, err := NewAuth(TypeK8s, s)
	require.NoError(t, err, "Expected no error on auth")
	name := "secret-name"
	namespace := "ns"
	mockSecret.EXPECT().
		GetSecret(
			name,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(map[string]interface{}{}, nil).
		Times(1)
	_, err = a.GetToken(name, namespace)
	require.EqualError(t, ErrAuthTokenNotFound, err.Error(), "Unexpected error on GetToken")
}

func TestDCOSGetToken(t *testing.T) {
	s, mockSecret := getSecretsMock(t)
	a, err := NewAuth(TypeDCOS, s)
	require.NoError(t, err, "Expected no error on auth")
	name := "secret-name"
	namespace := "ns"
	token := "auth-token"
	key := namespace + "/" + name
	mockSecret.EXPECT().
		GetSecret(
			key,
			nil,
		).
		Return(map[string]interface{}{key: token}, nil).
		Times(1)
	gotToken, err := a.GetToken(name, namespace)
	require.NoError(t, err, "Unexpected error on GetToken")
	require.Equal(t, gotToken, token, "Unexpected token returned")
}

func TestDCOSGetTokenWithNoNamespace(t *testing.T) {
	s, mockSecret := getSecretsMock(t)
	a, err := NewAuth(TypeDCOS, s)
	require.NoError(t, err, "Expected no error on auth")
	name := "secret-name"
	token := "auth-token"
	key := name
	mockSecret.EXPECT().
		GetSecret(
			key,
			nil,
		).
		Return(map[string]interface{}{key: token}, nil).
		Times(1)
	gotToken, err := a.GetToken(name, "")
	require.NoError(t, err, "Unexpected error on GetToken")
	require.Equal(t, gotToken, token, "Unexpected token returned")
}

func TestDCOSGetTokenWithEmptyToken(t *testing.T) {
	s, mockSecret := getSecretsMock(t)
	a, err := NewAuth(TypeDCOS, s)
	require.NoError(t, err, "Expected no error on auth")
	name := "secret-name"
	namespace := "ns"
	key := namespace + "/" + name
	mockSecret.EXPECT().
		GetSecret(
			key,
			nil,
		).
		Return(map[string]interface{}{}, nil).
		Times(1)
	_, err = a.GetToken(name, namespace)
	require.EqualError(t, ErrAuthTokenNotFound, err.Error(), "Unexpected error on GetToken")
}

func getSecretsMock(t *testing.T) (secrets.Secrets, *mock.MockSecrets) {
	mockCtrl := gomock.NewController(t)
	mockSecret := mock.NewMockSecrets(mockCtrl)
	return mockSecret, mockSecret
}
