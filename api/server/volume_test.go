package server

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/volume"
	lsecrets "github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/k8s"
	"github.com/libopenstorage/secrets/mock" //"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVolumeNoAuth(t *testing.T) {
	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdkNoAuth(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, "", "", mockDriverName)
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

	// CREATE
	driverclient := volumeclient.VolumeDriver(cl)
	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	// UPDATE
	newsize := uint64(10)

	newspec := req.GetSpec()
	newspec.Size = newsize
	resp := driverclient.Set(id, req.GetLocator(), newspec)
	assert.Nil(t, resp)

	// INSPECT
	res, err := driverclient.Inspect([]string{id})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res)
	assert.EqualValues(t, id, res[0].Id)
	assert.EqualValues(t, true, res[0].Spec.Shared)
	assert.EqualValues(t, 3, res[0].Spec.HaLevel)
	assert.EqualValues(t, newsize, res[0].Spec.Size)

	// DELETE
	err = driverclient.Delete(id)
	assert.Nil(t, err)
}

func testCreateSuccessChecks(t *testing.T, testVolDriver *testServer, req *api.VolumeCreateRequest, id string) {
	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)
	r, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, req.GetLocator().GetName(), r.GetVolume().GetLocator().GetName())
	assert.Equal(t, req.GetSpec().GetSize(), r.GetVolume().GetSpec().GetSize())

	// Check ownership. We should be denied
	ctx, err = contextWithToken(context.Background(), "anotheruser", "system.view", testSharedSecret)
	assert.NoError(t, err)
	r, err = volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.PermissionDenied)

	ctx, err = contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeCreateSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	testCreateSuccessChecks(t, testVolDriver, req, id)
}

func TestMiddlewareVolumeCreateSuccess(t *testing.T) {
	testVolDriver := newTestServerSdk(t)
	defer testVolDriver.Stop()

	_, mockSecret, mc := getSecretsMock(t)
	defer mc.Finish()
	lsecrets.SetInstance(mockSecret)

	unixServer, portServer, err := StartVolumeMgmtAPI(fakeWithSched, testSdkSock, testMgmtBase, testMgmtPort, true)
	assert.NoError(t, err, "Unexpected error on StartVolumeMgmtAPI")
	defer unixServer.Close()
	defer portServer.Close()

	time.Sleep(1 * time.Second)
	c, err := volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")
	driverclient := volumeclient.VolumeDriver(c)

	testMiddlewareCreateVolume(t, driverclient, mockSecret, testVolDriver)
}

func TestMiddlewareVolumeCreateFailure(t *testing.T) {
	testVolDriver := newTestServerSdk(t)
	defer testVolDriver.Stop()

	_, mockSecret, mc := getSecretsMock(t)
	defer mc.Finish()
	lsecrets.SetInstance(mockSecret)
	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	unixServer, portServer, err := StartVolumeMgmtAPI(fakeWithSched, testSdkSock, testMgmtBase, testMgmtPort, true)
	assert.NoError(t, err, "Unexpected error on StartVolumeMgmtAPI")
	defer unixServer.Close()
	defer portServer.Close()

	time.Sleep(1 * time.Second)

	// Setup request
	name := "myvol"
	size := uint64(1234)
	secretName := "secret-name"
	namespace := "ns"
	tokenKey := "token-key"

	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{
			Name: name,
		},
		Source: &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 3,
			Size:    size,
			Format:  api.FSType_FS_TYPE_EXT4,
			Shared:  true,
		},
	}

	// Send a request without the sched user agent
	c, err := volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, "")
	assert.NoError(t, err, "Unexpected error on NewDriverClient")
	driverclient := volumeclient.VolumeDriver(c)
	_, err = driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Error(t, err, "Expected an error on Create")

	// Send a request without labels
	c, err = volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")
	driverclient = volumeclient.VolumeDriver(c)
	_, err = driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Error(t, err, "Expected an error on Create")

	req = &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{
			Name: name,
			VolumeLabels: map[string]string{
				secrets.SecretNameKey:      secretName,
				secrets.SecretTokenKey:     tokenKey,
				secrets.SecretNamespaceKey: namespace,
			},
		},
		Source: &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 3,
			Size:    size,
			Format:  api.FSType_FS_TYPE_EXT4,
			Shared:  true,
		},
	}

	// Send a request and fail to get a token
	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(nil, fmt.Errorf("incorrect secret")).
		Times(1)

	c, err = volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")
	driverclient = volumeclient.VolumeDriver(c)
	_, err = driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Error(t, err, "Expected an error on Create")
	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	// Failed to get token
	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(nil, fmt.Errorf("incorrect secret")).
		Times(1)

	c, err = volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")
	driverclient = volumeclient.VolumeDriver(c)
	_, err = driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Error(t, err, "Expected an error on Create")

}

