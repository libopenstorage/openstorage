package server

import (
	"fmt"
	"testing"
	"time"

	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/osdconfig"
	"github.com/stretchr/testify/assert"
)

func TestGetClusterConfSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretsConfig := &osdconfig.SecretsConfig{
		ClusterSecretKey: "cluster-secret-key",
		SecretType:       "vault",
		Vault: &osdconfig.VaultConfig{
			Address:    "/vault/addr",
			BasePath:   "1.1.1.1",
			CACert:     "vault-ca-cert",
			ClientCert: "vault-client-cert",
			Token:      "vault--dummy-token",
		},
	}

	clusterConfig := &osdconfig.ClusterConfig{
		ClusterId: "dummy-cluster-id",
		Created:   time.Now(),
		Secrets:   secretsConfig,
		Version:   "x.y.z",
		Kvdb: &osdconfig.KvdbConfig{
			Discovery: []string{"2.2.2.2"},
			Password:  "kvdb-pass",
			Username:  "kvdb",
		},
	}
	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		GetClusterConf().
		Return(clusterConfig, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.GetClusterConf()
	assert.NoError(t, err)

	assert.Equal(t, resp.ClusterId, clusterConfig.ClusterId)
	assert.Equal(t, resp.Version, clusterConfig.Version)
	assert.Equal(t, resp.Kvdb.Password, clusterConfig.Kvdb.Password)
}

func TestSetClusterConfSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	secretsConfig := &osdconfig.SecretsConfig{
		ClusterSecretKey: "cluster-secret-key",
		SecretType:       "vault",
		Vault: &osdconfig.VaultConfig{
			Address:    "/vault/addr",
			BasePath:   "1.1.1.1",
			CACert:     "vault-ca-cert",
			ClientCert: "vault-client-cert",
			Token:      "vault--dummy-token",
		},
	}

	clusterConfig := &osdconfig.ClusterConfig{
		ClusterId: "dummy-cluster-id",
		Secrets:   secretsConfig,
		Version:   "x.y.z",
		Kvdb: &osdconfig.KvdbConfig{
			Discovery: []string{"2.2.2.2"},
			Password:  "kvdb-pass",
			Username:  "kvdb",
		},
	}

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		SetClusterConf(clusterConfig).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.SetClusterConf(clusterConfig)
	assert.NoError(t, resp)
}

func TestGetNodeConfSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	nodeID := "dummy-node-id"
	nodeConfig := &osdconfig.NodeConfig{
		NodeId: nodeID,
		Storage: &osdconfig.StorageConfig{
			Devices:          []string{"/dev/sdb", "/dev/sdc"},
			MaxDriveSetCount: 5,
			MaxCount:         5,
		},
		Network: &osdconfig.NetworkConfig{
			DataIface: "eth0",
			MgtIface:  "dummy",
		},
	}

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		GetNodeConf(nodeID).
		Return(nodeConfig, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.GetNodeConf(nodeID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, nodeID, resp.NodeId)
	assert.Equal(t, nodeConfig.Storage.Devices[0], "/dev/sdb")
	assert.Equal(t, nodeConfig.Storage.Devices[1], "/dev/sdc")
}

func TestGetNodeConfFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	nodeID := "dummy-node-id"
	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		GetNodeConf(nodeID).
		Return(nil, fmt.Errorf("error in getting node config"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.GetNodeConf(nodeID)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestSetNodeConfSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	nodeID := "dummy-node-id"
	nodeConfig := &osdconfig.NodeConfig{
		NodeId: nodeID,
		Storage: &osdconfig.StorageConfig{
			Devices:          []string{"/dev/sdb", "/dev/sdc"},
			MaxDriveSetCount: 5,
			MaxCount:         5,
		},
		Network: &osdconfig.NetworkConfig{
			DataIface: "eth0",
			MgtIface:  "dummy",
		},
	}

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		SetNodeConf(nodeConfig).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.SetNodeConf(nodeConfig)
	assert.NoError(t, resp)
}
