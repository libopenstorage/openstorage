package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kubernetes-csi/csi-test/utils"

	"github.com/golang/mock/gomock"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/openstorage/cluster"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	policy "github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const (
	mockDriverName = "mock"
	version        = "v1"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service and REST server
type testServer struct {
	m  *mockdriver.MockVolumeDriver
	mc *gomock.Controller
}

// Struct used for creation and setup of cluster api testing
type testCluster struct {
	c       *mockcluster.MockCluster
	mc      *gomock.Controller
	oldInst func() (cluster.Cluster, error)
}

func newTestCluster(t *testing.T) *testCluster {
	tester := &testCluster{}

	// Save already set value of cluster.Inst to set it back
	// when we finish the tests by the defer()
	tester.oldInst = clustermanager.Inst

	// Create mock controller
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})

	// Create a new mock cluster
	tester.c = mockcluster.NewMockCluster(tester.mc)

	// Override cluster.Inst to return our mock cluster
	clustermanager.Inst = func() (cluster.Cluster, error) {
		return tester.c, nil
	}

	return tester
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)

	setupMockDriver(tester, t)
	return tester
}

func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

func setupMockDriver(tester *testServer, t *testing.T) {
	volumedrivers.Add(mockDriverName, func(map[string]string) (volume.VolumeDriver, error) {
		return tester.m, nil
	})

	var err error

	// Register mock driver
	err = volumedrivers.Register(mockDriverName, nil)
	assert.Nil(t, err)
}

func testRestServer(t *testing.T) (*httptest.Server, *testServer) {
	vapi := &volAPI{}
	router := mux.NewRouter()
	// Register all routes from the App
	for _, route := range vapi.Routes() {
		router.Methods(route.verb).
			Path(route.path).
			Name(mockDriverName).
			Handler(http.HandlerFunc(route.fn))
	}

	ts := httptest.NewServer(router)
	testVolDriver := newTestServer(t)
	// Initialise storage policy manager
	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)
	policy.Init(kv)
	// Disable policy enforcement for regular vol_ops test
	_, err = policy.Inst()
	assert.NoError(t, err)

	return ts, testVolDriver
}

func testClusterServer(t *testing.T) (*httptest.Server, *testCluster) {
	tc := newTestCluster(t)
	capi := newClusterAPI()
	router := mux.NewRouter()
	// Register all routes from the App
	for _, route := range capi.Routes() {
		router.Methods(route.verb).
			Path(route.path).
			Name(mockDriverName).
			Handler(http.HandlerFunc(route.fn))
	}

	ts := httptest.NewServer(router)
	return ts, tc
}

func (c *testCluster) MockCluster() *mockcluster.MockCluster {
	return c.c
}

func (c *testCluster) Finish() {
	clustermanager.Inst = c.oldInst
	c.mc.Finish()
}

func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove(mockDriverName)

	// Check mocks
	s.mc.Finish()
}
