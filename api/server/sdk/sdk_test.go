/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sdk

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

	jwt "github.com/dgrijalva/jwt-go"
	bucket "github.com/libopenstorage/openstorage/bucket"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/role"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/libopenstorage/openstorage/alerts"
	mockalerts "github.com/libopenstorage/openstorage/alerts/mock"
	"github.com/libopenstorage/openstorage/api"
	mockbucketdriver "github.com/libopenstorage/openstorage/bucket/drivers/mock"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/loadbalancer"
	policy "github.com/libopenstorage/openstorage/pkg/storagepolicy"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	mockDriverName   = "mock"
	testUds          = "/tmp/sdk-test.sock"
	testSharedSecret = "mysecret"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service
type testServer struct {
	conn   *grpc.ClientConn
	server *Server
	m      *mockdriver.MockVolumeDriver
	b      *mockbucketdriver.MockBucketDriver
	c      *mockcluster.MockCluster
	a      *mockalerts.MockFilterDeleter
	mc     *gomock.Controller
	gw     *httptest.Server
	port   string
	gwport string
}

func init() {
	logrus.SetLevel(logrus.PanicLevel)
}

func setupMockDriver(tester *testServer, t *testing.T) {
	volumedrivers.Add(mockDriverName, func(map[string]string) (volume.VolumeDriver, error) {
		return tester.m, nil
	})

	var err error

	// Register mock driver
	err = volumedrivers.Register(mockDriverName, nil)
	require.Nil(t, err)
}

func setupMockBucketDriver(tester *testServer, t *testing.T) {
	var err error
	driverMap := make(map[string]bucket.BucketDriver)
	driverMap[DefaultDriverName] = tester.b
	tester.server.UseBucketDrivers(driverMap)
	require.Nil(t, err)
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}
	tester.setPorts()

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.b = mockbucketdriver.NewMockBucketDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)
	tester.a = mockalerts.NewMockFilterDeleter(tester.mc)

	setupMockDriver(tester, t)

	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, kvdb.LogFatalErrorCB)
	require.NoError(t, err)
	kvdb.SetInstance(kv)
	// Init storage policy manager
	_, err = policy.Init()
	sp, err := policy.Inst()
	require.NotNil(t, sp)

	// Setup simple driver
	os.Remove(testUds)
	tester.server, err = New(&ServerConfig{
		DriverName:          mockDriverName,
		Net:                 "tcp",
		Address:             ":" + tester.port,
		RestPort:            tester.gwport,
		Socket:              testUds,
		Cluster:             tester.c,
		StoragePolicy:       sp,
		AlertsFilterDeleter: tester.a,
		AccessOutput:        ioutil.Discard,
		AuditOutput:         ioutil.Discard,
		Security: &SecurityConfig{
			Tls: &TLSConfig{
				CertFile: "test_certs/server.crt",
				KeyFile:  "test_certs/server.key",
			},
		},
		RoundRobinBalancer: loadbalancer.NewNullBalancer(),
	})

	require.Nil(t, err)

	tester.m.EXPECT().StartVolumeWatcher().Return().Times(1)
	tester.m.EXPECT().GetVolumeWatcher(&api.VolumeLocator{}, make(map[string]string)).DoAndReturn(func(a *api.VolumeLocator, l map[string]string) (chan *api.Volume, error) {
		ch := make(chan *api.Volume, 1)
		tester.server.watcherCtxCancel()
		return ch, nil
	}).Times(1)

	err = tester.server.Start()
	require.Nil(t, err)

	// Read the CA cert data
	caCertdata, err := ioutil.ReadFile("test_certs/insecure_ca.crt")
	require.Nil(t, err)

	// Get TLS dial options
	dopts, err := grpcserver.GetTlsDialOptions(caCertdata)
	require.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpcserver.Connect("localhost:"+tester.port, dopts)
	require.Nil(t, err)

	// Setup REST gateway
	mux, err := tester.server.restGateway.restServerSetupHandlers()
	require.NoError(t, err)
	require.NotNil(t, mux)
	tester.gw = httptest.NewServer(mux)

	// Add mock bucket driver to the server
	setupMockBucketDriver(tester, t)

	return tester
}