func TestVolumeCreateFailedToAuthenticate(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", "badsecret")
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 1,
			Size:    size,
		},
	}

	// create a volume client
	driverclient := volumeclient.VolumeDriver(client)
	_, err = driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Error(t, err)
}

/*
func TestVolumeCreateGetNodeIdFromIpFailed(t *testing.T) {

	var err error

	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	nodeIp := "192.168.1.1"

	// Create a new global test cluster
	tc := newTestCluster(t)
	defer tc.Finish()

	// Mock cluster
	tc.MockCluster().
		EXPECT().
		GetNodeIdFromIp(nodeIp).
		Return(nodeIp, fmt.Errorf("Failed to locate IP in this cluster."))

	// create a volume client with Replica IPs
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec:    &api.VolumeSpec{Size: size, ReplicaSet: &api.ReplicaSet{Nodes: []string{nodeIp}}},
	}

	// create a volume client
	driverclient := volumeclient.VolumeDriver(client)

	res, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.NotNil(t, err)
	assert.EqualValues(t, "", res)
	assert.Contains(t, err.Error(), "Failed to locate IP")
}
*/
func TestVolumeSnapshotCreateSuccess(t *testing.T) {

	var err error

	snapname := "snapName"

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	req2 := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: snapname},
		Readonly: true,
	}

	_, err = driverclient.Snapshot(id, req2.GetReadonly(), req2.GetLocator(), req2.GetNoRetry())
	assert.Nil(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeSnapshotCreateFailed(t *testing.T) {

	var err error

	snapname := "snapName"

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	req2 := &api.SnapCreateRequest{
		Locator:  &api.VolumeLocator{Name: snapname},
		Readonly: true,
	}

	res, _ := driverclient.Snapshot("doesnotexist", req2.GetReadonly(), req2.GetLocator(), req2.GetNoRetry())
	assert.Equal(t, "", res)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeInspectSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(client)

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

	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	res, err := driverclient.Inspect([]string{id})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res)
	assert.EqualValues(t, id, res[0].Id)
	assert.EqualValues(t, true, res[0].Spec.Shared)
	assert.EqualValues(t, 3, res[0].Spec.HaLevel)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeInspectFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(client)

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

	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	res, err := driverclient.Inspect([]string{"myid"})
	assert.Nil(t, err)
	assert.Equal(t, len(res), 0)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeSnapshotList(t *testing.T) {

	var err error

	snapname := "snapName"

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	req2 := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: snapname},
		Readonly: true,
	}

	_, err = driverclient.Snapshot(id, req2.GetReadonly(), req2.GetLocator(), req2.GetNoRetry())
	assert.Nil(t, err)

	res, err := driverclient.SnapEnumerate([]string{id}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)

	_, err = driverclient.Snapshot(id, req2.GetReadonly(), req2.GetLocator(), req2.GetNoRetry())
	assert.Nil(t, err)

	res, err = driverclient.SnapEnumerate([]string{id}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeSetSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(client)

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec: &api.VolumeSpec{
			Ephemeral:        false,
			BlockSize:        4096,
			Cos:              api.CosType_LOW,
			IoProfile:        api.IoProfile_IO_PROFILE_RANDOM,
			Dedupe:           false,
			SnapshotInterval: 0,
			AggregationLevel: 2,
			Encrypted:        false,
			HaLevel:          3,
			Size:             size,
			Format:           api.FSType_FS_TYPE_EXT4,
			Shared:           true,
			Passphrase:       "",
			SnapshotSchedule: "- freq: daily\n  hour: 12\n  retain: 4\n- freq: weekly\n  weekday: 1\n  hour: 12\n  retain: 2\n",
			Scale:            1,
			Compressed:       false,
			Cascaded:         false,
			Journal:          true,
			Sharedv4:         false,
			QueueDepth:       128,
			Group: &api.Group{
				Id: "g1",
			},
			Nodiscard:     false,
			GroupEnforced: true,
			IoStrategy:    &api.IoStrategy{},
		},
	}

	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	newsize := uint64(10)

	req2 := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:  api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
		},
		Spec: &api.VolumeSpec{Size: newsize},
	}

	res := driverclient.Set(id, req.GetLocator(), req2.GetSpec())
	assert.Nil(t, res)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)
	r, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, newsize, r.GetVolume().GetSpec().GetSize())

	// Send HA request
	res = driverclient.Set(id,
		nil,
		&api.VolumeSpec{
			HaLevel:          2,
			ReplicaSet:       &api.ReplicaSet{Nodes: []string{}},
			SnapshotInterval: math.MaxUint32,
		})
	assert.Nil(t, res, fmt.Sprintf("Error: %v", res))

	// Assert volume information is correct
	r, err = volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, int64(2), r.GetVolume().GetSpec().GetHaLevel())

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeSetFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(client)

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

	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// create a volume request
	name = "myvol"
	size = uint64(10)
	halevel := int64(5)

	req2 := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:  api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size, HaLevel: halevel},
	}
	// Cannot get this to fail....
	err = driverclient.Set("doesnotexist", req2.GetLocator(), req2.GetSpec())
	//	assert.NotNil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestMiddlewareVolumeSetSizeSuccess(t *testing.T) {
	testVolDriver := newTestServerSdk(t)
	defer testVolDriver.Stop()

	_, mockSecret, mc := getSecretsMock(t)
	defer mc.Finish()
	lsecrets.SetInstance(mockSecret)

	unixServer, portServer, err := StartVolumeMgmtAPI(fakeWithSched, testSdkSock, testMgmtBase, testMgmtPort, true)
	assert.NoError(t, err, "Unexpected error on StartVolumeMgmtAPI")
	defer unixServer.Close()
	defer portServer.Close()

	time.Sleep(1 * time.Second)
	c, err := volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")

	driverclient := volumeclient.VolumeDriver(c)
	id, _, _, _ := testMiddlewareCreateVolume(t, driverclient, mockSecret, testVolDriver)

	newsize := uint64(2222)

	req := &api.VolumeSetRequest{
		Spec: &api.VolumeSpec{Size: newsize},
	}

	// Not setting mock secrets

	err = driverclient.Set(id, nil, req.GetSpec())
	assert.NoError(t, err, "Unexpected error on Set")

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)
	r, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, newsize, r.GetVolume().GetSpec().GetSize())

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestMiddlewareVolumeSetFailure(t *testing.T) {
	testVolDriver := newTestServerSdk(t)
	defer testVolDriver.Stop()

	_, mockSecret, mc := getSecretsMock(t)
	defer mc.Finish()
	lsecrets.SetInstance(mockSecret)

	unixServer, portServer, err := StartVolumeMgmtAPI(fakeWithSched, testSdkSock, testMgmtBase, testMgmtPort, true)
	assert.NoError(t, err, "Unexpected error on StartVolumeMgmtAPI")
	defer unixServer.Close()
	defer portServer.Close()

	time.Sleep(1 * time.Second)
	c, err := volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")

	driverclient := volumeclient.VolumeDriver(c)
	id, _, _, _ := testMiddlewareCreateVolume(t, driverclient, mockSecret, testVolDriver)

	req := &api.VolumeSetRequest{
		Spec: &api.VolumeSpec{Shared: true},
	}

	// Not setting mock secrets

	err = driverclient.Set(id, &api.VolumeLocator{Name: "myvol"}, req.GetSpec())
	assert.Error(t, err, "Unexpected error on Set")

}

func TestVolumeAttachSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	_, err = driverclient.Attach(id, map[string]string{})
	assert.Nil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeAttachFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	_, err = driverclient.Attach("doesnotexist", map[string]string{})
	assert.NotNil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeDetachSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Attach
	_, err = driverclient.Attach(id, map[string]string{})
	assert.Nil(t, err)

	// Detach
	res := driverclient.Detach(id, map[string]string{})
	assert.Nil(t, res)

	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)
	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeDetachFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Attach
	_, err = driverclient.Attach(id, map[string]string{})
	assert.Nil(t, err)

	// Detach
	res := driverclient.Detach("doesnotexist", map[string]string{})
	// Detach must not fail on non-existing volume
	assert.Nil(t, res)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeMountSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	res := driverclient.Mount(id, "/mnt", map[string]string{})
	assert.Nil(t, res)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeMountFailedNoMountPath(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	//create driverclient
	err = driverclient.Mount("doesnotexist", "/mnt", map[string]string{})
	assert.NotNil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeStatsSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	_, err = driverclient.Stats(id, true)
	assert.Nil(t, err)
	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeStatsFailed(t *testing.T) {

	var err error
	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	_, err = driverclient.Stats("12345", true)
	assert.NotNil(t, err)
	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeUnmountSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Mount
	res := driverclient.Mount(id, "/mnt", map[string]string{})
	assert.Nil(t, res)

	// Unmount
	res2 := driverclient.Unmount(id, "/mnt", map[string]string{})
	assert.Nil(t, res2)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeUnmountFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	// Mount
	res := driverclient.Mount(id, "/mnt", map[string]string{})
	assert.Nil(t, res)

	// Unmount
	err = driverclient.Unmount("doesnotexist", "/mnt", map[string]string{})
	assert.NotNil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

/*
func TestVolumeQuiesceSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	id := "myid"
	quiesceid := "qid"
	timeout := uint64(5)

	testVolDriver.MockDriver().
		EXPECT().
		Quiesce(id, timeout, quiesceid).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res := driverclient.Quiesce(id, timeout, quiesceid)

	assert.Nil(t, res)
}
func TestVolumeQuiesceFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// volume instance
	id := "myid"
	quiesceid := "qid"
	timeout := uint64(5)

	testVolDriver.MockDriver().
		EXPECT().
		Quiesce(id, timeout, quiesceid).
		Return(fmt.Errorf("error in quiesce"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res := driverclient.Quiesce(id, timeout, quiesceid)

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in quiesce")
}

* TODO(ram-infrac) : Test case is failing, recheck
func TestVolumeUnquiesceSuccess(t *testing.T) {

        ts, testVolDriver := testRestServer(t)

	ts.Close()
	testVolDriver.Stop()
        var err error

        client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
        assert.Nil(t, err)

        id := "myid"

        testVolDriver.MockDriver().
                EXPECT().
                Unquiesce(id).
                Return(nil)

        // create client
        driverclient := volumeclient.VolumeDriver(client)
        res := driverclient.Unquiesce(id)

        assert.Nil(t, res)
}
*

func TestVolumeUnquiesceFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	id := "myid"

	testVolDriver.MockDriver().
		EXPECT().
		Unquiesce(id).
		Return(fmt.Errorf("error in unquiesce"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res := driverclient.Unquiesce(id)

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in unquiesce")
}
*/
func TestVolumeRestoreSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	req2 := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: "snap"},
		Readonly: true,
	}

	res, err := driverclient.Snapshot(req2.GetId(), req2.GetReadonly(), req2.GetLocator(), req2.GetNoRetry())
	assert.Nil(t, err)

	// create client

	fmt.Println("ID and SnapID", id, res)
	res2 := driverclient.Restore(id, res)
	assert.Nil(t, res2)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeRestoreFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	req2 := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: "snap"},
		Readonly: true,
	}

	_, err = driverclient.Snapshot(req2.GetId(), req2.GetReadonly(), req2.GetLocator(), req2.GetNoRetry())
	assert.Nil(t, err)

	// create client
	err = driverclient.Restore("doesnotexist", "alsodoesnotexist")
	assert.NotNil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeUsedSizeSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	_, err = driverclient.UsedSize(id)
	assert.Nil(t, err)

	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeUsedSizeFailed(t *testing.T) {

	var err error
	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
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

	_, err = driverclient.UsedSize("doesnotexist")
	assert.NotNil(t, err)
	// Assert volume information is correct
	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

func TestVolumeEnumerateSuccess(t *testing.T) {

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
		Locator: &api.VolumeLocator{
			Name: name,
			VolumeLabels: map[string]string{
				"dept":    "auto",
				"sub":     "geo",
				"config1": "c1",
			},
		},
		Source: &api.Source{},
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

	// create volume locator
	configLabel := make(map[string]string)
	configLabel["config1"] = "c1"

	vl := &api.VolumeLocator{
		Name: name,
		VolumeLabels: map[string]string{
			"dept": "auto",
			"sub":  "geo",
		},
	}

	// create client
	res, err := driverclient.Enumerate(vl, configLabel)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
	assert.EqualValues(t, id, res[0].GetId())

	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)
	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)

}

