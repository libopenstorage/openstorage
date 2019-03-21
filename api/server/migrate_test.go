package server

import (
	"context"
	"github.com/libopenstorage/openstorage/api"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrateStart(t *testing.T) {
	var err error
	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, "fake", version, token, "", "fake")
	assert.NoError(t, err)

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 3,
			Size:    size,
			Format:  api.FSType_FS_TYPE_EXT4,
			Shared:  true,
		},
	}

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(cl)
	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	goodRequest := &api.CloudMigrateStartRequest{
		TaskId:    "123456",
		Operation: api.CloudMigrate_MigrateVolume,
		ClusterId: "clusterID",
		TargetId:  id,
	}

	// Start Migrate
	resp, err := volumeclient.VolumeDriver(cl).CloudMigrateStart(goodRequest)
	assert.Nil(t, err)
	assert.NotNil(t, resp.TaskId)
	assert.Equal(t, goodRequest.TaskId, resp.TaskId)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestMigrateCancel(t *testing.T) {
	var err error
	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, "fake", version, token, "", "fake")
	assert.NoError(t, err)

	goodRequest := &api.CloudMigrateCancelRequest{
		TaskId: "goodTaskID",
	}

	// Cancel Migrate
	err = volumeclient.VolumeDriver(cl).CloudMigrateCancel(goodRequest)
	assert.Nil(t, err)
}

func TestMigrateStatus(t *testing.T) {
	var err error
	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, "fake", version, token, "", "fake")
	assert.NoError(t, err)

	// Get Migrate status
	resp, err := volumeclient.VolumeDriver(cl).CloudMigrateStatus(&api.CloudMigrateStatusRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp.Info))
}