func newTestServerAuth(t *testing.T) *testServer {
	tester := &testServer{}
	tester.setPorts()

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)
	tester.c = mockcluster.NewMockCluster(tester.mc)
	tester.a = mockalerts.NewMockFilterDeleter(tester.mc)

	setupMockDriver(tester, t)

	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, kvdb.LogFatalErrorCB)
	require.NoError(t, err)
	kvdb.SetInstance(kv)
	// Init storage policy manager
	_, err = policy.Init()
	sp, err := policy.Inst()
	require.NotNil(t, sp)

	rm, err := role.NewSdkRoleManager(kv)
	require.NoError(t, err)

	selfsignedJwt, err := auth.NewJwtAuth(&auth.JwtAuthConfig{
		SharedSecret:  []byte(testSharedSecret),
		UsernameClaim: auth.UsernameClaimTypeName,
	})

	// Setup simple driver
	os.Remove(testUds)
	tester.server, err = New(&ServerConfig{
		DriverName:          mockDriverName,
		Net:                 "tcp",
		Address:             ":" + tester.port,
		RestPort:            tester.gwport,
		Socket:              testUds,
		Cluster:             tester.c,
		StoragePolicy:       sp,
		AlertsFilterDeleter: tester.a,
		AccessOutput:        ioutil.Discard,
		AuditOutput:         ioutil.Discard,
		Security: &SecurityConfig{
			Role: rm,
			Tls: &TLSConfig{
				CertFile: "test_certs/server.crt",
				KeyFile:  "test_certs/server.key",
			},
			Authenticators: map[string]auth.Authenticator{
				"testcode": selfsignedJwt,
			},
		},
	})
	require.Nil(t, err)
	tester.m.EXPECT().StartVolumeWatcher().Return().Times(1)
	tester.m.EXPECT().GetVolumeWatcher(&api.VolumeLocator{}, make(map[string]string)).DoAndReturn(func(a *api.VolumeLocator, l map[string]string) (chan *api.Volume, error) {
		ch := make(chan *api.Volume, 1)
		tester.server.watcherCtxCancel()
		return ch, nil
	}).Times(1)

	err = tester.server.Start()
	require.Nil(t, err)

	// Read the CA cert data
	caCertdata, err := ioutil.ReadFile("test_certs/insecure_ca.crt")
	require.Nil(t, err)

	// Get TLS dial options
	dopts, err := grpcserver.GetTlsDialOptions(caCertdata)
	require.Nil(t, err)

	// Setup a connection to the driver
	tester.conn, err = grpcserver.Connect("localhost:"+tester.port, dopts)
	require.Nil(t, err)

	// Setup REST gateway
	mux, err := tester.server.restGateway.restServerSetupHandlers()
	require.NoError(t, err)
	require.NotNil(t, mux)
	tester.gw = httptest.NewServer(mux)
	return tester
}

func (s *testServer) setPorts() {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	port := r.Intn(1000) + 30000

	s.port = fmt.Sprintf("%d", port)
	s.gwport = fmt.Sprintf("%d", port+1)
}

func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

func (s *testServer) MockBucketDriver() *mockbucketdriver.MockBucketDriver {
	return s.b
}

func (s *testServer) MockCluster() *mockcluster.MockCluster {
	return s.c
}

func (s *testServer) MockFilterDeleter() *mockalerts.MockFilterDeleter {
	return s.a
}

func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove("mock")

	// Shutdown servers
	if s.conn != nil {
		s.conn.Close()
	}
	s.m.EXPECT().StopVolumeWatcher().Return().AnyTimes()
	if s.server != nil {
		s.server.Stop()
	}
	if s.gw != nil {
		s.gw.Close()
	}

	// Check mocks
	s.mc.Finish()
}

func (s *testServer) Conn() *grpc.ClientConn {
	return s.conn
}

func (s *testServer) Server() grpcserver.Server {
	return s.server.netServer
}

func (s *testServer) UdsServer() grpcserver.Server {
	return s.server.udsServer
}

func (s *testServer) GatewayURL() string {
	return s.gw.URL
}

