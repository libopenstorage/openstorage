package server

import (
	"fmt"
	"testing"

	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	//"github.com/libopenstorage/openstorage/services"
	"github.com/stretchr/testify/assert"
)

func TestEnterMaintenanceModeSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceEnterMaintenanceMode(false).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ServiceEnterMaintenanceMode(false)

	assert.NoError(t, err)
}

func TestEnterMaintenanceModeFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceEnterMaintenanceMode(false).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ServiceEnterMaintenanceMode(false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestExitMaintenanceModeSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceExitMaintenanceMode().
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ServiceExitMaintenanceMode()

	assert.NoError(t, err)
}

func TestExitMaintenanceModeFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceExitMaintenanceMode().
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ServiceExitMaintenanceMode()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestServiceAddDriveSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	op := "start"
	drive := "/dev/testdev"
	journal := false
	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceAddDrive(op, drive, journal).
		Return("Success", nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	status, err := restClient.ServiceAddDrive(op, drive, journal)

	assert.NoError(t, err)
	assert.Contains(t, status, "Success")
}

func TestServiceAddDriveFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	op := "start"
	drive := "/dev/testdev"
	journal := false
	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceAddDrive(op, drive, journal).
		Return("", fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	_, err = restClient.ServiceAddDrive(op, drive, journal)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestServiceReplaceDriveSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	op := "start"
	source := "/dev/testdev1"
	dest := "/dev/testdev2"
	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceReplaceDrive(op, source, dest).
		Return("Success", nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	status, err := restClient.ServiceReplaceDrive(op, source, dest)

	assert.NoError(t, err)
	assert.Contains(t, status, "Success")
}

func TestServiceReplaceDriveFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	op := "start"
	src := "/dev/testdev1"
	dest := "/dev/testdev2"
	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceReplaceDrive(op, src, dest).
		Return("", fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	_, err = restClient.ServiceReplaceDrive(op, src, dest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestServiceRebalancePoolSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	op := "start"
	poolID := 0
	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceRebalancePool(op, poolID).
		Return("In progress", nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	status, err := restClient.ServiceRebalancePool(op, poolID)

	assert.NoError(t, err)
	assert.Contains(t, status, "In progress")
}

func TestServiceRebalancePoolFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	op := "start"
	poolID := 0
	// mock the cluster schedulePolicy response
	tc.MockClusterService().
		EXPECT().
		ServiceRebalancePool(op, poolID).
		Return("", fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	_, err = restClient.ServiceRebalancePool(op, poolID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}
