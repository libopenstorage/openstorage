package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/libopenstorage/openstorage/secrets"
)

const (
	secretKeyOkMsg   = "Secret Key set successfully."
	secretLoginOkMsg = "Secrets Login Succeeded"
	secretLoginCheck = "Secrets Login Check Succeeded"
)

//TODO: Add swagger yaml
func (c *clusterApi) setClusterSecretKey(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	method := "setClusterSecretKey"
	secretKey := params["clustersecretkey"][0]
	override, _ := strconv.ParseBool(params["override"][0])

	if secretKey == "" {
		c.sendError(c.name, method, w, "Missing cluster key", http.StatusInternalServerError)
		return
	}

	err := c.SecretManager.Secret.SetClusterSecretKey(secretKey, override)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Cluster" + secretKeyOkMsg + "\n"))
}

//TODO: Add swagger yaml
func (c *clusterApi) secretsLogin(w http.ResponseWriter, r *http.Request) {

	var dcReq secrets.SecretLoginRequest
	method := "secretsLogin"
	params := r.URL.Query()

	if err := json.NewDecoder(r.Body).Decode(&dcReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	secretStore, convErr := strconv.ParseInt(params["secret"][0], 10, 64)
	if convErr != nil {
		c.sendError(c.name, method, w, convErr.Error(), http.StatusInternalServerError)
		return
	}

	err := c.SecretManager.Secret.SecretLogin(int(secretStore), dcReq.SecretConfig)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(secretLoginOkMsg + "\n"))
}

//TODO: Add swagger yaml
func (c *clusterApi) setSecret(w http.ResponseWriter, r *http.Request) {

	method := "setSecret"
	params := r.URL.Query()
	secretKey := params["secretid"][0]
	secretValue := params["secretvalue"][0]

	if secretKey == "" {
		c.sendError(c.name, method, w, "Missing secret key", http.StatusInternalServerError)
		return
	}
	if secretValue == "" {
		c.sendError(c.name, method, w, "Missing secret value", http.StatusInternalServerError)
		return
	}

	err := c.SecretManager.Secret.SetSecret(secretKey, secretValue)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(secretKeyOkMsg + "\n"))
}

//TODO: Add swagger yaml
func (c *clusterApi) getSecret(w http.ResponseWriter, r *http.Request) {

	method := "getSecret"
	params := r.URL.Query()
	secretid := params["secretid"][0]

	if secretid == "" {
		c.sendError(c.name, method, w, "Missing secret ID", http.StatusInternalServerError)
		return
	}

	secretValue, err := c.SecretManager.Secret.GetSecret(secretid)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(secretValue))
}

//TODO: Add swagger yaml
func (c *clusterApi) secretLoginCheck(w http.ResponseWriter, r *http.Request) {

	method := "secretLoginCheck"
	err := c.SecretManager.Secret.CheckSecretLogin()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(secretLoginCheck + "\n"))
}