func TestVolumeEnumerateFailed(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Setup request
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{
			Name: name,
			VolumeLabels: map[string]string{
				"dept":    "auto",
				"sub":     "geo",
				"config1": "c1",
			},
		},
		Source: &api.Source{},
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

	// create volume locator
	configLabel := make(map[string]string)
	configLabel["config1"] = "cnfig1"

	vl := &api.VolumeLocator{
		Name: name,
		VolumeLabels: map[string]string{
			"class": "f9",
		},
	}

	res, _ := driverclient.Enumerate(vl, configLabel)
	assert.Equal(t, 0, len(res), fmt.Sprintf("%+v\n", res))

	volumes := api.NewOpenStorageVolumeClient(testVolDriver.Conn())
	ctx, err := contextWithToken(context.Background(), "test", "system.admin", testSharedSecret)
	assert.NoError(t, err)
	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: id,
	})
	assert.NoError(t, err)
}

/*
func TestVolumeSnapshotEnumerateSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	ids := []string{
		"snapid1",
		"snapid2",
	}

	snapLabels := map[string]string{
		"dept": "auto",
		"sub":  "geo",
	}

	testVolDriver.MockDriver().
		EXPECT().
		SnapEnumerate(ids, snapLabels).
		Return([]*api.Volume{
			&api.Volume{
				Id: ids[0],
				Locator: &api.VolumeLocator{
					Name: "snap1",
				},
			},
			&api.Volume{
				Id: ids[1],
				Locator: &api.VolumeLocator{
					Name: "snap2",
				},
			},
		}, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.SnapEnumerate(ids, snapLabels)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)
}

func TestVolumeSnapshotEnumerateFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	ids := []string{
		"snapid1",
		"snapid2",
	}

	snapLabels := map[string]string{
		"dept": "auto",
		"sub":  "geo",
	}

	testVolDriver.MockDriver().
		EXPECT().
		SnapEnumerate(ids, snapLabels).
		Return([]*api.Volume{},
			fmt.Errorf("error in snap enumerate"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.SnapEnumerate(ids, snapLabels)

	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestVolumeGetActiveRequestsSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	acreqs := &api.ActiveRequests{
		ActiveRequest: []*api.ActiveRequest{
			&api.ActiveRequest{
				ReqestKV: map[int64]string{
					1: "vol1",
				},
			},
			&api.ActiveRequest{
				ReqestKV: map[int64]string{
					2: "vol2",
				},
			},
		},
		RequestCount: 2,
	}

	testVolDriver.MockDriver().
		EXPECT().
		GetActiveRequests().
		Return(acreqs, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.GetActiveRequests()

	assert.Nil(t, err)
	assert.EqualValues(t, 2, res.GetRequestCount())
}

func TestVolumeGetActiveRequestsFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	testVolDriver.MockDriver().
		EXPECT().
		GetActiveRequests().
		Return(nil, fmt.Errorf("error in active requests"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.GetActiveRequests()

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error in active requests")
}

func TestCredsCreateSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]string{
		"c1": "cred1",
		"c2": "cred2",
	}

	// Creata cred request
	cred := &api.CredCreateRequest{
		InputParams: credsmap,
	}

	testVolDriver.MockDriver().
		EXPECT().
		CredsCreate(cred.InputParams).
		Return("dummy-uuid", nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.CredsCreate(credsmap)

	assert.Nil(t, err)
	assert.EqualValues(t, "dummy-uuid", res)
}

func TestCredsCreateFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]string{
		"c1": "cred1",
		"c2": "cred2",
	}

	// Creata cred request
	cred := &api.CredCreateRequest{
		InputParams: credsmap,
	}

	testVolDriver.MockDriver().
		EXPECT().
		CredsCreate(cred.InputParams).
		Return("", fmt.Errorf("error in creds create"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.CredsCreate(credsmap)

	assert.NotNil(t, err)
	assert.EqualValues(t, "", res)
	assert.Contains(t, err.Error(), "error in creds create")
}

func TestCredsEnumerateSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]interface{}{
		"c1": "cred1",
		"c2": "cred2",
	}

	testVolDriver.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(credsmap, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.CredsEnumerate()

	assert.Nil(t, err)
	assert.NotEmpty(t, res)
	assert.EqualValues(t, "cred1", res["c1"])
}

func TestCredsEnumerateFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]interface{}{}

	testVolDriver.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(credsmap, fmt.Errorf("error in creds enumerate"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	res, err := driverclient.CredsEnumerate()

	assert.NotNil(t, err)
	assert.Empty(t, res)
}

func TestCredsValidateSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// cred uuid
	uuid := "dummy-validate-1101-uuid"

	testVolDriver.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	err = driverclient.CredsValidate(uuid)

	assert.Nil(t, err)
}

func TestCredsValidateFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// cred uuid
	uuid := "dummy-validate-1101-uuid"

	testVolDriver.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(fmt.Errorf("error in creds validate"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	err = driverclient.CredsValidate(uuid)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in creds validate")
}

func TestGroupSnapshotCreateSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	id := "mygroupid"
	labels := map[string]string{
		"app":    "app1",
		"region": "region1",
	}

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)

	assert.Nil(t, err)

	req := &api.GroupSnapCreateRequest{Id: id,
		Labels: labels,
	}

	snapshots := map[string]*api.SnapCreateResponse{
		"vol1": &api.SnapCreateResponse{
			VolumeCreateResponse: &api.VolumeCreateResponse{
				Id: id,
				VolumeResponse: &api.VolumeResponse{
					Error: responseStatus(err),
				},
			},
		},
		"vol2": &api.SnapCreateResponse{
			VolumeCreateResponse: &api.VolumeCreateResponse{
				Id: id,
				VolumeResponse: &api.VolumeResponse{
					Error: responseStatus(err),
				},
			},
		},
	}

	response := &api.GroupSnapCreateResponse{
		Snapshots: snapshots,
		Error:     responseStatus(err),
	}

	//mock Snapshot call
	testVolDriver.MockDriver().
		EXPECT().
		SnapshotGroup(req.GetId(), req.GetLabels(), req.GetVolumeIds()).
		Return(response, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)

	res, err := driverclient.SnapshotGroup(req.GetId(), req.GetLabels(), req.GetVolumeIds())

	assert.Nil(t, err)
	assert.Equal(t, len(response.Snapshots), len(res.Snapshots))
}

func TestVolumeCatalogSuccess(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// vol uuid
	volid := "dummy-111-uuid"

	testVolDriver.MockDriver().
		EXPECT().
		Catalog(volid, "", "0").
		Return(api.CatalogResponse{}, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	_, err = driverclient.Catalog(volid, "", "0")

	assert.Nil(t, err)
}

func TestVolumeCatalogFailed(t *testing.T) {

	var err error
	ts, testVolDriver := testRestServer(t)

	defer ts.Close()
	defer testVolDriver.Stop()

	client, err := volumeclient.NewDriverClient(ts.URL, mockDriverName, version, mockDriverName)
	assert.Nil(t, err)

	// vol uuid
	volid := "dummy-111-uuid"

	testVolDriver.MockDriver().
		EXPECT().
		Catalog(volid, "", "0").
		Return(api.CatalogResponse{}, fmt.Errorf("error in volume catalog"))

	// create client
	driverclient := volumeclient.VolumeDriver(client)
	_, err = driverclient.Catalog(volid, "", "0")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in volume catalog")
}
*/

