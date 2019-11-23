package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
)

func (vd *volAPI) credsEnumerate(w http.ResponseWriter, r *http.Request) {
	method := "credsEnumerate"

	ctx, err := vd.annotateContext(r)

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	client := api.NewOpenStorageCredentialsClient(conn)
	resp, err := client.Enumerate(ctx, &api.SdkCredentialEnumerateRequest{})

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	credentials := make(map[string]interface{})

	for _, credentialId := range resp.CredentialIds {
		credResp, err := client.Inspect(ctx, &api.SdkCredentialInspectRequest{
			CredentialId: credentialId,
		})

		if err != nil {
			vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		credentials[credResp.CredentialId] = credResp
	}

	if err := json.NewEncoder(w).Encode(credentials); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (vd *volAPI) credsCreate(w http.ResponseWriter, r *http.Request) {
	method := "credsCreate"
	var (
		err error
	)

	input := api.CredCreateRequest{
		InputParams: make(map[string]string),
	}

	if err = json.NewDecoder(r.Body).Decode(&input.InputParams); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	useProxy := false

	if len(input.InputParams[api.OptCredProxy]) != 0 {
		useProxy, err = strconv.ParseBool(input.InputParams[api.OptCredProxy])

		if err != nil {
			vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	sdkCreds, err := api.GetSdkCreds(input.InputParams)

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkReq := &api.SdkCredentialCreateRequest{
		Name:           input.InputParams[api.OptName],
		Bucket:         input.InputParams[api.OptCredBucket],
		EncryptionKey:  input.InputParams[api.OptCredEncrKey],
		UseProxy:       useProxy,
		CredentialType: sdkCreds,
	}
	ctx, err := vd.annotateContext(r)

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := api.NewOpenStorageCredentialsClient(conn)
	resp, err := client.Create(ctx, sdkReq)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&api.CredCreateResponse{
		UUID: resp.CredentialId,
	}); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (vd *volAPI) credsDelete(w http.ResponseWriter, r *http.Request) {
	method := "credsDelete"
	vars := mux.Vars(r)
	uuid, ok := vars["uuid"]

	if !ok {
		vd.sendError(vd.name, method, w, "Could not parse form for uuid", http.StatusBadRequest)
		return
	}

	// Initialize volume driver with connection
	ctx, err := vd.annotateContext(r)

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := api.NewOpenStorageCredentialsClient(conn)

	_, err = client.Delete(ctx, &api.SdkCredentialDeleteRequest{
		CredentialId: uuid,
	})

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (vd *volAPI) credsValidate(w http.ResponseWriter, r *http.Request) {
	method := "credsValidate"
	vars := mux.Vars(r)
	uuid, ok := vars["uuid"]
	if !ok {
		vd.sendError(vd.name, method, w, "Could not parse form for uuid", http.StatusBadRequest)
		return
	}
	// Initialize volume driver with connection
	ctx, err := vd.annotateContext(r)

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := api.NewOpenStorageCredentialsClient(conn)

	_, err = client.Validate(ctx, &api.SdkCredentialValidateRequest{
		CredentialId: uuid,
	})

	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)
}
