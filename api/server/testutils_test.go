package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/libopenstorage/openstorage/api"
	mockapi "github.com/libopenstorage/openstorage/api/mock"
	servermock "github.com/libopenstorage/openstorage/api/server/mock"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	"github.com/libopenstorage/openstorage/cluster"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/config"
	sdkauth "github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/loadbalancer"
	"github.com/libopenstorage/openstorage/pkg/role"
	"github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	"github.com/libopenstorage/openstorage/volume/drivers/fake"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/mock"
	schedopsk8s "github.com/portworx/sched-ops/k8s/core"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	testSharedSecret = "mysecret"
	testSdkSock      = "/tmp/sdk.sock"
	testMgmtPort     = uint16(11111)
	testMgmtBase     = "/tmp"
	testMockURL      = "http://localhost:11111"
	mockDriverName   = "mock"
	version          = "v1"
	fakeWithSched    = "fake-sched"
)

var (
	cm     cluster.Cluster
	credId string
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service and REST server
type testServer struct {
	conn                    *grpc.ClientConn
	m                       *mockdriver.MockVolumeDriver
	c                       cluster.Cluster
	k8sops                  *servermock.MockOps
	originalOps             schedopsk8s.Ops
	s                       *mockapi.MockOpenStoragePoolServer
	mc                      *gomock.Controller
	sdk                     *sdk.Server
	rm                      role.RoleManager
	port                    string
	gwport                  string
	overrideSchedDriverName string
}

func (s *testServer) Stop() {
	s.conn.Close()
	s.sdk.Stop()

	// Check mocks
	s.mc.Finish()

	// Remove from registry
	volumedrivers.Remove(mockDriverName)

	if s.overrideSchedDriverName != "" {
		OverrideSchedDriverName = s.overrideSchedDriverName
	}
}

func (s *testServer) setPorts() {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	port := r.Intn(20000) + 10000

	s.port = fmt.Sprintf("%d", port)
	s.gwport = fmt.Sprintf("%d", port+1)
}

func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) DisableGuestAccess() error {
	_, err := s.rm.Update(context.Background(), &api.SdkRoleUpdateRequest{
		Role: &api.SdkRole{
			Name: "system.guest",
			Rules: []*api.SdkRule{
				{
					Services: []string{"!*"},
					Apis:     []string{"!*"},
				},
			},
		},
	})
	return err
}

// Struct used for creation and setup of cluster api testing
type testCluster struct {
	c       *mockcluster.MockCluster
	mc      *gomock.Controller
	oldInst func() (cluster.Cluster, error)
}

func (c *testCluster) MockCluster() *mockcluster.MockCluster {
	return c.c
}

func (c *testCluster) Finish() {
	clustermanager.Inst = c.oldInst
	c.mc.Finish()
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
	kv, err := kvdb.New(mem.Name, "fake_test", []string{}, nil, kvdb.LogFatalErrorCB)
	if err != nil {
		logrus.Panicf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}

	// Clear previous cluster
	clustermanager.ClearInst()

	// Need to setup a fake cluster. No need to start it.
	err = clustermanager.Init(config.ClusterConfig{
		ClusterId: "fakecluster",
		NodeId:    "fakeNode",
	})
	if err != nil {
		logrus.Panicf("Unable to initialize the cluster manager")
	}

	cm, err = clustermanager.Inst()
	if err != nil {
		logrus.Panicf("Unable to initialize cluster manager: %v", err)
	}

	// Restart fake driver
	volumedrivers.Remove("fake")
	err = volumedrivers.Add("fake", fake.Init)
	if err != nil {
		logrus.Panicf("Unable to initialize fake driver")
	}
	if err := volumedrivers.Register("fake", map[string]string{}); err != nil {
		logrus.Panicf("Unable to start volume driver fake: %v", err)
	}

	// Restart fakesched driver
	volumedrivers.Remove(fakeWithSched)
	err = volumedrivers.Add(fakeWithSched, fake.Init)
	if err != nil {
		logrus.Panicf("Unable to initialize fake-sched driver")
	}
	if err := volumedrivers.Register(fakeWithSched, map[string]string{}); err != nil {
		logrus.Panicf("Unable to start volume driver fake: %v", err)
	}
}

