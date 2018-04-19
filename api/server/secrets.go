package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/secrets"
)

// TODO: Add swagger yaml
func (c *clusterApi) setDefaultSecretKey(w http.ResponseWriter, r *http.Request) {

	method := "setDefaultSecretKey"
	var secReq secrets.DefaultSecretKeyRequest

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.SecretManager.SecretSetDefaultSecretKey(
		secReq.DefaultSecretKey,
		secReq.Override,
	)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// TODO: Add swagger yaml
func (c *clusterApi) getDefaultSecretKey(w http.ResponseWriter, r *http.Request) {

	method := "getDefaultSecretKey"

	secretValue, err := c.SecretManager.SecretGetDefaultSecretKey()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp := &secrets.GetSecretResponse{
		SecretValue: secretValue,
	}
	json.NewEncoder(w).Encode(secResp)
}

// TODO: Add swagger yaml
func (c *clusterApi) secretsLogin(w http.ResponseWriter, r *http.Request) {
	var secReq secrets.SecretLoginRequest
	method := "secretsLogin"

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.SecretManager.SecretLogin(secReq.SecretType, secReq.SecretConfig)
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

// TODO: Add swagger yaml
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

	err := c.SecretManager.SecretSet(secretID[0], secReq.SecretValue)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// TODO: Add swagger yaml
func (c *clusterApi) getSecret(w http.ResponseWriter, r *http.Request) {

	method := "getSecret"
	params := r.URL.Query()
	secretID := params[secrets.SecretKey]

	if len(secretID) == 0 || secretID[0] == "" {
		c.sendError(c.name, method, w, "Missing secret ID", http.StatusBadRequest)
		return
	}

	secretValue, err := c.SecretManager.SecretGet(secretID[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp := &secrets.GetSecretResponse{
		SecretValue: secretValue,
	}

	json.NewEncoder(w).Encode(secResp)
}

// TODO: Add swagger yaml
func (c *clusterApi) secretLoginCheck(w http.ResponseWriter, r *http.Request) {

	method := "secretLoginCheck"
	err := c.SecretManager.SecretCheckLogin()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
