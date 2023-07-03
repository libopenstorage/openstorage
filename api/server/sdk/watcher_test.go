package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestSdkVolumeWatch(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()

	// Setup client
	c := api.NewOpenStorageWatchClient(s.Conn())

	req := api.SdkWatchRequest{
		EventType: &api.SdkWatchRequest_VolumeEvent{
			VolumeEvent: &api.SdkVolumeWatchRequest{
				Labels: make(map[string]string),
			},
		},
	}
	id := "myid"
	size := uint64(1234)
	name := id
	vol := &api.Volume{
		Id:     id,
		Status: api.VolumeStatus_VOLUME_STATUS_UP,
		State:  api.VolumeState_VOLUME_STATE_DELETED,
		Locator: &api.VolumeLocator{
			Name: name,
		},
		Spec: &api.VolumeSpec{
			Size: size,
		},
	}

	// Set up expectations for the mock client
	s.MockDriver().
		EXPECT().
		Enumerate(&api.VolumeLocator{}, nil).
		Return([]*api.Volume{vol}, nil).
		Times(1)

	client, err := c.Watch(context.Background(), &req)
	assert.NoError(t, err)

	r, err := client.Recv()
	if err != nil {
		fmt.Println(err)
	}
	v := r.GetVolumeEvent()
	assert.Equal(t, v.GetName(), "myid")
}
