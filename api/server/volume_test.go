package server

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	version = "v1"
)

func TestJunk(t *testing.T) {

	var err error
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	require.NoError(t, err)

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
	testVolDriver.MockDriver().
		EXPECT().
		Create(req.GetLocator(), req.GetSource(), req.GetSpec()).
		Return(id, nil)

	// create a volume client
	driverclient := volumeclient.VolumeDriver(cl)
	res, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())

	assert.Nil(t, err)
	assert.Equal(t, id, res)

	fmt.Println("Yay -0---- ", ts.URL, " --- ", res)
}
