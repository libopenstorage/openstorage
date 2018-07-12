package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	sched "github.com/libopenstorage/openstorage/schedpolicy"
)

// swagger:operation GET /cluster/schedpolicy schedpolicy schedPolicyEnumerate
//
// List schedule policies
//
// This will list all of schedule policy
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: success
//     schema:
//        type: array
//        items:
//           $ref: '#/definitions/SchedPolicy'
func (c *clusterApi) schedPolicyEnumerate(w http.ResponseWriter, r *http.Request) {
	method := "schedPolicyEnumerate"
	schedPolicies, err := c.SchedPolicyManager.SchedPolicyEnumerate()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(schedPolicies)
}

// swagger:operation GET /cluster/schedpolicy/{name} schedpolicy schedPolicyGet
//
// Get policy details
//
// This will return the requested schedule policy details
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: path
//   description: Retrive details of given policy name
//   required: true
//   type: string
// responses:
//   '200':
//     description: success
//     schema:
//      $ref: '#/definitions/SchedPolicy'
func (c *clusterApi) schedPolicyGet(w http.ResponseWriter, r *http.Request) {
	method := "schedPolicyGet"
	vars := mux.Vars(r)
	schedName, ok := vars[sched.SchedName]

	if !ok || schedName == "" {
		c.sendError(c.name, method, w, "Missing Schedule Policy Name", http.StatusBadRequest)
		return
	}

	schedPolicy, err := c.SchedPolicyManager.SchedPolicyGet(schedName)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(schedPolicy)
}

// swagger:operation POST /cluster/schedpolicy schedpolicy schedPolicyCreate
//
// Create schedule policy
//
// This creates scheudle policy which will allow user to create snapshot schedule
//
// ---
// produces:
// - application/json
// parameters:
// - name: schedpolicy
//   in: body
//   description: policy name and schedule to create
//   required: true
//   schema:
//    $ref: '#/definitions/SchedPolicy'
// responses:
//   '200':
//     description: success
func (c *clusterApi) schedPolicyCreate(w http.ResponseWriter, r *http.Request) {

	method := "schedPolicyCreate"
	var schedReq sched.SchedPolicy

	if err := json.NewDecoder(r.Body).Decode(&schedReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.SchedPolicyManager.SchedPolicyCreate(schedReq.Name, schedReq.Schedule)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation PUT /cluster/schedpolicy schedpolicy schedPolicyUpdate
//
// Update schedule policy
//
// This will update specified schedule policy
//
// ---
// produces:
// - application/json
// parameters:
// - name: schedpolicy
//   in: body
//   description: policy name and schedule to update
//   required: true
//   schema:
//    $ref: '#/definitions/SchedPolicy'
// responses:
//   '200':
//     description: success
func (c *clusterApi) schedPolicyUpdate(w http.ResponseWriter, r *http.Request) {

	method := "schedPolicyUpdate"
	var schedReq sched.SchedPolicy

	if err := json.NewDecoder(r.Body).Decode(&schedReq); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.SchedPolicyManager.SchedPolicyUpdate(schedReq.Name, schedReq.Schedule)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// swagger:operation DELETE /cluster/schedpolicy/{name} schedpolicy shedPolicyDelete
//
// Delete schedule policy
//
// This will delete specified schedule policy
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: path
//   description: policy name and schedule to delete
//   required: true
//   type: string
// responses:
//   '200':
//     description: success
func (c *clusterApi) schedPolicyDelete(w http.ResponseWriter, r *http.Request) {

	method := "schedPolicyDelete"

	vars := mux.Vars(r)
	schedName, ok := vars[sched.SchedName]

	if !ok || schedName == "" {
		c.sendError(c.name, method, w, "Missing Schedule Policy Name", http.StatusBadRequest)
		return
	}
	err := c.SchedPolicyManager.SchedPolicyDelete(schedName)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
