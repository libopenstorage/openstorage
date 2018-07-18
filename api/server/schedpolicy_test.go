package server

import (
	"fmt"
	"testing"

	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	sched "github.com/libopenstorage/openstorage/schedpolicy"
	"github.com/stretchr/testify/assert"
)

func TestSchedPolicyCreateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	schedule := "freq:periodic\nperiod:120000\n"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyCreate(name, schedule).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyCreate(name, schedule)

	assert.NoError(t, err)
}

func TestSchedPolicyCreateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	schedule := "freq:periodic\nperiod:120000\n"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyCreate(name, schedule).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyCreate(name, schedule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSchedPolicyUpdateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	schedule := "freq:periodic\nperiod:120000\n"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyUpdate(name, schedule).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyUpdate(name, schedule)

	assert.NoError(t, err)
}

func TestSchedPolicyUpdateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	schedule := "freq:periodic\nperiod:120000\n"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyUpdate(name, schedule).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyUpdate(name, schedule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSchedPolicyDeleteSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyDelete(name).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyDelete(name)

	assert.NoError(t, err)
}

func TestSchedPolicyDeleteWithCaseSensitiveName(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "TestSchedPolicy1"
	// mock the cluster schedulePolicy response
	// this should be recived it as "TestSchedPolicy1" only
	tc.MockCluster().
		EXPECT().
		SchedPolicyDelete(name).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyDelete(name)

	assert.NoError(t, err)
}

func TestSchedPolicyDeleteFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyDelete(name).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.SchedPolicyDelete(name)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSchedPolicyEnumerateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyEnumerate().
		Return([]*sched.SchedPolicy{
			&sched.SchedPolicy{
				Name:     name,
				Schedule: "testsche:test",
			},
		}, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	schedPolicy, err := restClient.SchedPolicyEnumerate()

	assert.NotNil(t, schedPolicy)
	assert.EqualValues(t, schedPolicy[0].Name, name)
	assert.NoError(t, err)
}

func TestSchedPolicyEnumerateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyEnumerate().
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	schedPolicy, err := restClient.SchedPolicyEnumerate()

	assert.Error(t, err)
	assert.Nil(t, schedPolicy)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSchedPolicyGetSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp1"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyGet(name).
		Return(&sched.SchedPolicy{
			Name:     name,
			Schedule: "testsched:test",
		}, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	schedPolicy, err := restClient.SchedPolicyGet(name)

	assert.NotNil(t, schedPolicy)
	assert.EqualValues(t, schedPolicy.Name, name)
	assert.NoError(t, err)
}

func TestSchedPolicyGetSuccessWithSensitiveName(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testSchedPolicy@Periodic1"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyGet(name).
		Return(&sched.SchedPolicy{
			Name:     name,
			Schedule: "freq:periodic\nperiod:120000\n",
		}, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	schedPolicy, err := restClient.SchedPolicyGet(name)

	assert.NotNil(t, schedPolicy)
	assert.EqualValues(t, schedPolicy.Name, name)
	assert.NoError(t, err)
}

func TestSchedPolicyGetFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testsp"
	// mock the cluster schedulePolicy response
	tc.MockCluster().
		EXPECT().
		SchedPolicyGet(name).
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	schedPolicy, err := restClient.SchedPolicyGet(name)

	assert.Error(t, err)
	assert.Nil(t, schedPolicy)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestSchedPolicyGetFailedBadParam(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := ""

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	schedPolicy, err := restClient.SchedPolicyGet(name)

	assert.Error(t, err)
	assert.Nil(t, schedPolicy)
	assert.Contains(t, err.Error(), "Missing policy name")
}
