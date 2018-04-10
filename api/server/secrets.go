package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/secrets"
)

const (
	secretKeyOkMsg   = "Secret Key set successfully"
	secretLoginOkMsg = "Secrets Login Succeeded"
	secretLoginCheck = "Secrets Login Check Succeeded"
)

// TODO: Add swagger yaml
func (c *clusterApi) setDefaultSecretKey(w http.ResponseWriter, r *http.Request) {

	method := "setDefaultSecretKey"
	var secReq secrets.DefaultSecretKeyRequest
	var secResp secrets.SecretResponse

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := c.SecretManager.SetDefaultSecretKey(
		secReq.DefaultSecretKey,
		secReq.Override,
	)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp.Status = "Default " + secretKeyOkMsg
	json.NewEncoder(w).Encode(secResp)
}

// TODO: Add swagger yaml
func (c *clusterApi) getDefaultSecretKey(w http.ResponseWriter, r *http.Request) {

	method := "getDefaultSecretKey"

	secretValue, err := c.SecretManager.GetDefaultSecretKey()
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
	var secResp secrets.SecretResponse
	method := "secretsLogin"

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := c.SecretManager.Login(secReq.SecretType, secReq.SecretConfig)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp.Status = secretLoginOkMsg
	json.NewEncoder(w).Encode(secResp)
}

// TODO: Add swagger yaml
func (c *clusterApi) setSecret(w http.ResponseWriter, r *http.Request) {

	method := "setSecret"
	var secReq secrets.SetSecretRequest
	var secResp secrets.SecretResponse
	params := r.URL.Query()
	secretID := params[secrets.SecretKey]

	if len(secretID) == 0 || secretID[0] == "" {
		c.sendError(c.name, method, w, "Missing secret ID", http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&secReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := c.SecretManager.Set(secretID[0], secReq.SecretValue)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp.Status = secretKeyOkMsg
	json.NewEncoder(w).Encode(secResp)
}

// TODO: Add swagger yaml
func (c *clusterApi) getSecret(w http.ResponseWriter, r *http.Request) {

	method := "getSecret"
	params := r.URL.Query()
	secretID := params[secrets.SecretKey]

	if len(secretID) == 0 || secretID[0] == "" {
		c.sendError(c.name, method, w, "Missing secret ID", http.StatusInternalServerError)
		return
	}

	secretValue, err := c.SecretManager.Get(secretID[0])
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

	var secResp secrets.SecretResponse
	method := "secretLoginCheck"
	err := c.SecretManager.CheckLogin()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secResp.Status = secretLoginCheck
	json.NewEncoder(w).Encode(secResp)
}
