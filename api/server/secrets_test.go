package server

import (
	"fmt"
	"testing"

	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/stretchr/testify/assert"
)

func TestSecretSetDefaultSecretKeySuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretKey := "testkey"
	overrideFlag := true
	// mock the cluster secret response
	tc.MockClusterSecrets().
		EXPECT().
		SecretSetDefaultSecretKey(secretKey, overrideFlag).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretSetDefaultSecretKey(secretKey, overrideFlag)

	assert.NoError(t, err)
}

func TestSecretSetDefaultSecretKeyFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretKey := "testClusterKey"
	overrideFlag := false
	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretSetDefaultSecretKey(secretKey, overrideFlag).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretSetDefaultSecretKey(secretKey, overrideFlag)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSecretGetDefaultSecretKeySuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	defaultSecretTest := "testclusterkeyval"
	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretGetDefaultSecretKey().
		Return(defaultSecretTest, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.SecretGetDefaultSecretKey()

	assert.NoError(t, err)
	assert.Equal(t, resp.(map[string]interface{})["SecretValue"], defaultSecretTest)
}

func TestSecretGetDefaultSecretKeyFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretGetDefaultSecretKey().
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.SecretGetDefaultSecretKey()

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestGetSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretKey := "testkey"
	secretValue := "testvalue"
	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretGet(secretKey).
		Return(secretValue, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	getResp, err := restClient.SecretGet(secretKey)

	assert.NoError(t, err)
	assert.EqualValues(t, getResp.(map[string]interface{})["SecretValue"], secretValue)
}

func TestSecretGetFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	testKey := "testkey"
	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretGet(testKey).
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	getResp, err := restClient.SecretGet(testKey)

	assert.Error(t, err)
	assert.Nil(t, getResp)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSetSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretKey := "testsecretkey"
	secretValue := "testsecretvalue"
	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretSet(secretKey, secretValue).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretSet(secretKey, secretValue)
	assert.NoError(t, err)
}

func TestSetFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretKey := "testsecretkey"
	secretValue := "testsecretvalue"
	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretSet(secretKey, secretValue).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretSet(secretKey, secretValue)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")

}

func TestVerifySuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretCheckLogin().
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretCheckLogin()

	assert.NoError(t, err)
}

func TestVerifyFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster secrets response
	tc.MockClusterSecrets().
		EXPECT().
		SecretCheckLogin().
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretCheckLogin()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSecretLoginSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster secrets response
	secretType := "teststore1"
	secretConfig := make(map[string]string)
	secretConfig["testconfig1"] = "testconfigdata1"
	secretConfig["testconfig2"] = "testconfigdata2"
	tc.MockClusterSecrets().
		EXPECT().
		SecretLogin(secretType, secretConfig).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretLogin(secretType, secretConfig)

	assert.NoError(t, err)
}

func TestSecretLoginFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster secrets response
	secretType := "teststore2"
	tc.MockClusterSecrets().
		EXPECT().
		SecretLogin(secretType, nil).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretLogin(secretType, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSecretLoginAccessDenied(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster secrets response
	secretType := "teststore2"
	tc.MockClusterSecrets().
		EXPECT().
		SecretLogin(secretType, nil).
		Return(fmt.Errorf("Not authenticated with the secrets endpoint"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SecretLogin(secretType, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not authenticated")
}