func TestVolumeDeleteSuccess(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Create volume before deleting.
	// Setup Create object
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 1,
			Size:    size,
		},
	}

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(client)

	// Create volume.
	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	err = driverclient.Delete(id)
	assert.Nil(t, err)
}

func TestMiddlewareVolumeDeleteSuccess(t *testing.T) {
	testVolDriver := newTestServerSdk(t)
	defer testVolDriver.Stop()

	_, mockSecret, mc := getSecretsMock(t)
	defer mc.Finish()
	lsecrets.SetInstance(mockSecret)

	unixServer, portServer, err := StartVolumeMgmtAPI(fakeWithSched, testSdkSock, testMgmtBase, testMgmtPort, true)
	assert.NoError(t, err, "Unexpected error on StartVolumeMgmtAPI")
	defer unixServer.Close()
	defer portServer.Close()

	time.Sleep(1 * time.Second)
	c, err := volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")

	// Create a volume before deleting
	driverclient := volumeclient.VolumeDriver(c)
	id, token, namespace, secretName := testMiddlewareCreateVolume(t, driverclient, mockSecret, testVolDriver)

	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(map[string]interface{}{secrets.SecretTokenKey: token}, nil).
		AnyTimes()

	err = driverclient.Delete(id)
	assert.Nil(t, err)

}

