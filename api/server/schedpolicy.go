package server

import (
	"encoding/json"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	"net/http"

	"github.com/gorilla/mux"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"

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
	ctx, err := c.annotateContext(r)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		schedulePolicyClient := api.NewOpenStorageSchedulePolicyClient(conn)
		resp, err := schedulePolicyClient.Enumerate(ctx, &api.SdkSchedulePolicyEnumerateRequest{})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		schedPolicies := make([]*sched.SchedPolicy, 0, len(resp.Policies))
		for _, policy := range resp.Policies {

			policyBytes, err := sdk.SdkSchedToRetainInternalSpecYamlByte(policy.Schedules)

			if err != nil {
				c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
				return
			}

			schedPolicies = append(schedPolicies, &sched.SchedPolicy{
				Name:     policy.Name,
				Schedule: string(policyBytes),
			})
		}

		if err := json.NewEncoder(w).Encode(schedPolicies); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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

	ctx, err := c.annotateContext(r)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		schedulePolicyClient := api.NewOpenStorageSchedulePolicyClient(conn)
		resp, err := schedulePolicyClient.Inspect(ctx, &api.SdkSchedulePolicyInspectRequest{
			Name: schedName,
		})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		policyBytes, err := sdk.SdkSchedToRetainInternalSpecYamlByte(resp.Policy.Schedules)

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		schedPolicy := &sched.SchedPolicy{
			Name:     resp.Policy.Name,
			Schedule: string(policyBytes),
		}

		if err := json.NewEncoder(w).Encode(schedPolicy); err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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

	ctx, err := c.annotateContext(r)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		schedulePolicyClient := api.NewOpenStorageSchedulePolicyClient(conn)

		intervals, err := sdk.RetainInternalSpecYamlByteToSdkSched([]byte(schedReq.Schedule))

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = schedulePolicyClient.Create(ctx, &api.SdkSchedulePolicyCreateRequest{
			SchedulePolicy: &api.SdkSchedulePolicy{
				Name:      schedReq.Name,
				Schedules: intervals,
			},
		})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
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

	ctx, err := c.annotateContext(r)

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	if conn, err := c.getConn(); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		schedulePolicyClient := api.NewOpenStorageSchedulePolicyClient(conn)

		intervals, err := sdk.RetainInternalSpecYamlByteToSdkSched([]byte(schedReq.Schedule))

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = schedulePolicyClient.Update(ctx, &api.SdkSchedulePolicyUpdateRequest{
			SchedulePolicy: &api.SdkSchedulePolicy{
				Name:      schedReq.Name,
				Schedules: intervals,
			},
		})

		if err != nil {
			c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
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

	inst, err := clustermanager.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.SchedPolicyDelete(schedName)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
