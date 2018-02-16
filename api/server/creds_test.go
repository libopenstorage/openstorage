package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/volume"
)

func TestClientCredCreate(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()
	var uuid string
	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)
	// S3 cloud provider
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "s3",
			api.OptCredRegion:    "east",
			api.OptCredEndpoint:  "s3.url.com",
			api.OptCredAccessKey: "s3accesskey",
			api.OptCredSecretKey: "s3secretekey",
		}).
		Return("gooduuid", nil).
		Times(1)
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "s3",
			api.OptCredRegion:    "east",
			api.OptCredEndpoint:  "s3.url.com",
			api.OptCredAccessKey: "",
			api.OptCredSecretKey: "",
		}).
		Return("", fmt.Errorf("Missing s3 access/secrete keys")).
		Times(1)

	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "azure",
			api.OptCredAzureAccountName: "azuretest",
			api.OptCredAzureAccountKey:  "azureaccountkey",
		}).
		Return("gooduuid", nil).
		Times(1)
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "azure",
			api.OptCredAzureAccountName: "",
			api.OptCredAzureAccountKey:  "",
		}).
		Return("", fmt.Errorf("Missing azure account name/keys")).
		Times(1)

	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "google",
			api.OptCredGoogleProjectID: "googletestproject",
			api.OptCredGoogleJsonKey:   "googlejsonkey",
		}).
		Return("gooduuid", nil).
		Times(1)
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "google",
			api.OptCredGoogleProjectID: "",
			api.OptCredGoogleJsonKey:   "",
		}).
		Return("", fmt.Errorf("Missing google project/json key")).
		Times(1)

	// Invoke CredsCreate for S3
	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "s3",
		api.OptCredRegion:    "east",
		api.OptCredEndpoint:  "s3.url.com",
		api.OptCredAccessKey: "s3accesskey",
		api.OptCredSecretKey: "s3secretekey",
	})
	require.NoError(t, err)
	require.Equal(t, uuid, "gooduuid")

	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "s3",
		api.OptCredRegion:    "east",
		api.OptCredEndpoint:  "s3.url.com",
		api.OptCredAccessKey: "",
		api.OptCredSecretKey: "",
	})
	require.Error(t, err)
	require.Equal(t, uuid, "")
	require.Contains(t, err.Error(), "Missing")

	// Azure cloud provider
	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "azure",
		api.OptCredAzureAccountName: "azuretest",
		api.OptCredAzureAccountKey:  "azureaccountkey",
	})
	require.NoError(t, err)
	require.Equal(t, uuid, "gooduuid")

	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "azure",
		api.OptCredAzureAccountName: "",
		api.OptCredAzureAccountKey:  "",
	})
	require.Error(t, err)
	require.Equal(t, uuid, "")
	require.Contains(t, err.Error(), "Missing")

	//Google

	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "google",
		api.OptCredGoogleProjectID: "googletestproject",
		api.OptCredGoogleJsonKey:   "googlejsonkey",
	})
	require.NoError(t, err)
	require.Equal(t, uuid, "gooduuid")
	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "google",
		api.OptCredGoogleProjectID: "",
		api.OptCredGoogleJsonKey:   "",
	})
	require.Error(t, err)
	require.Equal(t, uuid, "")
	require.Contains(t, err.Error(), "Missing")

}

func TestClientCredsValidateAndDelete(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

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

func TestClientCredsList(t *testing.T) {
	ts, testVolDriver := testRestServer(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)
	enumerateData := make(map[string]interface{}, 0)
	testVolDriver.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil).
		Times(1)
	_, err = client.VolumeDriver(cl).CredsEnumerate()
	require.NoError(t, err)
}