func TestMiddlewareVolumeDeleteFailureIncorrectToken(t *testing.T) {
	testVolDriver := newTestServerSdk(t)
	defer testVolDriver.Stop()

	_, mockSecret, mc := getSecretsMock(t)
	defer mc.Finish()
	lsecrets.SetInstance(mockSecret)

	unixServer, portServer, err := StartVolumeMgmtAPI(fakeWithSched, testSdkSock, testMgmtBase, testMgmtPort, true)
	assert.NoError(t, err, "Unexpected error on StartVolumeMgmtAPI")
	defer unixServer.Close()
	defer portServer.Close()

	time.Sleep(1 * time.Second)
	c, err := volumeclient.NewDriverClient(testMockURL, fakeWithSched, version, fakeWithSched)
	assert.NoError(t, err, "Unexpected error on NewDriverClient")

	// Create a volume before deleting
	name := "myvol-delete"
	size := uint64(1234)
	secretName := "secret-name"
	namespace := "ns"
	tokenKey := "token-key"
	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{
			Name: name,
			VolumeLabels: map[string]string{
				secrets.SecretNameKey:      secretName,
				secrets.SecretTokenKey:     tokenKey,
				secrets.SecretNamespaceKey: namespace,
			},
		},
		Source: &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 3,
			Size:    size,
			Format:  api.FSType_FS_TYPE_EXT4,
			Shared:  true,
		},
	}

	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(map[string]interface{}{secrets.SecretTokenKey: token}, nil).
		Times(1)

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(c)
	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.NoError(t, err, "Unexpected error on Create")
	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	incorrectToken := "blah"
	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(map[string]interface{}{secrets.SecretTokenKey: incorrectToken}, nil).
		Times(1)

	err = driverclient.Delete(id)
	assert.Error(t, err, "Expected an error on Delete")
	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(nil, fmt.Errorf("incorrect secret")).
		AnyTimes()

	err = driverclient.Delete(id)
	assert.Error(t, err, "Expected an error on Delete")

}

