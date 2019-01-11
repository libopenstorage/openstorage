package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/sirupsen/logrus"

	"github.com/golang/mock/gomock"
	sdkauth "github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/stretchr/testify/assert"

	"github.com/libopenstorage/openstorage/api/server/sdk"
	"github.com/libopenstorage/openstorage/cluster"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/role"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	testSharedSecret = "mysecret"
	testSdkSock      = "/tmp/sdk.sock"
	mockDriverName   = "mock"
	version          = "v1"
)

var (
	cm cluster.Cluster
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service and REST server
type testServer struct {
	conn *grpc.ClientConn
	m    *mockdriver.MockVolumeDriver
	c    cluster.Cluster
	mc   *gomock.Controller
	sdk  *sdk.Server
}

// Struct used for creation and setup of cluster api testing
type testCluster struct {
	c       *mockcluster.MockCluster
	mc      *gomock.Controller
	oldInst func() (cluster.Cluster, error)
}

func init() {
	setupFakeDriver()
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

func setupFakeDriver() {
	kv, err := kvdb.New(mem.Name, "fake_test", []string{}, nil, logrus.Panicf)
	if err != nil {
		logrus.Panicf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}
	// Need to setup a fake cluster. No need to start it.
	clustermanager.Init(config.ClusterConfig{
		ClusterId: "fakecluster",
		NodeId:    "fakeNode",
	})
	cm, err = clustermanager.Inst()
	if err != nil {
		logrus.Panicf("Unable to initialize cluster manager: %v", err)
	}

	// Requires a non-nil cluster
	if err := volumedrivers.Register("fake", map[string]string{}); err != nil {
		logrus.Panicf("Unable to start volume driver fake: %v", err)
	}
}

func newTestServerSdk(t *testing.T) *testServer {
	tester := &testServer{}

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)

	// Create a role manager
	kv, err := kvdb.New(mem.Name, "role", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)
	rm, err := role.NewSdkRoleManager(kv)
	assert.NoError(t, err)

	os.Remove(testSdkSock)
	selfsignedJwt, err := auth.NewJwtAuth(&auth.JwtAuthConfig{
		SharedSecret: []byte(testSharedSecret),
	})
	assert.NoError(t, err)
	tester.sdk, err = sdk.New(&sdk.ServerConfig{
		DriverName:   "fake",
		Net:          "tcp",
		Address:      ":8123",
		RestPort:     "8124",
		Cluster:      tester.c,
		Socket:       testSdkSock,
		AccessOutput: ioutil.Discard,
		AuditOutput:  ioutil.Discard,
		Security: &sdk.SecurityConfig{
			Role: rm,
			Authenticators: map[string]auth.Authenticator{
				"testcode": selfsignedJwt,
			},
		},
	})
	assert.Nil(t, err)
	err = tester.sdk.Start()
	assert.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpcserver.Connect("localhost:8123", []grpc.DialOption{grpc.WithInsecure()})
	assert.Nil(t, err)

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

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
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
	return ts, testVolDriver
}

func testRestServerSdk(t *testing.T) (*httptest.Server, *testServer) {
	vapi := newVolumeAPI(mockDriverName, testSdkSock)
	router := mux.NewRouter()
	// Register all routes from the App
	for _, route := range vapi.Routes() {
		router.Methods(route.verb).
			Path(route.path).
			Name(mockDriverName).
			Handler(http.HandlerFunc(route.fn))
	}

	ts := httptest.NewServer(router)
	testVolDriver := newTestServerSdk(t)
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
	s.conn.Close()
	s.sdk.Stop()

	// Check mocks
	s.mc.Finish()

	// Remove from registry
	volumedrivers.Remove(mockDriverName)
}

func createToken(name, role, secret string) (string, error) {
	claims := &sdkauth.Claims{
		Issuer: "testcode",
		Name:   name,
		Email:  "test@openstorage",
		Roles:  []string{role},
	}
	signature := &sdkauth.Signature{
		Key:  []byte(secret),
		Type: jwt.SigningMethodHS256,
	}
	options := &sdkauth.Options{
		Expiration: time.Now().Add(1 * time.Hour).Unix(),
	}
	return sdkauth.Token(claims, signature, options)
}

func contextWithToken(ctx context.Context, name, role, secret string) (context.Context, error) {
	token, err := createToken(name, role, secret)
	if err != nil {
		return nil, err
	}
	md := metadata.New(map[string]string{
		"authorization": "bearer " + token,
	})
	return metadata.NewOutgoingContext(ctx, md), nil
}
