package server

import (
	"encoding/json"
	"net/http"

	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/secrets"
)

// swagger:operation PUT /cluster/secrets/defaultsecretkey secrets setDefaultSecretKey
//
// Set cluster secret key
//
// This will set the cluster wide default secret key
//
// ---
// produces:
// - application/json
// parameters:
// - name: defaultkey
//   in: body
//   description: default secret key
//   required: true
//   schema:
//    $ref: '#/definitions/DefaultSecretKeyRequest'
// responses:
//   '200':
//     description: success
func (c *clusterApi) setDefaultSecretKey(w http.ResponseWriter, r *http.Request) {
	method := "setDefaultSecretKey"
	var secReq secrets.DefaultSecretKeyRequest

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.SecretSetDefaultSecretKey(secReq.DefaultSecretKey, secReq.Override)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation GET /cluster/secrets/defaultsecretkey secrets getDefaultSecretKey
//
// Get cluster secret key
//
// This will return the cluster wide secret key
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: returns cluster wide secret key
//      schema:
//       $ref: '#/definitions/GetSecretResponse'
func (c *clusterApi) getDefaultSecretKey(w http.ResponseWriter, r *http.Request) {
	method := "getDefaultSecretKey"

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secretValue, err := inst.SecretGetDefaultSecretKey()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp := &secrets.GetSecretResponse{
		SecretValue: secretValue,
	}
	json.NewEncoder(w).Encode(secResp)
}

// swagger:operation POST /cluster/secrets/login secrets secretsLogin
//
// Start session with secret store
//
// This will initiate session with secret store
//
// ---
// produces:
// - application/json
// parameters:
// - name: SecretLoginConfig
//   in: body
//   description: config for login to secret store
//   required: true
//   schema:
//    $ref: '#/definitions/SecretLoginRequest'
// responses:
//   '200':
//     description: success
func (c *clusterApi) secretsLogin(w http.ResponseWriter, r *http.Request) {
	var secReq secrets.SecretLoginRequest
	method := "secretsLogin"

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.SecretLogin(secReq.SecretType, secReq.SecretConfig)
	if err != nil {
		if err == secrets.ErrNotAuthenticated {
			c.sendError(c.name, method, w, err.Error(), http.StatusUnauthorized)
			return
		}
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation PUT /cluster/secrets/ secrets setSeceret
//
// Set Secret Value
//
// This will set secrets data/value against given key
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: key/id for secrets
//   required: true
//   type: string
// - name: secretvalue
//   in : body
//   description:  value/data for secrets
//   required: true
//   schema:
//    $ref: '#/definitions/SetSecretRequest'
// responses:
//   '200':
//     description: success
func (c *clusterApi) setSecret(w http.ResponseWriter, r *http.Request) {
	method := "setSecret"
	var secReq secrets.SetSecretRequest
	params := r.URL.Query()
	secretID := params[secrets.SecretKey]

	if len(secretID) == 0 || secretID[0] == "" {
		c.sendError(c.name, method, w, "Missing secret ID", http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.SecretSet(secretID[0], secReq.SecretValue)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation GET /cluster/secrets/ secrets getSecret
//
// Get the seceret value/data for given key
//
// This will return the value/data for given secret key
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: secret id/key whose value to be retrived
//   type: string
//   required: true
// responses:
//   '200':
//      description: returns the value/data for given key
//      schema:
//       $ref: '#/definitions/GetSecretResponse'
func (c *clusterApi) getSecret(w http.ResponseWriter, r *http.Request) {
	method := "getSecret"
	params := r.URL.Query()
	secretID := params[secrets.SecretKey]

	if len(secretID) == 0 || secretID[0] == "" {
		c.sendError(c.name, method, w, "Missing secret ID", http.StatusBadRequest)
		return
	}

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secretValue, err := inst.SecretGet(secretID[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp := &secrets.GetSecretResponse{
		SecretValue: secretValue,
	}
	json.NewEncoder(w).Encode(secResp)
}

// swagger:operation GET /cluster/secrets/verify secrets secretLoginCheck
//
// Validates session with secret store
//
// This will return error if session is not estabilished with secrets store
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: validates session with secret store
func (c *clusterApi) secretLoginCheck(w http.ResponseWriter, r *http.Request) {
	method := "secretLoginCheck"

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.SecretCheckLogin()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
