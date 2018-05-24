package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/libopenstorage/openstorage/objectstore"
)

// swagger:operation GET /cluster/objectstore objectstore objectStoreInspect
//
// Lists Objectstore
//
// This will list current object stores
//
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: query
//   description: ID of objectstore to inspect
//   required: true
//   type: string
// responses:
//   '200':
//     description: success
//     schema:
//      $ref: '#/definitions/ObjectstoreInfo'
func (c *clusterApi) objectStoreInspect(w http.ResponseWriter, r *http.Request) {
	method := "objectStoreInspect"
	params := r.URL.Query()
	objstoreID := params[objectstore.ObjectStoreID]

	if len(objstoreID) == 0 || objstoreID[0] == "" {
		c.sendError(c.name, method, w, "Missing Objectstore ID", http.StatusBadRequest)
		return
	}

	objInfo, err := c.ObjectStoreManager.ObjectStoreInspect(objstoreID[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(objInfo)
}

// swagger:operation POST /cluster/objectstore objectstore objectStoreCreate
//
// Create an Object store
//
// This creates the volumes required to run the object store
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: volume on which object store to run
//   required: true
//   type: string
// responses:
//   '200':
//     description: success
//     schema:
//      $ref: '#/definitions/ObjectstoreInfo'
func (c *clusterApi) objectStoreCreate(w http.ResponseWriter, r *http.Request) {
	method := "objectStoreCreate"
	params := r.URL.Query()
	volumeName := params[objectstore.VolumeName]

	if len(volumeName) == 0 || volumeName[0] == "" {
		c.sendError(c.name, method, w, "Missing volume name", http.StatusBadRequest)
		return
	}

	objInfo, err := c.ObjectStoreManager.ObjectStoreCreate(volumeName[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(objInfo)
}

// swagger:operation PUT /cluster/objectstore objectstore objectStoreUpdate
//
// Updates object store
//
// This will enable/disable object store functionality.
//
// ---
// produces:
// - application/json
// parameters:
// - name: enable
//   in: query
//   description: enable/disable flag for object store
//   required: true
//   type: boolean
// - name: id
//   in: query
//   description: ID of objectstore to update
//   type: string
// responses:
//   '200':
//     description: success
func (c *clusterApi) objectStoreUpdate(w http.ResponseWriter, r *http.Request) {
	method := "objectStoreUpdate"
	params := r.URL.Query()
	strEnable := params[objectstore.Enable]
	objstoreID := params[objectstore.ObjectStoreID]

	if len(objstoreID) == 0 || objstoreID[0] == "" {
		c.sendError(c.name, method, w, "Missing Objectstore ID", http.StatusBadRequest)
		return
	}

	if len(strEnable) == 0 && strEnable[0] == "" {
		c.sendError(c.name, method, w, "enable parameter not set", http.StatusInternalServerError)
		return
	}

	enable, err := strconv.ParseBool(strEnable[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = c.ObjectStoreManager.ObjectStoreUpdate(objstoreID[0], enable)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation DELETE /cluster/objectstore objectstore objectStoreDelete
//
// Delete object store
//
// This will delete object store on node
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: ID of objectstore to delete
//   type: string
// responses:
//   '200':
//     description: success
func (c *clusterApi) objectStoreDelete(w http.ResponseWriter, r *http.Request) {
	method := "objectStoreDelete"
	params := r.URL.Query()
	objstoreID := params[objectstore.ObjectStoreID]

	if len(objstoreID) == 0 || objstoreID[0] == "" {
		c.sendError(c.name, method, w, "Missing Objectstore ID", http.StatusBadRequest)
		return
	}
	err := c.ObjectStoreManager.ObjectStoreDelete(objstoreID[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
