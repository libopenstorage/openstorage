package volume

import (
	"crypto/tls"
	"encoding/json"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientTLS(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var vol *api.Volume

		json.NewEncoder(w).Encode(vol)
	}))

	defer ts.Close()

	clnt, err := NewDriverClient(ts.URL, "pxd", "", "")
	require.NoError(t, err)

	clnt.SetTLS(&tls.Config{InsecureSkipVerify: true})

	_, err = VolumeDriver(clnt).Inspect([]string{"12345"})

	require.NoError(t, err)
}

func TestClientCredCreate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request *api.CredCreateRequest
		var response *api.CredCreateResponse
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Failed decode input parameters", http.StatusBadRequest)
			return
		}
		if len(request.InputParams) == 0 {
			http.Error(w, "No input provided", http.StatusBadRequest)
			return
		}
		if _, ok := request.InputParams[api.OptCredType]; !ok {
			http.Error(w, "No input provided", http.StatusBadRequest)
			return
		}
		if request.InputParams[api.OptCredType] != "s3" &&
			request.InputParams[api.OptCredType] != "google" &&
			request.InputParams[api.OptCredType] != "azure" {
			http.Error(w, "Unsuported Cloud provider", http.StatusBadRequest)
			return

		}
		if request.InputParams[api.OptCredType] == "s3" {
			reqion := request.InputParams[api.OptCredRegion]
			endPoint := request.InputParams[api.OptCredEndpoint]
			accessKey := request.InputParams[api.OptCredAccessKey]
			secret := request.InputParams[api.OptCredSecretKey]
			if reqion == "" || endPoint == "" || accessKey == "" || secret == "" {
				http.Error(w, "No input provided", http.StatusBadRequest)
				return
			}
		}
		if request.InputParams[api.OptCredType] == "google" {
			projectID := request.InputParams[api.OptCredGoogleProjectID]
			jsonKey := request.InputParams[api.OptCredGoogleJsonKey]
			if projectID == "" || jsonKey == "" {
				http.Error(w, "No input provided", http.StatusBadRequest)
				return
			}
		}
		if request.InputParams[api.OptCredType] == "azure" {
			accName := request.InputParams[api.OptCredAzureAccountName]
			accessKey := request.InputParams[api.OptCredAzureAccountKey]
			if accName == "" || accessKey == "" {
				http.Error(w, "No input provided", http.StatusBadRequest)
				return
			}
		}
		json.NewEncoder(w).Encode(response)
	}))

	defer ts.Close()

	clnt, err := NewDriverClient(ts.URL, "pxd", "", "")
	require.NoError(t, err)

	input := make(map[string]string, 0)
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.Error(t, err)
	input[api.OptCredType] = "s3"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.Error(t, err)
	input[api.OptCredRegion] = "abc"
	input[api.OptCredEndpoint] = "http.xy.abc.bz.com"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.Error(t, err)
	input[api.OptCredAccessKey] = "myaccessley"
	input[api.OptCredSecretKey] = "OptCredSecretKey"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.NoError(t, err)

	input[api.OptCredType] = "google"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.Error(t, err)
	input[api.OptCredGoogleJsonKey] = "abc"
	input[api.OptCredGoogleProjectID] = "defgh34ijk"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.NoError(t, err)

	input[api.OptCredType] = "azure"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.Error(t, err)
	input[api.OptCredAzureAccountName] = "abc"
	input[api.OptCredAzureAccountKey] = "defgh34ijk"
	_, err = VolumeDriver(clnt).CredsCreate(input)
	require.NoError(t, err)
}

func TestClientCredsValidateiAndDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		uuid := r.Form.Get(api.OptCredUUID)
		if uuid == "" {
			http.Error(w, "Missing uuid param", http.StatusBadRequest)
			return
		}
		response := &api.VolumeResponse{}
		json.NewEncoder(w).Encode(response)

	}))
	defer ts.Close()

	//clnt, err := NewDriverClient(ts.URL, "pxd", volume.APIVersion, "")
	clnt, err := NewDriverClient(ts.URL, "pxd", "", "")
	require.NoError(t, err)

	err = VolumeDriver(clnt).CredsValidate("")
	require.Error(t, err)
	err = VolumeDriver(clnt).CredsValidate("adcs23-345678-2345qwe-kupl5890")
	require.NoError(t, err)

	err = VolumeDriver(clnt).CredsDelete("")
	require.Error(t, err)
	err = VolumeDriver(clnt).CredsDelete("adcs23-345678-2345qwe-kupl5890")
	require.NoError(t, err)

}

func TestClientCredsList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]interface{}, 0)
		json.NewEncoder(w).Encode(response)

	}))
	defer ts.Close()

	//clnt, err := NewDriverClient(ts.URL, "pxd", volume.APIVersion, "")
	clnt, err := NewDriverClient(ts.URL, "pxd", "", "")
	require.NoError(t, err)

	_, err = VolumeDriver(clnt).CredsList()
	require.NoError(t, err)

}
