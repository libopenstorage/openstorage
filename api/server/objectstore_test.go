package server

import (
	"fmt"
	"testing"

	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/objectstore"
	"github.com/stretchr/testify/assert"
)

func TestObjectStoreInspectSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objInfo := &objectstore.ObjectstoreInfo{
		UUID:     "bbf89474-053b-45c1-b24f-d1dbac52638c",
		VolumeID: "328808731955060606",
		Enabled:  false,
	}
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreInspect().
		Return(objInfo, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.ObjectStoreInspect()

	assert.NoError(t, err)
	assert.Equal(t, resp.UUID, objInfo.UUID)
	assert.Equal(t, resp.VolumeID, objInfo.VolumeID)
}

func TestObjectStoreInspectFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreInspect().
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	_, err = restClient.ObjectStoreInspect()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}
func TestObjectStoreCreateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testvol1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreCreate(name).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreCreate(name)

	assert.NoError(t, err)
}

func TestObjectStoreCreateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testvol1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreCreate(name).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreCreate(name)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestObjectStoreUpdateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := true
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreUpdate(name).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreUpdate(name)

	assert.NoError(t, err)
}

func TestObjectStoreUpdateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := false
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreUpdate(name).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreUpdate(name)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestObjectStoreDeleteSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreDelete().
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreDelete()

	assert.NoError(t, err)
}

func TestObjectStoreDeleteFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreDelete().
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreDelete()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}