func createToken(name, role, secret string) (string, error) {
	claims := &auth.Claims{
		Issuer: "testcode",
		Name:   name,
		Email:  name + "@openstorage.org",
		Roles:  []string{role},
	}
	signature := &auth.Signature{
		Key:  []byte(secret),
		Type: jwt.SigningMethodHS256,
	}
	options := &auth.Options{
		Expiration: time.Now().Add(1 * time.Hour).Unix(),
	}
	return auth.Token(claims, signature, options)
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

func TestSdkGateway(t *testing.T) {
	s := newTestServer(t)
	defer s.Stop()

	// Check we can get the swagger.json file
	res, err := http.Get(s.GatewayURL() + "/swagger.json")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Check we get the swagger-ui
	res, err = http.Get(s.GatewayURL() + "/swagger-ui")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Check unhandled address
	res, err = http.Get(s.GatewayURL() + "/this-should-not-work")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	// Check the gateway works
	// First setup the mock
	id := "id"
	name := "name"
	cluster := api.Cluster{
		Id:     name,
		NodeId: "somenodeid",
		Status: api.Status_STATUS_NOT_IN_QUORUM,
	}
	s.MockCluster().EXPECT().Enumerate().Return(cluster, nil).Times(1)
	s.MockCluster().EXPECT().Uuid().Return(id).Times(1)

	// Then send the request
	res, err = http.Get(s.GatewayURL() + "/v1/clusters/inspectcurrent")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Setup mock for CORS request
	s.MockCluster().EXPECT().Enumerate().Return(cluster, nil).Times(1)
	s.MockCluster().EXPECT().Uuid().Return(id).Times(1)

	// Try cross-origin reqeuest, should get allowed
	reqOrigin := "openstorage.io"
	req, err := http.NewRequest("GET", s.GatewayURL()+"/v1/clusters/inspectcurrent", nil)
	require.NoError(t, err)
	req.Header.Add("origin", reqOrigin)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))

}

