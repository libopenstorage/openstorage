package server

import (
	"encoding/json"
	"net/http"

	"github.com/libopenstorage/openstorage/services"
)

// swagger:operation POST /cluster/service/entermaintenance service serviceEnterMaintenanceMode
//
// Enter Maintenance mode
//
// This will take node out of cluster and allows user to perform physical maintenance of node.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: success
func (c *clusterApi) serviceEnterMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	method := "serviceEnterMaintenanceMode"

	err := c.ServiceManager.ServiceEnterMaintenanceMode(false)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation PUT /cluster/service/exitmaintenance service serviceExitMaintenanceMode
//
// Exit Maintenance mode
//
// This will put node back into cluster
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: success
func (c *clusterApi) serviceExitMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	method := "serviceExitMaintenanceMode"

	err := c.ServiceManager.ServiceExitMaintenanceMode()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation POST /cluster/service/drive service serviceAddDrive
//
// Add drive to cluster storage pool
//
// This will add drive to cluster storage pool
//
// ---
// produces:
// - application/json
// parameters:
// - name: adddrive
//   in: body
//   description: params to perform add drive operation
//   required: true
//   schema:
//    $ref: '#definitions/AddDrive'
// responses:
//   '200':
//     description: success
//     schema:
//      $ref: '#/definitions/ServiceMessage'
func (c *clusterApi) serviceAddDrive(w http.ResponseWriter, r *http.Request) {
	method := "serviceAddDrive"
	var srvReq services.AddDrive

	if err := json.NewDecoder(r.Body).Decode(&srvReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := c.ServiceManager.ServiceAddDrive(srvReq.Operation, srvReq.Drive, srvReq.Journal)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	srvResp := &services.ServiceMessage{
		Status:  status,
		Version: "v1",
	}

	json.NewEncoder(w).Encode(srvResp)
}

// swagger:operation PUT /cluster/service/drive service serviceReplaceDrive
//
// Replace old drive with new drive
//
// This will allow you to replace old drive with new drive
//
// ---
// produces:
// - application/json
// parameters:
// - name: replacedrive
//   in: body
//   description: Source and target drive to perform drive replace operation.
//   required: true
//   schema:
//    $ref: '#/definitions/ReplaceDrive'
// responses:
//   '200':
//     description: success
//     schema:
//      $ref: '#/definitions/ServiceMessage'
func (c *clusterApi) serviceReplaceDrive(w http.ResponseWriter, r *http.Request) {
	method := "serviceReplaceDrive"
	var srvReq services.ReplaceDrive

	if err := json.NewDecoder(r.Body).Decode(&srvReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := c.ServiceManager.ServiceReplaceDrive(srvReq.Operation, srvReq.Source, srvReq.Target)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	srvResp := &services.ServiceMessage{
		Status:  status,
		Version: "v1",
	}

	json.NewEncoder(w).Encode(srvResp)
}

// swagger:operation PUT /cluster/service/rebalancepool service serviceRebalancePool
//
// Rebalance Pool
//
// Pool rebalance does spread data across all available drives in the pool.
//
// ---
// produces:
// - application/json
// parameters:
// - name: rebalancepool
//   in: body
//   description : Params to perform Pool rebalance
//   required: true
//   schema:
//    $ref: '#/definitions/RebalancePool'
// responses:
//   '200':
//     description: success
//     schema:
//      $ref: '#/definitions/ServiceMessage'
func (c *clusterApi) serviceRebalancePool(w http.ResponseWriter, r *http.Request) {
	method := "serviceRebalancePool"
	var srvReq services.RebalancePool

	if err := json.NewDecoder(r.Body).Decode(&srvReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	status, err := c.ServiceManager.ServiceRebalancePool(srvReq.Operation, srvReq.PoolID)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	srvResp := &services.ServiceMessage{
		Status:  status,
		Version: "v1",
	}

	json.NewEncoder(w).Encode(srvResp)
}
