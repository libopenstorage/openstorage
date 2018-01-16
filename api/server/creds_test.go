package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	client "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
	vol_drivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const (
	mockDriverName = "mock"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service
type testServer struct {
	m  *mockdriver.MockVolumeDriver
	mc *gomock.Controller
}

func setupMockDriver(tester *testServer, t *testing.T) {
	vol_drivers.Add(mockDriverName, func(map[string]string) (volume.VolumeDriver, error) {
		return tester.m, nil
	})

	var err error

	// Register mock driver
	err = vol_drivers.Register(mockDriverName, nil)
	assert.Nil(t, err)
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

func (s *testServer) Stop() {
	// Remove from registry
	vol_drivers.Remove(mockDriverName)

	// Check mocks
	s.mc.Finish()
}

func TestClientCredsValidateAndDelete(t *testing.T) {
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
	defer ts.Close()
	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver := newTestServer(t)
	defer testVolDriver.Stop()

	testVolDriver.MockDriver().EXPECT().CredsDelete("gooduuid").Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CredsDelete("baduuid").Return(fmt.Errorf("Invalid UUID")).Times(1)

	testVolDriver.MockDriver().EXPECT().CredsValidate("gooduuid").Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CredsValidate("baduuid").Return(fmt.Errorf("Invalid UUID")).Times(1)

	// Delete creds
	err = client.VolumeDriver(cl).CredsDelete("gooduuid")
	require.NoError(t, err)
	err = client.VolumeDriver(cl).CredsDelete("baduuid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid UUID")
	err = client.VolumeDriver(cl).CredsDelete("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "404")
	//Validate creds
	err = client.VolumeDriver(cl).CredsValidate("gooduuid")
	require.NoError(t, err)
	err = client.VolumeDriver(cl).CredsValidate("baduuid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid UUID")
	err = client.VolumeDriver(cl).CredsValidate("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "404")

}
