package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/api/server"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/stretchr/testify/assert"
	"go.pedge.io/dlog"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"

	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

// testServer is a simple struct used abstract
// the creation and setup of mock server
type testServer struct {
	client *client.Client
	m      *mockdriver.MockVolumeDriver
	c      *mockcluster.MockCluster
	mc     *gomock.Controller
}

const (
	host       string = "http://127.0.0.1:2376"
	mgmtPort   uint16 = 2376
	pluginPort uint16 = 2377
	driver     string = "mock"
	version    string = "v1"
)

// Init function to setup the http server

func init() {
	startServer()
}

func startServer() {
	err := server.StartVolumeMgmtAPI(
		driver,
		volume.DriverAPIBase,
		mgmtPort,
	)

	if err != nil {
		dlog.Errorf("Error starting the server")
	}

	// adding sleep to avoid race condition of connection refused.
	time.Sleep(1 * time.Second)
}

func setupMocks() *testServer {

	var ts = &testServer{}

	// Add driver to registry
	ts.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	ts.m = mockdriver.NewMockVolumeDriver(ts.mc)
	ts.c = mockcluster.NewMockCluster(ts.mc)

	err := volumedrivers.Add(driver, func(map[string]string) (volume.VolumeDriver, error) {
		return ts.m, nil
	})

	if err != nil {
		dlog.Errorf("Failed to add the driver [%s] for tests", driver)
	}

	// Register the mock driver
	err = volumedrivers.Register(driver, nil)

	return ts
}

func TestServerStart(t *testing.T) {

	ts := setupMocks()
	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)
	assert.NotNil(t, ts.client)
}

func TestVolumeCreateSuccess(t *testing.T) {

	var err error
	ts := setupMocks()
	defer ts.Stop()

	// create a request
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)
	assert.NotNil(t, ts.client)

	// Setup request
	name := "myvol"
	size := uint64(1234)

	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec:    &api.VolumeSpec{Size: size},
	}

	// Setup mock functions
	id := "myid"
	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil),
	)

	// create a volume client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())

	assert.Nil(t, err)
	assert.Equal(t, id, res)
}

func TestVolumeCreateFailed(t *testing.T) {
	var err error

	ts := setupMocks()
	defer ts.Stop()

	// create a request
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)
	assert.NotNil(t, ts.client)

	// Setup mock functions
	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Create(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("error in create")),
	)

	req := &api.VolumeCreateRequest{}

	// create a volume client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())

	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestVolumeDeleteSuccess(t *testing.T) {
	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	// Setup mock
	id := "myid"

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Delete(id).
			Return(nil).
			Times(1),
	)

	driverclient := volumeclient.VolumeDriver(ts.client)

	err = driverclient.Delete(id)
	assert.Nil(t, err)
}

func TestVolumeDeleteFailed(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	// Setup mock

	id := "myid"

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Delete(gomock.Any()).
			Return(fmt.Errorf("error in delete")).
			Times(1),
	)

	driverclient := volumeclient.VolumeDriver(ts.client)

	err = driverclient.Delete(id)
	assert.NotNil(t, err)
}

func TestSnapshotCreateSuccess(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	id := "myid"
	name := "snapName"

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	req := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: name},
		Readonly: true,
	}

	//mock Snapshot call
	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Snapshot(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(id, nil).
			Times(1),
	)

	//make the call
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Snapshot(req.GetId(), req.GetReadonly(), req.GetLocator())

	assert.Nil(t, err)
	assert.Equal(t, id, res)

}

func TestVolumeInspectSuccess(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	id := "myid"
	var size uint64 = 1234
	name := "inspectVol"

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil).
			Times(1),
	)

	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Inspect([]string{id})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res)
	assert.Equal(t, res[0].GetId(), id)

}

func TestVolumeSetSuccess(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	// create a volume request

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:  api.VolumeActionParam_VOLUME_ACTION_PARAM_ON},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Set(id, req.GetLocator(), req.GetSpec()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// create driver client

	driverclient := volumeclient.VolumeDriver(ts.client)

	res := driverclient.Set(id, req.GetLocator(), req.GetSpec())
	assert.Nil(t, res)

}

func TestVolumeAttachSuccess(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Attach(id, gomock.Any()).
			Return("", nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// create driver client

	driverclient := volumeclient.VolumeDriver(ts.client)

	_, err = driverclient.Attach(id, req.GetOptions())

	assert.Nil(t, err)

}

func TestVolumeAttachFailed(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Attach(id, gomock.Any()).
			Return("", fmt.Errorf("some error")),
	)

	// create driver client

	driverclient := volumeclient.VolumeDriver(ts.client)

	_, err = driverclient.Attach(id, req.GetOptions())

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "some error")

}

func TestVolumeDetachSuccess(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Detach(id, gomock.Any()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// create client

	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Detach(id, req.GetOptions())

	assert.Nil(t, res)
}

func TestVolumeDetachFailed(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Detach(id, gomock.Any()).
			Return(fmt.Errorf("Error in detaching")),
	)

	// create client

	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Detach(id, req.GetOptions())

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "Error in detaching")
}

func TestVolumeMountSuccess(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error

	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")
	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	//create request
	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach:    api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			MountPath: "/mnt"},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(

		ts.MockDriver().
			EXPECT().
			Mount(id, gomock.Any(), gomock.Any()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	//create driverclient

	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Mount(id, req.GetAction().GetMountPath(), req.GetOptions())
	assert.Nil(t, res)
}

func TestVolumeMountFailedNoMountPath(t *testing.T) {

	ts := setupMocks()

	defer ts.Stop()

	var err error

	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")
	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	//create request
	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach:    api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			MountPath: ""},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	//create driverclient

	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Mount(id, req.GetAction().GetMountPath(), req.GetOptions())
	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "Invalid mount path")
}

func TestVolumeStatsSuccess(t *testing.T) {

	ts := setupMocks()
	defer ts.Stop()

	var err error
	ts.client, err = volumeclient.NewDriverClient(host, driver, version, "")
	assert.Nil(t, err)

	bytesUsed := uint64(1234)
	writeBytes := uint64(1234)

	id := "myid"
	//req := &api.Stats{BytesUsed: bytesUsed}

	ts.MockDriver().
		EXPECT().
		Stats(id, gomock.Any()).
		Return(
			&api.Stats{
				BytesUsed:  bytesUsed,
				WriteBytes: writeBytes},
			nil)

	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Stats(id, true)

	assert.Nil(t, err)
	assert.Equal(t, bytesUsed, res.BytesUsed)

}

// MockDriver helper method.
func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

// MockCluster helper method.
func (s *testServer) MockCluster() *mockcluster.MockCluster {
	return s.c
}

// Stop method to to remove the driver and check mocks.
func (s *testServer) Stop() {
	// Remove from registry
	volumedrivers.Remove("mock")
	// Check mocks
	s.mc.Finish()
}
