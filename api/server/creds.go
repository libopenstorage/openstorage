package server

import (
	"encoding/json"
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	"net/http"
)

func (vd *volAPI) credsEnumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	method := "credsEnumerate"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	creds, err := d.CredsEnumerate()
	if err != nil {
		e := fmt.Errorf("Failed to get credential list: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(creds)
}

func (vd *volAPI) credsCreate(w http.ResponseWriter, r *http.Request) {
	var err error
	var input api.CredCreateRequest
	Response := &api.CredCreateResponse{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Response.CredErr = err
		json.NewEncoder(w).Encode(Response)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	Response.UUID, Response.CredErr = d.CredsCreate(input.InputParams)
	json.NewEncoder(w).Encode(Response)
}

func (vd *volAPI) credsDelete(w http.ResponseWriter, r *http.Request) {
	var err error
	volumeResponse := &api.VolumeResponse{}

	err = r.ParseForm()
	if err != nil {
		volumeResponse.Error = err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	uuid := r.Form.Get(api.OptCredUUID)
	if uuid == "" {
		volumeResponse.Error = fmt.Sprintf("Missing uuid param")
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	if err := d.CredsDelete(uuid); err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

func (vd *volAPI) credsValidate(w http.ResponseWriter, r *http.Request) {
	var err error
	volumeResponse := &api.VolumeResponse{}
	err = r.ParseForm()
	if err != nil {
		volumeResponse.Error = err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	uuid := r.Form.Get(api.OptCredUUID)
	if uuid == "" {
		volumeResponse.Error = fmt.Sprintf("Missing uuid param")
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	if err := d.CredsValidate(uuid); err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}
