package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
)

func (vd *volAPI) credsEnumerate(w http.ResponseWriter, r *http.Request) {
	method := "credsEnumerate"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	creds, err := d.CredsEnumerate()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(creds)
}

func (vd *volAPI) credsCreate(w http.ResponseWriter, r *http.Request) {
	method := "credsCreate"
	var input api.CredCreateRequest
	response := &api.CredCreateResponse{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	response.UUID, err = d.CredsCreate(input.InputParams)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func (vd *volAPI) credsDelete(w http.ResponseWriter, r *http.Request) {
	method := "credsDelete"
	vars := mux.Vars(r)
	uuid, ok := vars["uuid"]
	if !ok {
		vd.sendError(vd.name, method, w, "Could not parse form for uuid", http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	if err = d.CredsDelete(uuid); err != nil {
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
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	if err := d.CredsValidate(uuid); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