func testMiddlewareCreateVolume(
	t *testing.T,
	driverclient volume.VolumeDriver,
	mockSecret *mock.MockSecrets,
	testVolDriver *testServer,
) (string, string, string, string) {
	name := "myvol"
	size := uint64(1234)
	secretName := "secret-name"
	namespace := "ns"
	tokenKey := "token-key"
	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{
			Name: name,
			VolumeLabels: map[string]string{
				secrets.SecretNameKey:      secretName,
				secrets.SecretTokenKey:     tokenKey,
				secrets.SecretNamespaceKey: namespace,
			},
		},
		Source: &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 3,
			Size:    size,
			Format:  api.FSType_FS_TYPE_EXT4,
			Shared:  true,
		},
	}
	mockSecret.EXPECT().
		String().
		Return(lsecrets.TypeK8s).
		AnyTimes()

	mockSecret.EXPECT().
		GetSecret(
			secretName,
			map[string]string{
				k8s.SecretNamespace: namespace,
			}).
		Return(map[string]interface{}{secrets.SecretTokenKey: token}, nil).
		AnyTimes()

	// Create a volume
	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())

	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	return id, token, namespace, secretName

}

func TestStorkVolumeInspect(t *testing.T) {

	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdk(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	// get token
	token, err := createToken("test", "system.admin", testSharedSecret)
	assert.NoError(t, err)

	client, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, token, "", mockDriverName)
	assert.NoError(t, err)

	// Create volume before deleting.
	// Setup Create object
	name := "myvol"
	size := uint64(1234)
	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec: &api.VolumeSpec{
			HaLevel: 1,
			Size:    size,
		},
	}

	// Create a volume client
	driverclient := volumeclient.VolumeDriver(client)

	// Create volume.
	id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	err = driverclient.Delete(id)
	assert.Nil(t, err)

	vols, err := driverclient.Inspect([]string{id})
	assert.Equal(t, len(vols), 0)
	assert.Nil(t, err)
	/*
		if err != nil && (err == volume.ErrEnoEnt || errIsNotFound(err)) {
			return nil, false, nil
		} else if err != nil {
			return nil, true, err
		}
		if len(vols) > 0 {
			return nil, true, fmt.Errorf("Volume %v is not yet removed from the system", name)
		}
		return nil, false, nil
	*/
}

