package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
)

type testCluster struct {
	c       *mockcluster.MockCluster
	mc      *gomock.Controller
	oldInst func() (cluster.Cluster, error)
}

func newTestClutser(t *testing.T) *testCluster {
	tester := &testCluster{}

	// Save already set value of cluster.Inst to set it back
	// when we finish the tests by the defer()
	tester.oldInst = cluster.Inst

	// Create mock controller
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})

	// Create a new mock cluster
	tester.c = mockcluster.NewMockCluster(tester.mc)

	// Override cluster.Inst to return our mock cluster
	cluster.Inst = func() (cluster.Cluster, error) {
		return tester.c, nil
	}

	return tester
}

func (c *testCluster) MockCluster() *mockcluster.MockCluster {
	return c.c
}

func (c *testCluster) Finish() {
	cluster.Inst = c.oldInst
	c.mc.Finish()
}

func TestServerNodeStatus(t *testing.T) {

	// Create a new global test cluster
	c := newTestClutser(t)
	defer c.Finish()

	// Create an instance of clusterAPI to get access to
	// nodeStatus receiver
	capi := &clusterApi{}

	// Send call to server
	ts := httptest.NewServer(http.HandlerFunc(capi.nodeStatus))
	restClient, err := client.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// Set expections
	c.MockCluster().
		EXPECT().
		NodeStatus().
		Return(api.Status_STATUS_OK, nil).
		Times(1)

	// Check status
	status, err := client.ClusterManager(restClient).NodeStatus()
	assert.NoError(t, err)
	assert.Equal(t, api.Status_STATUS_OK, status)
}