func newTestServerSdkNoAuth(t *testing.T) *testServer {
	tester := &testServer{}
	tester.setPorts()
	setupFakeDriver()

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)
	tester.k8sops = servermock.NewMockOps(tester.mc)

	tester.originalOps = schedopsk8s.Instance()
	schedopsk8s.SetInstance(tester.k8sops)
	tester.s = mockapi.NewMockOpenStoragePoolServer(tester.mc)

	kv, err := kvdb.New(mem.Name, "test", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)
	kvdb.SetInstance(kv)
	stp, err := storagepolicy.Init()
	if err != nil {
		stp, _ = storagepolicy.Inst()
	}
	assert.NotNil(t, stp)

	os.Remove(testSdkSock)
	tester.sdk, err = sdk.New(&sdk.ServerConfig{
		DriverName:         "fake",
		Net:                "tcp",
		Address:            ":" + tester.port,
		RestPort:           tester.gwport,
		StoragePolicy:      stp,
		Cluster:            tester.c,
		Socket:             testSdkSock,
		AccessOutput:       ioutil.Discard,
		AuditOutput:        ioutil.Discard,
		RoundRobinBalancer: loadbalancer.NewNullBalancer(),
	})
	assert.Nil(t, err)
	err = tester.sdk.Start()
	assert.Nil(t, err)

	// Register the drivers with SDK
	// The tests use "fake" and "mock" both interchangeably
	// Some of the test set the UserAgent in the REST client to mock
	fakeDriver, err := volumedrivers.Get(fake.Name)
	assert.NoError(t, err)

	driverMap := map[string]volume.VolumeDriver{
		fake.Name:             fakeDriver,
		sdk.DefaultDriverName: fakeDriver,
		mockDriverName:        fakeDriver,
	}
	tester.sdk.UseVolumeDrivers(driverMap)

	// Setup a connection to the driver
	tester.conn, err = grpcserver.Connect("localhost:"+tester.port, []grpc.DialOption{grpc.WithInsecure()})
	assert.Nil(t, err)

	return tester
}

func newTestServerSdk(t *testing.T) *testServer {
	tester := &testServer{}
	tester.setPorts()
	setupFakeDriver()

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)
	tester.k8sops = servermock.NewMockOps(tester.mc)

	tester.originalOps = schedopsk8s.Instance()
	schedopsk8s.SetInstance(tester.k8sops)
	tester.s = mockapi.NewMockOpenStoragePoolServer(tester.mc)

	// Create a role manager
	kv, err := kvdb.New(mem.Name, "test", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)
	kvdb.SetInstance(kv)
	rm, err := role.NewSdkRoleManager(kv)
	assert.NoError(t, err)
	tester.rm = rm

	// Do not check for error, just initialize it
	stp, err := storagepolicy.Init()
	if err != nil {
		stp, _ = storagepolicy.Inst()
	}
	assert.NotNil(t, stp)

	os.Remove(testSdkSock)
	selfsignedJwt, err := sdkauth.NewJwtAuth(&sdkauth.JwtAuthConfig{
		SharedSecret:  []byte(testSharedSecret),
		UsernameClaim: sdkauth.UsernameClaimTypeName,
	})
	assert.NoError(t, err)
	tester.sdk, err = sdk.New(&sdk.ServerConfig{
		DriverName:         "fake",
		Net:                "tcp",
		Address:            ":" + tester.port,
		RestPort:           tester.gwport,
		Cluster:            tester.c,
		Socket:             testSdkSock,
		StoragePolicy:      stp,
		AccessOutput:       ioutil.Discard,
		AuditOutput:        ioutil.Discard,
		RoundRobinBalancer: loadbalancer.NewNullBalancer(),
		Security: &sdk.SecurityConfig{
			Role: rm,
			Authenticators: map[string]sdkauth.Authenticator{
				"testcode": selfsignedJwt,
			},
		},
	})
	assert.Nil(t, err)
	err = tester.sdk.Start()
	assert.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpcserver.Connect("localhost:"+tester.port, []grpc.DialOption{grpc.WithInsecure()})
	assert.Nil(t, err)

	// Create credential for cloudBackup testing
	credentials := api.NewOpenStorageCredentialsClient(tester.conn)
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	resp, err := credentials.Create(ctx, &api.SdkCredentialCreateRequest{
		Name: "goodCred",
		CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
			AwsCredential: &api.SdkAwsCredentialRequest{
				AccessKey: "dummy-access",
				SecretKey: "dummy-secret",
				Endpoint:  "dummy-endpoint",
				Region:    "dummy-region",
			},
		},
	})
	assert.NoError(t, err)
	credId = resp.GetCredentialId()

	// Register the drivers with SDK
	// The tests use "fake" and "mock" both interchangeably
	// Some of the test set the UserAgent in the REST client to mock
	fakeDriver, err := volumedrivers.Get(fake.Name)
	assert.NoError(t, err)

	driverMap := map[string]volume.VolumeDriver{
		fake.Name:             fakeDriver,
		fakeWithSched:         fakeDriver,
		sdk.DefaultDriverName: fakeDriver,
		mockDriverName:        fakeDriver,
	}
	tester.sdk.UseVolumeDrivers(driverMap)

	// Override sched driver name
	tester.overrideSchedDriverName = OverrideSchedDriverName
	OverrideSchedDriverName = "fake"

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
	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, kvdb.LogFatalErrorCB)
	assert.NoError(t, err)
	kvdb.SetInstance(kv)
	_, err = storagepolicy.Init()

	return ts, testVolDriver
}