func TestSdkWithNoVolumeDriverThenAddOne(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "fake_test", []string{}, nil, kvdb.LogFatalErrorCB)
	if err != nil {
		logrus.Panicf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		logrus.Panicf("Failed to set KVDB instance")
	}
	clustermanager.Init(config.ClusterConfig{
		ClusterId: "fakecluster",
		NodeId:    "fakeNode",
	})
	cm, err := clustermanager.Inst()
	startChan := make(chan bool)
	go func() {
		cm.Start(false, "9002", "")
		close(startChan)
	}()
	defer cm.Shutdown()
	<-startChan
	if err := volumedrivers.Register("fake", map[string]string{}); err != nil {
		t.Fatalf("Unable to start volume driver fake: %v", err)
	}

	// Setup SDK Server with no volume driver
	alert, err := alerts.NewFilterDeleter()
	require.NoError(t, err)

	sp, err := policy.Inst()
	os.Remove(testUds)
	tester := &testServer{}
	tester.setPorts()
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})

	server, err := New(&ServerConfig{
		Net:                 "tcp",
		Address:             ":" + tester.port,
		RestPort:            tester.gwport,
		Socket:              testUds,
		Cluster:             cm,
		StoragePolicy:       sp,
		AlertsFilterDeleter: alert,
		AccessOutput:        ioutil.Discard,
		AuditOutput:         ioutil.Discard,
		Security: &SecurityConfig{
			Tls: &TLSConfig{
				CertFile: "test_certs/server.crt",
				KeyFile:  "test_certs/server.key",
			},
		},
	})
	require.Nil(t, err)

	err = server.Start()
	require.Nil(t, err)
	defer func() {
		server.Stop()
	}()

	// Read the CA cert data
	caCertdata, err := ioutil.ReadFile("test_certs/insecure_ca.crt")
	require.Nil(t, err)

	// Get TLS dial options
	dopts, err := grpcserver.GetTlsDialOptions(caCertdata)
	require.Nil(t, err)

	// Setup a connection to the driver
	conn, err := grpc.Dial("localhost:"+tester.port, dopts...)
	require.Nil(t, err)

	// Setup API names that depend on the volume driver
	// To get the names, look at api.pb.go and search for grpc.Invoke or c.cc.Invoke
	apis := []string{
		"/openstorage.api.OpenStorageVolume/Create",
		"/openstorage.api.OpenStorageVolume/Clone",
		"/openstorage.api.OpenStorageVolume/Delete",
		"/openstorage.api.OpenStorageVolume/Inspect",
		"/openstorage.api.OpenStorageVolume/Stats",
		"/openstorage.api.OpenStorageVolume/Enumerate",
		"/openstorage.api.OpenStorageVolume/EnumerateWithFilters",
		"/openstorage.api.OpenStorageVolume/SnapshotCreate",
		"/openstorage.api.OpenStorageVolume/SnapshotRestore",
		"/openstorage.api.OpenStorageVolume/SnapshotEnumerate",
		"/openstorage.api.OpenStorageVolume/SnapshotEnumerateWithFilters",
		"/openstorage.api.OpenStorageVolume/SnapshotScheduleUpdate",
		"/openstorage.api.OpenStorageWatch/Watch",
		"/openstorage.api.OpenStorageMountAttach/Attach",
		"/openstorage.api.OpenStorageMountAttach/Detach",
		"/openstorage.api.OpenStorageMountAttach/Mount",
		"/openstorage.api.OpenStorageMountAttach/Unmount",
		"/openstorage.api.OpenStorageCloudBackup/Create",
		"/openstorage.api.OpenStorageCloudBackup/Restore",
		"/openstorage.api.OpenStorageCloudBackup/Delete",
		"/openstorage.api.OpenStorageCloudBackup/DeleteAll",
		"/openstorage.api.OpenStorageCloudBackup/EnumerateWithFilters",
		"/openstorage.api.OpenStorageCloudBackup/Status",
		"/openstorage.api.OpenStorageCloudBackup/Catalog",
		"/openstorage.api.OpenStorageCloudBackup/History",
		"/openstorage.api.OpenStorageCloudBackup/StateChange",
		"/openstorage.api.OpenStorageCloudBackup/SchedCreate",
		"/openstorage.api.OpenStorageCloudBackup/SchedDelete",
		"/openstorage.api.OpenStorageCloudBackup/SchedEnumerate",
		"/openstorage.api.OpenStorageCredentials/Create",
		"/openstorage.api.OpenStorageCredentials/Enumerate",
		"/openstorage.api.OpenStorageCredentials/Inspect",
		"/openstorage.api.OpenStorageCredentials/Delete",
		"/openstorage.api.OpenStorageCredentials/Validate",
	}

	// The main purpose of this test is to make sure that the server
	// does not panic using a nil point to a driver
	for _, api := range apis {
		err = conn.Invoke(context.Background(), api, nil, nil)
		require.Error(t, err)
		serverError, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, serverError.Code(), codes.Unavailable)
		require.Contains(t, serverError.Message(), "Resource")
	}

	// Check the driver is not loaded
	identities := api.NewOpenStorageIdentityClient(conn)
	id, err := identities.Version(context.Background(), &api.SdkIdentityVersionRequest{})
	require.NoError(t, err)
	require.Contains(t, id.GetVersion().GetDriver(), "no driver")

	// Now add the volume driver
	d, err := volumedrivers.Get("fake")
	require.NoError(t, err)
	driverMap := map[string]volume.VolumeDriver{"fake": d, DefaultDriverName: d}
	server.UseVolumeDrivers(driverMap)

	// Identify that the driver is now running
	id, err = identities.Version(context.Background(), &api.SdkIdentityVersionRequest{})
	require.NoError(t, err)
	require.Equal(t, "fake", id.GetVersion().GetDriver())

	// This part of the test we cannot simply send nils for request and response
	// because real data is being passed. Therefore, a single call will satisfy that
	// the driver is working now.
	volumes := api.NewOpenStorageVolumeClient(conn)
	r, err := volumes.Create(context.Background(), &api.SdkVolumeCreateRequest{
		Name: "myvol",
		Spec: &api.VolumeSpec{
			Size:    uint64(12345),
			HaLevel: 1,
		},
	})
	require.NoError(t, err)
	require.True(t, len(r.GetVolumeId()) != 0)
}