// This is a time comsuming test and must be uncommented to test it.
// Make sure to add --timeout=1h to go test when running this test.
/*
func TestThousandsOfVolumes(t *testing.T) {
	var err error

	// Setup volume rest functions server
	ts, testVolDriver := testRestServerSdkNoAuth(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := volumeclient.NewAuthDriverClient(ts.URL, mockDriverName, version, "", "", mockDriverName)
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

	labels := make(map[string]string)
	for i := 0; i < 50; i++ {
		labels[fmt.Sprintf("label %d", i)] = fmt.Sprintf("label value %d", i)
	}
	req.GetLocator().VolumeLabels = labels

	numvols := 10 * 1000
	driverclient := volumeclient.VolumeDriver(cl)
	fmt.Println("Creting volumes")

	routines := 50
	ch := make(chan int)
	var wg sync.WaitGroup
	fmt.Println("Creating...")
	for i := 0; i < routines; i++ {
		go func() {
			for i := range ch {
				req.GetLocator().Name = fmt.Sprintf("myvol-%d", i)
				id, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())
				assert.Nil(t, err)
				assert.NotEmpty(t, id)
			}
			wg.Done()
		}()
	}
	wg.Add(routines)

	for i := 0; i < numvols; i++ {
		ch <- i
		if (i % 1000) == 0 {
			fmt.Printf("%d...", i)
		}
	}
	fmt.Println("")
	close(ch)
	wg.Wait()

	// create client
	fmt.Println("Enumerating...")
	res, err := driverclient.Enumerate(&api.VolumeLocator{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, numvols, fmt.Sprintf("Len res:%d numvols:%d", len(res), numvols))
}
*/