func testRestServerSdkNoAuth(t *testing.T) (*httptest.Server, *testServer) {
	os.Remove(testSdkSock)
	testVolDriver := newTestServerSdkNoAuth(t)

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
	return ts, testVolDriver
}

func testRestServerSdk(t *testing.T) (*httptest.Server, *testServer) {
	os.Remove(testSdkSock)
	testVolDriver := newTestServerSdk(t)

	vapi := newVolumeAPI("fake", testSdkSock)
	router := mux.NewRouter()
	// Register all routes from the App
	for _, route := range vapi.Routes() {
		router.Methods(route.verb).
			Path(route.path).
			Name(mockDriverName).
			Handler(http.HandlerFunc(route.fn))
	}

	ts := httptest.NewServer(router)
	return ts, testVolDriver
}

func testServerRegisterRoute(
	routeFunc func(w http.ResponseWriter, r *http.Request),
	preRouteCheck func(w http.ResponseWriter, r *http.Request) bool,
) func(w http.ResponseWriter, r *http.Request) {
	return routeFunc
}

func testpreRouteCheck(http.ResponseWriter, *http.Request) bool {
	return true
}

func testRestServerSdkWithAuthMw(t *testing.T) (*httptest.Server, *testServer) {
	os.Remove(testSdkSock)
	testVolDriver := newTestServerSdk(t)

	vapi := newVolumeAPI("fake", testSdkSock)
	router := mux.NewRouter()

	// Add the Middleware
	router, err := GetVolumeAPIRoutesWithAuth("pxd",
		testSdkSock,
		mux.NewRouter(),
		testServerRegisterRoute,
		testpreRouteCheck)
	assert.NoError(t, err)

	// Register all routes from the App
	for _, route := range vapi.Routes() {
		router.Methods(route.verb).
			Path(route.path).
			Name(mockDriverName).
			Handler(http.HandlerFunc(route.fn))
	}

	ts := httptest.NewServer(router)
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

func createToken(name, role, secret string) (string, error) {
	claims := &sdkauth.Claims{
		Issuer: "testcode",
		Name:   name,
		Email:  name + "@openstorage.org",
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

func getSecretsMock(t *testing.T) (secrets.Secrets, *mock.MockSecrets, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockSecret := mock.NewMockSecrets(mockCtrl)
	return mockSecret, mockSecret, mockCtrl
}
