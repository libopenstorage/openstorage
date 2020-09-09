package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/urfave/negroni"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/server/sdk"
	clustermanager "github.com/libopenstorage/openstorage/cluster/manager"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/options"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
)

const (
	schedDriverPostFix = "-sched"

	// We set it to 128Mi to support large number of volumes. Before, the client
	// was using 4Mi and it would not allow the support of over 5k volumes.
	// We increased to a very large value to support over 100k volumes.
	maxMsgSize = 128 * 1024 * 1024
)

type volAPI struct {
	restBase

	sdkUds   string
	conn     *grpc.ClientConn
	dummyMux *runtime.ServeMux
	mu       sync.Mutex
}

func responseStatus(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func newVolumeAPI(name, sdkUds string) restServer {
	return &volAPI{
		restBase: restBase{version: volume.APIVersion, name: name},
		sdkUds:   sdkUds,
		dummyMux: runtime.NewServeMux(),
	}
}

func (vd *volAPI) String() string {
	return vd.name
}

func (vd *volAPI) getConn() (*grpc.ClientConn, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()
	if vd.conn == nil {
		var err error
		vd.conn, err = grpcserver.Connect(
			vd.sdkUds,
			[]grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)),
			})
		if err != nil {
			return nil, fmt.Errorf("Failed to connect to gRPC handler: %v", err)
		}
	}
	return vd.conn, nil
}

func (vd *volAPI) annotateContext(r *http.Request) (context.Context, error) {
	// This creates a context and populates the authentication token
	// using the same function as the SDK REST Gateway
	ctx, err := runtime.AnnotateContext(context.Background(), vd.dummyMux, r)
	if err != nil {
		return ctx, err
	}
	// If a header exists in the request fetch the requested driver name if provided
	// and pass it in the grpc context as a metadata key value
	userAgent := r.Header.Get("User-Agent")
	if len(userAgent) > 0 {
		// Check if the request is coming from a container orchestrator
		clientName := strings.Split(userAgent, "/")
		if len(clientName) > 0 {
			return grpcserver.AddMetadataToContext(ctx, sdk.ContextDriverKey, clientName[0]), nil
		}
	}
	return ctx, nil
}

func (vd *volAPI) getVolDriver(r *http.Request) (volume.VolumeDriver, error) {
	// Check if the driver has registered by it's user agent name
	userAgent := r.Header.Get("User-Agent")
	if len(userAgent) > 0 {
		clientName := strings.Split(userAgent, "/")
		if len(clientName) > 0 {
			d, err := volumedrivers.Get(clientName[0])
			if err == nil {
				return d, nil
			}
		}
	}

	// Check if the driver has registered a scheduler-based driver
	d, err := volumedrivers.Get(vd.name + schedDriverPostFix)
	if err == nil {
		return d, nil
	}

	// default
	return volumedrivers.Get(vd.name)
}

func (vd *volAPI) parseID(r *http.Request) (string, error) {
	if id, err := vd.parseParam(r, "id"); err == nil {
		return id, nil
	}

	return "", fmt.Errorf("could not parse snap ID")
}

func (vd *volAPI) parseParam(r *http.Request, param string) (string, error) {
	vars := mux.Vars(r)
	if id, ok := vars[param]; ok {
		return id, nil
	}
	return "", fmt.Errorf("could not parse %s", param)
}

func (vd *volAPI) nodeIPtoIds(nodes []string) ([]string, error) {
	nodeIds := make([]string, 0)

	// Get cluster instance
	c, err := clustermanager.Inst()
	if err != nil {
		return nodeIds, err
	}

	if c == nil {
		return nodeIds, fmt.Errorf("failed to get cluster instance.")
	}

	for _, idIp := range nodes {
		if idIp != "" {
			id, err := c.GetNodeIdFromIp(idIp)
			if err != nil {
				return nodeIds, err
			}
			nodeIds = append(nodeIds, id)
		}
	}

	return nodeIds, err
}

// Convert any replica set node values which are IPs to the corresponding Node ID.
// Update the replica set node list.
func (vd *volAPI) updateReplicaSpecNodeIPstoIds(rspecRef *api.ReplicaSet) error {
	if rspecRef != nil && len(rspecRef.Nodes) > 0 {
		nodeIds, err := vd.nodeIPtoIds(rspecRef.Nodes)
		if err != nil {
			return err
		}

		if len(nodeIds) > 0 {
			rspecRef.Nodes = nodeIds
		}
	}

	return nil
}

// Creates a single volume with given spec.
func (vd *volAPI) create(w http.ResponseWriter, r *http.Request) {
	var dcRes api.VolumeCreateResponse
	var dcReq api.VolumeCreateRequest
	method := "create"

	if err := json.NewDecoder(r.Body).Decode(&dcReq); err != nil {
		fmt.Println("returning error here")
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	spec := dcReq.GetSpec()
	if spec.VolumeLabels == nil {
		spec.VolumeLabels = make(map[string]string)
	}
	for k, v := range dcReq.Locator.GetVolumeLabels() {
		spec.VolumeLabels[k] = v
	}

	volumes := api.NewOpenStorageVolumeClient(conn)
	id, err := volumes.Create(ctx, &api.SdkVolumeCreateRequest{
		Name:   dcReq.Locator.GetName(),
		Labels: dcReq.Locator.GetVolumeLabels(),
		Spec:   dcReq.GetSpec(),
	})

	dcRes.VolumeResponse = &api.VolumeResponse{Error: responseStatus(err)}
	if err == nil {
		dcRes.Id = id.GetVolumeId()
	}

	json.NewEncoder(w).Encode(&dcRes)
}

func processErrorForVolSetResponse(action *api.VolumeStateAction, err error, resp *api.VolumeSetResponse) {
	if err == nil || resp == nil {
		return
	}

	if action != nil && (action.IsUnMount() || action.IsDetach()) {
		if sdk.IsErrorNotFound(err) {
			resp.VolumeResponse = &api.VolumeResponse{}
			resp.Volume = &api.Volume{}
		} else {
			resp.VolumeResponse = &api.VolumeResponse{
				Error: err.Error(),
			}
		}
	} else if err != nil {
		resp.VolumeResponse = &api.VolumeResponse{
			Error: err.Error(),
		}
	}
}

// swagger:operation PUT /osd-volumes/{id} volume setVolume
//
// Updates a single volume with given spec.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// - name: spec
//   in: body
//   description: spec to set volume with
//   required: true
//   schema:
//         "$ref": "#/definitions/VolumeSetRequest"
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeSetResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeSetResponse"
func (vd *volAPI) volumeSet(w http.ResponseWriter, r *http.Request) {
	var (
		volumeID string
		err      error
		req      api.VolumeSetRequest
		resp     api.VolumeSetResponse
	)
	method := "volumeSet"

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	if volumeID, err = vd.parseID(r); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	setActions := ""
	if req.Action != nil {
		setActions = fmt.Sprintf("Mount=%v Attach=%v", req.Action.Mount, req.Action.Attach)
	}

	vd.logRequest(method, string(volumeID)).Infoln(setActions)
	volumes := api.NewOpenStorageVolumeClient(conn)
	mountAttachClient := api.NewOpenStorageMountAttachClient(conn)

	detachOptions := &api.SdkVolumeDetachOptions{}
	attachOptions := &api.SdkVolumeAttachOptions{}
	if req.Options["SECRET_NAME"] != "" {
		attachOptions.SecretName = req.Options["SECRET_NAME"]
	}
	if req.Options["SECRET_KEY"] != "" {
		attachOptions.SecretKey = req.Options["SECRET_KEY"]
	}
	if req.Options["SECRET_CONTEXT"] != "" {
		attachOptions.SecretContext = req.Options["SECRET_CONTEXT"]
	}
	if req.Options[options.OptionsForceDetach] == "true" {
		detachOptions.Force = true
	}
	if req.Options[options.OptionsUnmountBeforeDetach] == "true" {
		detachOptions.UnmountBeforeDetach = true
	}
	if req.Options[options.OptionsRedirectDetach] == "true" {
		detachOptions.Redirect = true
	}
	if req.Options["FASTPATH_STATE"] != "" {
		attachOptions.Fastpath = req.Options["FASTPATH_STATE"]
	}

	unmountOptions := &api.SdkVolumeUnmountOptions{}
	if req.Options["DELETE_AFTER_UNMOUNT"] == "true" {
		unmountOptions.DeleteMountPath = true
	}
	if req.Options["WAIT_BEFORE_DELETE"] == "true" {
		unmountOptions.NoDelayBeforeDeletingMountPath = false
	} else {
		unmountOptions.NoDelayBeforeDeletingMountPath = true
	}

	if req.Locator != nil || req.Spec != nil {
		// Only update spec if spec and locator are not nil.
		vol, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
			VolumeId: volumeID,
		})
		if err != nil {
			// Return error here for ha-update operation where the Action is nil
			if !sdk.IsErrorNotFound(err) || req.Action == nil {
				vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
				return
			}

			vd.logRequest(method, string(volumeID)).Infoln("Ignoring unmount/detach action on deleted volume.")
		} else {
			updateReq := &api.SdkVolumeUpdateRequest{VolumeId: volumeID}
			if req.Locator != nil && len(req.Locator.VolumeLabels) > 0 {
				updateReq.Labels = req.Locator.VolumeLabels
			}
			if req.Spec != nil {
				if err = vd.updateReplicaSpecNodeIPstoIds(req.Spec.ReplicaSet); err != nil {
					vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
					return
				}

				updateReq.Spec = getVolumeUpdateSpec(req.Spec, vol.GetVolume())
			}

			if _, err := volumes.Update(ctx, updateReq); err != nil {
				vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	if req.Action != nil {
		if req.Action.IsAttach() {
			_, err = mountAttachClient.Attach(ctx, &api.SdkVolumeAttachRequest{
				VolumeId:      volumeID,
				Options:       attachOptions,
				DriverOptions: req.GetOptions(),
			})
		} else if req.Action.IsDetach() {
			_, err = mountAttachClient.Detach(ctx, &api.SdkVolumeDetachRequest{
				VolumeId:      volumeID,
				Options:       detachOptions,
				DriverOptions: req.GetOptions(),
			})
		}

		if err == nil {
			if req.Action.IsMount() {
				if req.Action.MountPath == "" {
					err = fmt.Errorf("Invalid mount path")
				} else {
					_, err = mountAttachClient.Mount(ctx, &api.SdkVolumeMountRequest{
						VolumeId:      volumeID,
						MountPath:     req.Action.MountPath,
						DriverOptions: req.GetOptions(),
					})
				}
			} else if req.Action.IsUnMount() {
				_, err = mountAttachClient.Unmount(ctx, &api.SdkVolumeUnmountRequest{
					VolumeId:      volumeID,
					MountPath:     req.Action.MountPath,
					Options:       unmountOptions,
					DriverOptions: req.GetOptions(),
				})
			}
		}
	}

	resVol, err2 := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: volumeID,
		Options: &api.VolumeInspectOptions{
			Deep: true,
		},
	})
	if err2 != nil {
		resp.Volume = &api.Volume{}
		if req.Action.IsAttach() {
			resp.VolumeResponse = &api.VolumeResponse{
				Error: responseStatus(err2),
			}
		}
	} else {
		resp.Volume = resVol.GetVolume()
	}
	// Do not clear inspect err for attach
	if err != nil {
		resp.VolumeResponse = &api.VolumeResponse{
			Error: responseStatus(err),
		}
	}
	json.NewEncoder(w).Encode(resp)

}

func getVolumeUpdateSpec(spec *api.VolumeSpec, vol *api.Volume) *api.VolumeSpecUpdate {
	newSpec := &api.VolumeSpecUpdate{}
	if spec == nil {
		return newSpec
	}

	newSpec.ReplicaSet = spec.ReplicaSet
	if spec.Shared != vol.Spec.Shared {
		newSpec.SharedOpt = &api.VolumeSpecUpdate_Shared{
			Shared: spec.Shared,
		}
	}

	if spec.Sharedv4 != vol.Spec.Sharedv4 {
		newSpec.Sharedv4Opt = &api.VolumeSpecUpdate_Sharedv4{
			Sharedv4: spec.Sharedv4,
		}
	}

	if spec.Passphrase != vol.Spec.Passphrase {
		newSpec.PassphraseOpt = &api.VolumeSpecUpdate_Passphrase{
			Passphrase: spec.Passphrase,
		}
	}

	if spec.Cos != vol.Spec.Cos && spec.Cos != 0 {
		newSpec.CosOpt = &api.VolumeSpecUpdate_Cos{
			Cos: spec.Cos,
		}
	}

	if spec.Journal != vol.Spec.Journal {
		newSpec.JournalOpt = &api.VolumeSpecUpdate_Journal{
			Journal: spec.Journal,
		}
	}

	if spec.Nodiscard != vol.Spec.Nodiscard {
		newSpec.NodiscardOpt = &api.VolumeSpecUpdate_Nodiscard{
			Nodiscard: spec.Nodiscard,
		}
	}

	newSpec.IoStrategy = spec.IoStrategy

	if spec.Sticky != vol.Spec.Sticky {
		newSpec.StickyOpt = &api.VolumeSpecUpdate_Sticky{
			Sticky: spec.Sticky,
		}
	}

	if spec.Scale != vol.Spec.Scale {
		newSpec.ScaleOpt = &api.VolumeSpecUpdate_Scale{
			Scale: spec.Scale,
		}
	}

	if spec.Size != vol.Spec.Size {
		newSpec.SizeOpt = &api.VolumeSpecUpdate_Size{
			Size: spec.Size,
		}
	}

	if spec.IoProfile != vol.Spec.IoProfile {
		newSpec.IoProfileOpt = &api.VolumeSpecUpdate_IoProfile{
			IoProfile: spec.IoProfile,
		}
	}

	if spec.Dedupe != vol.Spec.Dedupe {
		newSpec.DedupeOpt = &api.VolumeSpecUpdate_Dedupe{
			Dedupe: spec.Dedupe,
		}
	}

	if spec.Sticky != vol.Spec.Sticky {
		newSpec.StickyOpt = &api.VolumeSpecUpdate_Sticky{
			Sticky: spec.Sticky,
		}
	}

	if spec.Group != vol.Spec.Group && spec.Group != nil {
		newSpec.GroupOpt = &api.VolumeSpecUpdate_Group{
			Group: spec.Group,
		}
	}

	if spec.QueueDepth != vol.Spec.QueueDepth {
		newSpec.QueueDepthOpt = &api.VolumeSpecUpdate_QueueDepth{
			QueueDepth: spec.QueueDepth,
		}
	}

	if spec.SnapshotSchedule != vol.Spec.SnapshotSchedule {
		newSpec.SnapshotScheduleOpt = &api.VolumeSpecUpdate_SnapshotSchedule{
			SnapshotSchedule: spec.SnapshotSchedule,
		}
	}

	if spec.SnapshotInterval != vol.Spec.SnapshotInterval && spec.SnapshotInterval != math.MaxUint32 {
		newSpec.SnapshotIntervalOpt = &api.VolumeSpecUpdate_SnapshotInterval{
			SnapshotInterval: spec.SnapshotInterval,
		}
	}

	if spec.HaLevel != vol.Spec.HaLevel && spec.HaLevel != 0 {
		newSpec.HaLevelOpt = &api.VolumeSpecUpdate_HaLevel{
			HaLevel: spec.HaLevel,
		}
	}
	if spec.ExportSpec != nil {
		newSpec.ExportSpecOpt = &api.VolumeSpecUpdate_ExportSpec{
			ExportSpec: spec.ExportSpec,
		}
	}
	if spec.MountOptions != nil {
		newSpec.MountOpt = &api.VolumeSpecUpdate_MountOptSpec{
			MountOptSpec: spec.MountOptions,
		}
	}
	if spec.Sharedv4MountOptions != nil {
		newSpec.Sharedv4MountOpt = &api.VolumeSpecUpdate_Sharedv4MountOptSpec{
			Sharedv4MountOptSpec: spec.Sharedv4MountOptions,
		}
	}

	if spec.FpPreference != vol.Spec.FpPreference {
		newSpec.FastpathOpt = &api.VolumeSpecUpdate_Fastpath{
			Fastpath: spec.FpPreference,
		}
	}

	if spec.Xattr != vol.Spec.Xattr {
		newSpec.XattrOpt = &api.VolumeSpecUpdate_Xattr{
			Xattr: spec.Xattr,
		}
	}
	if spec.ScanPolicy != nil {
		newSpec.ScanPolicyOpt = &api.VolumeSpecUpdate_ScanPolicy{
			ScanPolicy: spec.ScanPolicy,
		}
	}

	return newSpec
}

// swagger:operation GET /osd-volumes/{id} volume inspectVolume
//
// Inspect volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//   '200':
//     description: volume get response
//     schema:
//         "$ref": "#/definitions/Volume"
func (vd *volAPI) inspect(w http.ResponseWriter, r *http.Request) {
	var err error
	var volumeID string

	method := "inspect"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	dk, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: volumeID,
		Options: &api.VolumeInspectOptions{
			Deep: true,
		},
	})
	dkVolumes := []*api.Volume{}
	if err != nil {
		// The Kubernetes Portworx intree driver has a bug when it tries to
		// check if the server is up and running. It sends a request to this
		// server to get a version, but instead of using the correct API, it sends a
		// request to get information about a volume called "version".
		// Since the intree driver is _not_ sending a authenticated call for this
		// check, it will be denied as an unathorized request. We will need to
		// return a 200 HTTP request as if the volume was not found.
		if volumeID == "versions" {
			// This is repetition of code, but is simple to understand
			json.NewEncoder(w).Encode(dkVolumes)
			return
		}

		// SDK returns a NotFound error for an invalid volume
		// Previously the REST server returned an empty array if a volume was not found
		if s, ok := status.FromError(err); ok && s.Code() != codes.NotFound {
			vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
			return
		}
	} else {
		dkVolumes = append(dkVolumes, dk.GetVolume())
	}

	json.NewEncoder(w).Encode(dkVolumes)
}

// swagger:operation DELETE /osd-volumes/{id} volume deleteVolume
//
// Delete volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) delete(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	method := "delete"
	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	vd.logRequest(method, volumeID).Infoln("")

	volumeResponse := &api.VolumeResponse{}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{VolumeId: volumeID})
	if err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation GET /osd-volumes volume enumerateVolumes
//
// Enumerate all volumes
//
// ---
// consumes:
// - multipart/form-data
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: User specified volume name (Case Sensitive)
//   required: false
//   type: string
// - name: Label
//   in: formData
//   description: |
//    Comma separated name value pairs
//    example: {"label1","label2"}
//   required: false
//   type: string
// - name: ConfigLabel
//   in: formData
//   description: |
//    Comma separated name value pairs
//    example: {"label1","label2"}
//   required: false
//   type: string
// - name: VolumeID
//   in: query
//   description: Volume UUID
//   required: false
//   type: string
//   format: uuid
// responses:
//   '200':
//      description: an array of volumes
//      schema:
//         type: array
//         items:
//            $ref: '#/definitions/Volume'
func (vd *volAPI) enumerate(w http.ResponseWriter, r *http.Request) {
	var locator api.VolumeLocator
	var configLabels map[string]string
	var err error
	var vols []*api.Volume

	method := "enumerate"

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)

	params := r.URL.Query()
	v := params[string(api.OptName)]
	if v != nil {
		locator.Name = v[0]
	}

	v = params[string(api.OptLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &locator.VolumeLabels); err != nil {
			e := fmt.Errorf("Failed to parse parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}

	v = params[string(api.OptConfigLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &configLabels); err != nil {
			e := fmt.Errorf("Failed to parse parse configLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
		// Add config labels to locator object.
		for l, _ := range configLabels {
			locator.VolumeLabels[l] = configLabels[l]
		}
	}

	v = params[string(api.OptVolumeID)]
	if v != nil {
		vols = make([]*api.Volume, 0, len(v))
		for _, s := range v {
			// They asked for inspect of specific volumes. We must honor with deep inspects
			resp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
				VolumeId: string(s),
				Options: &api.VolumeInspectOptions{
					Deep: true,
				},
			})
			if err == nil {
				if resp.GetVolume() != nil {
					vols = append(vols, resp.GetVolume())
				}
			} else if sdk.IsErrorNotFound(err) {
				continue
			} else {
				e := fmt.Errorf("Failed to inspect volumeID: %s", err.Error())
				vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
				return
			}
		}
		json.NewEncoder(w).Encode(vols)
		return
	}

	// Enumerate and Inspect
	resp, err := volumes.InspectWithFilters(ctx, &api.SdkVolumeInspectWithFiltersRequest{
		Name:   locator.Name,
		Labels: locator.VolumeLabels,
	})
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	vols = make([]*api.Volume, len(resp.GetVolumes()))
	for i, vol := range resp.GetVolumes() {
		vols[i] = vol.GetVolume()
	}
	json.NewEncoder(w).Encode(vols)
}

// swagger:operation POST /osd-snapshots snapshot createSnap
//
// Take a snapshot of volume in SnapCreateRequest
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: id to get volume with
//   required: true
//   type: integer
// - name: spec
//   in: body
//   description: spec to create snap with
//   required: true
//   schema:
//    "$ref": "#/definitions/SnapCreateRequest"
// responses:
//    '200':
//      description: an array of volumes
//      schema:
//       "$ref": '#/definitions/SnapCreateResponse'
//    default:
//     description: unexpected error
//     schema:
//      "$ref": "#/definitions/SnapCreateResponse"
func (vd *volAPI) snap(w http.ResponseWriter, r *http.Request) {
	var snapReq api.SnapCreateRequest
	var snapRes api.SnapCreateResponse
	method := "snap"

	if err := json.NewDecoder(r.Body).Decode(&snapReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	vd.logRequest(method, string(snapReq.Id)).Infoln("")

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	snapRes.VolumeCreateResponse = &api.VolumeCreateResponse{}

	if snapReq.Readonly {
		res, err := volumes.SnapshotCreate(ctx, &api.SdkVolumeSnapshotCreateRequest{VolumeId: snapReq.Id, Name: snapReq.Locator.Name, Labels: snapReq.Locator.VolumeLabels})
		if err != nil {
			snapRes.VolumeCreateResponse.VolumeResponse = &api.VolumeResponse{
				Error: err.Error(),
			}
		} else {
			snapRes.VolumeCreateResponse.Id = res.GetSnapshotId()
		}
	} else {
		res, err := volumes.Clone(ctx, &api.SdkVolumeCloneRequest{ParentId: snapReq.Id, Name: snapReq.Locator.Name})
		if err != nil {
			snapRes.VolumeCreateResponse.VolumeResponse = &api.VolumeResponse{
				Error: err.Error(),
			}
		} else {
			snapRes.VolumeCreateResponse.Id = res.GetVolumeId()
		}
	}
	json.NewEncoder(w).Encode(&snapRes)
}

// swagger:operation POST /osd-snapshots/restore/{id} snapshot restoreSnap
//
// Restore snapshot with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id of snapshot to restore
//   required: true
//   type: integer
// responses:
//  '200':
//    description: Restored volume
//    schema:
//     "$ref": '#/definitions/VolumeResponse'
//  default:
//   description: unexpected error
//   schema:
//    "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) restore(w http.ResponseWriter, r *http.Request) {
	var volumeID, snapID string
	var err error
	method := "restore"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	params := r.URL.Query()
	v := params[api.OptSnapID]
	if v != nil {
		snapID = v[0]
	} else {
		vd.sendError(vd.name, method, w, "Missing "+api.OptSnapID+" param",
			http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	volumeResponse := &api.VolumeResponse{}
	volumes := api.NewOpenStorageVolumeClient(conn)
	_, err = volumes.SnapshotRestore(ctx, &api.SdkVolumeSnapshotRestoreRequest{VolumeId: volumeID, SnapshotId: snapID})
	if err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation GET /osd-snapshots snapshot enumerateSnaps
//
// Enumerate snapshots.
//
// ---
// consumes:
// - multipart/form-data
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: Volume name that maps to this snap
//   required: false
//   type: string
// - name: VolumeLabels
//   in: formData
//   description: |
//    Comma separated volume labels
//    example: {"label1","label2"}
//   required: false
//   type: string
// - name: SnapLabels
//   in: formData
//   description: |
//    Comma separated snap labels
//    example: {"label1","label2"}
//   required: false
//   type: string
// - name: uuid
//   in: query
//   description: Snap UUID
//   required: false
//   type: string
//   format: uuid
// responses:
//  '200':
//   description: an array of snapshots
//   schema:
//    type: array
//    items:
//     $ref: '#/definitions/Volume'
func (vd *volAPI) snapEnumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	var labels map[string]string
	var ids []string

	method := "snapEnumerate"
	params := r.URL.Query()
	v := params[string(api.OptLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &labels); err != nil {
			e := fmt.Errorf("Failed to parse parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}

	v, ok := params[string(api.OptVolumeID)]
	if v != nil && ok {
		ids = make([]string, len(params))
		for i, s := range v {
			ids[i] = string(s)
		}
	}

	request := &api.SdkVolumeSnapshotEnumerateWithFiltersRequest{}
	if len(ids) > 0 {
		request.VolumeId = ids[0]
	}
	if len(labels) > 0 {
		request.Labels = labels
	}
	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	resp, err := volumes.SnapshotEnumerateWithFilters(ctx, request)
	if err != nil {
		e := fmt.Errorf("Failed to enumerate snaps: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	snaps := make([]*api.Volume, 0)
	for _, s := range resp.GetVolumeSnapshotIds() {
		vol, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{VolumeId: s})
		if err == nil {
			snaps = append(snaps, vol.GetVolume())
		} else if sdk.IsErrorNotFound(err) {
			continue
		} else {
			vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	json.NewEncoder(w).Encode(snaps)
}

// swagger:operation GET /osd-volumes/stats/{id} volume statsVolume
//
// Get stats for volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//  '200':
//   description: volume set response
//   schema:
//    "$ref": "#/definitions/Stats"
func (vd *volAPI) stats(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	params := r.URL.Query()
	// By default always report /proc/diskstats style stats.
	cumulative := true
	if opt, ok := params[string(api.OptCumulative)]; ok {
		if boolValue, err := strconv.ParseBool(strings.Join(opt[:], "")); !ok {
			e := fmt.Errorf("Failed to parse %s option: %s",
				api.OptCumulative, err.Error())
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		} else {
			cumulative = boolValue
		}
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	stats, err := d.Stats(volumeID, cumulative)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

/*
 * Removed until we understand why this function if failing calling the SDK
 *
func (vd *volAPI) stats(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	var method = "stats"
	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	params := r.URL.Query()
	// By default always report /proc/diskstats style stats.
	cumulative := true
	if opt, ok := params[string(api.OptCumulative)]; ok {
		if boolValue, err := strconv.ParseBool(strings.Join(opt[:], "")); !ok {
			e := fmt.Errorf("Failed to parse %s option: %s",
				api.OptCumulative, err.Error())
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		} else {
			cumulative = boolValue
		}
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	resp, err := volumes.Stats(ctx, &api.SdkVolumeStatsRequest{VolumeId: volumeID, NotCumulative: !cumulative})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(resp.GetStats())
}
*/

// swagger:operation GET /osd-volumes/usedsize/{id} volume usedSizeVolume
//
// Get Used size of volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//  '200':
//   description: volume set response
//   type: integer
//   format: int64
func (vd *volAPI) usedsize(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	used, err := d.UsedSize(volumeID)
	if err != nil {
		e := fmt.Errorf("Failed to get used size: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(used)
}

/*
 * Removed until we understand why this function if failing calling the SDK
 *
func (vd *volAPI) usedsize(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error
	var method = "usedsize"
	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	// Get context with auth token
	ctx, err := vd.annotateContext(r)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get gRPC connection
	conn, err := vd.getConn()
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)
	resp, err := volumes.CapacityUsage(ctx, &api.SdkVolumeCapacityUsageRequest{VolumeId: volumeID})

	if err != nil {
		e := fmt.Errorf("Failed to get used size: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(resp.GetCapacityUsageInfo().TotalBytes)
}
*/

// swagger:operation POST /osd-volumes/requests/{id} volume requestsVolume
//
// Get Requests for volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/ActiveRequests"
func (vd *volAPI) requests(w http.ResponseWriter, r *http.Request) {
	var err error

	method := "requests"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	requests, err := d.GetActiveRequests()
	if err != nil {
		e := fmt.Errorf("Failed to get active requests: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(requests)
}

func (vd *volAPI) volumeusage(w http.ResponseWriter, r *http.Request) {
	var err error

	method := "volumeusage"
	volumeID, err := vd.parseID(r)
	if err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	capacityInfo, err := d.CapacityUsage(volumeID)
	if err != nil || capacityInfo.Error != nil {
		var e error
		if err != nil {
			e = fmt.Errorf("Failed to get CapacityUsage: %s", err.Error())
		} else {
			e = fmt.Errorf("Failed to get CapacityUsage: %s", capacityInfo.Error.Error())
		}
		vd.sendError(vd.name, method, w, e.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(capacityInfo)
}

// swagger:operation GET /osd-volumes/quiesce/{id} volume quiesceVolume
//
// Quiesce volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) quiesce(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error
	method := "quiesce"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	params := r.URL.Query()
	timeoutStr := params[api.OptTimeoutSec]
	var timeoutSec uint64
	if timeoutStr != nil {
		var err error
		timeoutSec, err = strconv.ParseUint(timeoutStr[0], 10, 64)
		if err != nil {
			vd.sendError(vd.name, method, w, api.OptTimeoutSec+" must be int",
				http.StatusBadRequest)
			return
		}
	}

	quiesceIdParam := params[api.OptQuiesceID]
	var quiesceId string
	if len(quiesceIdParam) > 0 {
		quiesceId = quiesceIdParam[0]
	}

	volumeResponse := &api.VolumeResponse{}
	if err := d.Quiesce(volumeID, timeoutSec, quiesceId); err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation POST /osd-volumes/unquiesce/{id} volume unquiesceVolume
//
// Unquiesce volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) unquiesce(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error
	method := "unquiesce"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	volumeResponse := &api.VolumeResponse{}
	if err := d.Unquiesce(volumeID); err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation POST /osd-snapshots/groupsnap volumegroup snapVolumeGroup
//
// Take a snapshot of volumegroup
//
// ---
// produces:
// - application/json
// parameters:
// - name: groupspec
//   in: body
//   description: GroupSnap create request
//   required: true
//   schema:
//    "$ref": "#/definitions/GroupSnapCreateRequest"
// responses:
//   '200':
//     description: group snap create response
//     schema:
//      "$ref": "#/definitions/GroupSnapCreateResponse"
//   default:
//     description: unexpected error
//     schema:
//      "$ref": "#/definitions/GroupSnapCreateResponse"
func (vd *volAPI) snapGroup(w http.ResponseWriter, r *http.Request) {
	var snapReq api.GroupSnapCreateRequest
	var snapRes *api.GroupSnapCreateResponse
	method := "snapGroup"

	if err := json.NewDecoder(r.Body).Decode(&snapReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	snapRes, err = d.SnapshotGroup(snapReq.Id, snapReq.Labels, snapReq.VolumeIds, snapReq.DeleteOnFailure)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(&snapRes)
}

// swagger:operation GET /osd-volumes/versions volume listVersions
//
// Lists API versions supported by this volumeDriver.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: Supported versions
//      schema:
//         type: array
//         items:
//            type: string
func (vd *volAPI) versions(w http.ResponseWriter, r *http.Request) {
	versions := []string{
		volume.APIVersion,
		// Update supported versions by adding them here
	}
	json.NewEncoder(w).Encode(versions)
}

// swagger:operation GET /osd-volumes/catalog/{id} volume catalogVolume
//
// Catalog lists the files and folders on volume with specified id.
// Path is optional and default the behaviour is a catalog on the root of the volume.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// - name: subfolder
//   in: query
//   description: Optional path inside mount to catalog.
//   required: false
//   type: string
// - name: depth
//   in: query
//   description: Folder depth we wish to return, default is all.
//   required: false
//   type: string
// responses:
//   '200':
//     description: volume catalog response
//     schema:
//       $ref: '#/definitions/CatalogResponse'
func (vd *volAPI) catalog(w http.ResponseWriter, r *http.Request) {
	var err error
	var volumeID string
	var subfolder string
	var depth = "0"

	method := "catalog"
	d, err := vd.getVolDriver(r)
	if err != nil {
		fmt.Println("Volume not found")
		notFound(w, r)
		return
	}

	if volumeID, err = vd.parseParam(r, "id"); err != nil {
		e := fmt.Errorf("Failed to parse ID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	params := r.URL.Query()
	folderParam := params[string(api.OptCatalogSubFolder)]
	if len(folderParam) > 0 {
		subfolder = folderParam[0]
	}

	depthParam := params[string(api.OptCatalogMaxDepth)]
	if len(depthParam) > 0 {
		depth = depthParam[0]
	}

	dk, err := d.Catalog(volumeID, subfolder, depth)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(dk)
}

// swagger:operation POST /osd-volumes/volservice/{id} volume VolumeService
//
// Does Volume Service operation in the background on a given volume
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
//   type: integer
// - name: VolumeServiceRequest
//   in: body
//   description: Contains the volume service command and parameters for the command
//   required: true
//   schema:
//         "$ref": "#/definitions/VolumeServiceRequest"
// responses:
//   '200':
//     description: volume service response
//     schema:
//       $ref: '#/definitions/VolumeServiceResponse'
//
func (vd *volAPI) VolService(w http.ResponseWriter, r *http.Request) {
	var (
		volumeID string
		err      error
		vsreq    api.VolumeServiceRequest
	)
	method := "Srv:VolService"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&vsreq)
	if err != nil {
		e := fmt.Errorf("Failed to Decode api.VolumeServiceRequest from HTTP request body: Err: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	vsresp, err := d.VolService(volumeID, &vsreq)
	if err != nil {
		e := fmt.Errorf("%s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(*vsresp)
}

func volVersion(route, version string) string {
	if version == "" {
		return "/" + route
	} else {
		return "/" + version + "/" + route
	}
}

func volPath(route, version string) string {
	return volVersion(api.OsdVolumePath+route, version)
}

func snapPath(route, version string) string {
	return volVersion(api.OsdSnapshotPath+route, version)
}

func credsPath(route, version string) string {
	return volVersion(api.OsdCredsPath+route, version)
}

func backupPath(route, version string) string {
	return volVersion(api.OsdBackupPath+route, version)
}

func migratePath(route, version string) string {
	return volVersion(route, version)
}

func (vd *volAPI) versionRoute() *Route {
	return &Route{verb: "GET", path: "/" + api.OsdVolumePath + "/versions", fn: vd.versions}

}
func (vd *volAPI) volumeCreateRoute() *Route {
	return &Route{verb: "POST", path: volPath("", volume.APIVersion), fn: vd.create}
}

func (vd *volAPI) volumeDeleteRoute() *Route {
	return &Route{verb: "DELETE", path: volPath("/{id}", volume.APIVersion), fn: vd.delete}
}

func (vd *volAPI) volumeSetRoute() *Route {
	return &Route{verb: "PUT", path: volPath("/{id}", volume.APIVersion), fn: vd.volumeSet}
}

func (vd *volAPI) volumeInspectRoute() *Route {
	return &Route{verb: "GET", path: volPath("/{id}", volume.APIVersion), fn: vd.inspect}
}

func (vd *volAPI) volumeEnumerateRoute() *Route {
	return &Route{verb: "GET", path: volPath("", volume.APIVersion), fn: vd.enumerate}
}

func (vd *volAPI) otherVolumeRoutes() []*Route {
	return []*Route{
		{verb: "GET", path: volPath("/stats", volume.APIVersion), fn: vd.stats},
		{verb: "GET", path: volPath("/stats/{id}", volume.APIVersion), fn: vd.stats},
		{verb: "GET", path: volPath("/usedsize", volume.APIVersion), fn: vd.usedsize},
		{verb: "GET", path: volPath("/usedsize/{id}", volume.APIVersion), fn: vd.usedsize},
		{verb: "GET", path: volPath("/requests", volume.APIVersion), fn: vd.requests},
		{verb: "GET", path: volPath("/requests/{id}", volume.APIVersion), fn: vd.requests},
		{verb: "GET", path: volPath("/usage", volume.APIVersion), fn: vd.volumeusage},
		{verb: "GET", path: volPath("/usage/{id}", volume.APIVersion), fn: vd.volumeusage},
		{verb: "POST", path: volPath("/quiesce/{id}", volume.APIVersion), fn: vd.quiesce},
		{verb: "POST", path: volPath("/unquiesce/{id}", volume.APIVersion), fn: vd.unquiesce},
		{verb: "GET", path: volPath("/catalog/{id}", volume.APIVersion), fn: vd.catalog},
		{verb: "POST", path: volPath("/volservice/{id}", volume.APIVersion), fn: vd.VolService},
	}
}

func (vd *volAPI) backupRoutes() []*Route {
	return []*Route{
		{verb: "POST", path: backupPath("", volume.APIVersion), fn: vd.cloudBackupCreate},
		{verb: "POST", path: backupPath("/group", volume.APIVersion), fn: vd.cloudBackupGroupCreate},
		{verb: "POST", path: backupPath("/restore", volume.APIVersion), fn: vd.cloudBackupRestore},
		{verb: "GET", path: backupPath("", volume.APIVersion), fn: vd.cloudBackupEnumerate},
		{verb: "DELETE", path: backupPath("", volume.APIVersion), fn: vd.cloudBackupDelete},
		{verb: "DELETE", path: backupPath("/all", volume.APIVersion), fn: vd.cloudBackupDeleteAll},
		{verb: "GET", path: backupPath("/status", volume.APIVersion), fn: vd.cloudBackupStatus},
		{verb: "GET", path: backupPath("/catalog", volume.APIVersion), fn: vd.cloudBackupCatalog},
		{verb: "GET", path: backupPath("/history", volume.APIVersion), fn: vd.cloudBackupHistory},
		{verb: "PUT", path: backupPath("/statechange", volume.APIVersion), fn: vd.cloudBackupStateChange},
		{verb: "POST", path: backupPath("/sched", volume.APIVersion), fn: vd.cloudBackupSchedCreate},
		{verb: "PUT", path: backupPath("/sched", volume.APIVersion), fn: vd.cloudBackupSchedUpdate},
		{verb: "POST", path: backupPath("/schedgroup", volume.APIVersion), fn: vd.cloudBackupGroupSchedCreate},
		{verb: "PUT", path: backupPath("/schedgroup", volume.APIVersion), fn: vd.cloudBackupGroupSchedUpdate},
		{verb: "DELETE", path: backupPath("/sched", volume.APIVersion), fn: vd.cloudBackupSchedDelete},
		{verb: "GET", path: backupPath("/sched", volume.APIVersion), fn: vd.cloudBackupSchedEnumerate},
	}
}

func (vd *volAPI) credsRoutes() []*Route {
	return []*Route{
		{verb: "GET", path: credsPath("", volume.APIVersion), fn: vd.credsEnumerate},
		{verb: "POST", path: credsPath("", volume.APIVersion), fn: vd.credsCreate},
		{verb: "DELETE", path: credsPath("/{uuid}", volume.APIVersion), fn: vd.credsDelete},
		{verb: "PUT", path: credsPath("/validate/{uuid}", volume.APIVersion), fn: vd.credsValidate},
	}
}

func (vd *volAPI) migrateRoutes() []*Route {
	return []*Route{
		{verb: "POST", path: migratePath(api.OsdMigrateStartPath, volume.APIVersion), fn: vd.cloudMigrateStart},
		{verb: "POST", path: migratePath(api.OsdMigrateCancelPath, volume.APIVersion), fn: vd.cloudMigrateCancel},
		{verb: "GET", path: migratePath(api.OsdMigrateStatusPath, volume.APIVersion), fn: vd.cloudMigrateStatus},
	}
}

func (vd *volAPI) snapRoutes() []*Route {
	return []*Route{
		{verb: "POST", path: snapPath("", volume.APIVersion), fn: vd.snap},
		{verb: "GET", path: snapPath("", volume.APIVersion), fn: vd.snapEnumerate},
		{verb: "POST", path: snapPath("/restore/{id}", volume.APIVersion), fn: vd.restore},
		{verb: "POST", path: snapPath("/snapshotgroup", volume.APIVersion), fn: vd.snapGroup},
	}
}

func (vd *volAPI) Routes() []*Route {
	routes := []*Route{vd.versionRoute(), vd.volumeCreateRoute(), vd.volumeSetRoute(), vd.volumeDeleteRoute(), vd.volumeInspectRoute(), vd.volumeEnumerateRoute()}
	routes = append(routes, vd.otherVolumeRoutes()...)
	routes = append(routes, vd.snapRoutes()...)
	routes = append(routes, vd.backupRoutes()...)
	routes = append(routes, vd.credsRoutes()...)
	routes = append(routes, vd.migrateRoutes()...)
	return routes
}

func (vd *volAPI) SetupRoutesWithAuth(
	router *mux.Router,
	authenticators map[string]auth.Authenticator,
) (*mux.Router, error) {
	// We setup auth middlewares for all the APIs that get invoked
	// from a Container Orchestrator.
	// - CREATE
	// - ATTACH/MOUNT
	// - DETACH/UNMOUNT
	// - DELETE
	// - ENUMERATE
	// For all other routes it is expected that the REST client uses an auth token

	authM := NewAuthMiddleware()

	// Setup middleware for Create
	nCreate := negroni.New()
	nCreate.Use(negroni.HandlerFunc(authM.createWithAuth))
	createRoute := vd.volumeCreateRoute()
	nCreate.UseHandlerFunc(createRoute.fn)
	router.Methods(createRoute.verb).Path(createRoute.path).Handler(nCreate)

	// Setup middleware for Delete
	nDelete := negroni.New()
	nDelete.Use(negroni.HandlerFunc(authM.deleteWithAuth))
	deleteRoute := vd.volumeDeleteRoute()
	nDelete.UseHandlerFunc(deleteRoute.fn)
	router.Methods(deleteRoute.verb).Path(deleteRoute.path).Handler(nDelete)

	// Setup middleware for Set
	nSet := negroni.New()
	nSet.Use(negroni.HandlerFunc(authM.setWithAuth))
	setRoute := vd.volumeSetRoute()
	nSet.UseHandlerFunc(setRoute.fn)
	router.Methods(setRoute.verb).Path(setRoute.path).Handler(nSet)

	// Setup middleware for Inspect
	nInspect := negroni.New()
	nInspect.Use(negroni.HandlerFunc(authM.inspectWithAuth))
	inspectRoute := vd.volumeInspectRoute()
	nSet.UseHandlerFunc(inspectRoute.fn)
	router.Methods(inspectRoute.verb).Path(inspectRoute.path).Handler(nInspect)

	// Setup middleware for enumerate
	nEnumerate := negroni.New()
	nEnumerate.Use(negroni.HandlerFunc(authM.enumerateWithAuth))
	enumerateRoute := vd.volumeEnumerateRoute()
	nSet.UseHandlerFunc(enumerateRoute.fn)
	router.Methods(enumerateRoute.verb).Path(enumerateRoute.path).Handler(nEnumerate)

	routes := []*Route{vd.versionRoute()}
	routes = append(routes, vd.otherVolumeRoutes()...)
	routes = append(routes, vd.snapRoutes()...)
	routes = append(routes, vd.backupRoutes()...)
	routes = append(routes, vd.migrateRoutes()...)
	for _, v := range routes {
		router.Methods(v.verb).Path(v.path).HandlerFunc(v.fn)
	}

	return router, nil
}

func (vd *volAPI) decorateCredRoutes(authenticators map[string]auth.Authenticator, router *mux.Router) *mux.Router {
	// Decorate creds endpoints with authentication
	credRoutes := vd.credsRoutes()
	securityMiddleware := newSecurityMiddleware(authenticators)
	for _, route := range credRoutes {
		router.Methods(route.GetVerb()).Path(route.GetPath()).HandlerFunc(securityMiddleware(route.fn))
	}

	return router
}

// GetVolumeAPIRoutes returns all the volume routes.
// A driver could use this function if it does not want openstorage
// to setup the REST server but it sets up its own and wants to add
// volume routes
func GetVolumeAPIRoutes(name, sdkUds string) []*Route {
	volMgmtApi := newVolumeAPI(name, sdkUds)
	return volMgmtApi.Routes()
}

func GetCredAPIRoutes() []*Route {
	volAPI := &volAPI{}
	return volAPI.credsRoutes()
}

func SetCredsAPIRoutesWithAuth(router *mux.Router, authenticators map[string]auth.Authenticator) *mux.Router {
	volAPI := &volAPI{}
	return volAPI.decorateCredRoutes(authenticators, router)
}

// ServerRegisterRoute is a callback function used by drivers to run their
// preRouteChecks before the actual volume route gets invoked
// This is added for legacy support before negroni middleware was added
type ServerRegisterRoute func(
	routeFunc func(w http.ResponseWriter, r *http.Request),
	preRouteCheck func(w http.ResponseWriter, r *http.Request) bool,
) func(w http.ResponseWriter, r *http.Request)

// GetVolumeAPIRoutesWithAuth returns a router with all the volume routes
// added to the router along with the auth middleware
// - preRouteCheckFn is a handler that gets executed before the actual volume handler
// is invoked. It is added for legacy support where negroni middleware was not used
func GetVolumeAPIRoutesWithAuth(
	name, sdkUds string,
	router *mux.Router,
	serverRegisterRoute ServerRegisterRoute,
	preRouteCheckFn func(http.ResponseWriter, *http.Request) bool,
) (*mux.Router, error) {
	vd := &volAPI{
		restBase: restBase{version: volume.APIVersion, name: name},
		sdkUds:   sdkUds,
		dummyMux: runtime.NewServeMux(),
	}

	authM := NewAuthMiddleware()

	// We setup auth middlewares for all the APIs that get invoked
	// from a Container Orchestrator.
	// - CREATE
	// - ATTACH/MOUNT
	// - DETACH/UNMOUNT
	// - DELETE
	// For all other routes it is expected that the REST client uses an auth token

	// Setup middleware for Create
	nCreate := negroni.New()
	nCreate.Use(negroni.HandlerFunc(authM.createWithAuth))
	createRoute := vd.volumeCreateRoute()
	nCreate.UseHandlerFunc(serverRegisterRoute(createRoute.fn, preRouteCheckFn))
	router.Methods(createRoute.verb).Path(createRoute.path).Handler(nCreate)

	// Setup middleware for Delete
	nDelete := negroni.New()
	nDelete.Use(negroni.HandlerFunc(authM.deleteWithAuth))
	deleteRoute := vd.volumeDeleteRoute()
	nDelete.UseHandlerFunc(serverRegisterRoute(deleteRoute.fn, preRouteCheckFn))
	router.Methods(deleteRoute.verb).Path(deleteRoute.path).Handler(nDelete)

	// Setup middleware for Set
	nSet := negroni.New()
	nSet.Use(negroni.HandlerFunc(authM.setWithAuth))
	setRoute := vd.volumeSetRoute()
	nSet.UseHandlerFunc(serverRegisterRoute(setRoute.fn, preRouteCheckFn))
	router.Methods(setRoute.verb).Path(setRoute.path).Handler(nSet)

	// Setup middleware for Inspect
	nInspect := negroni.New()
	nInspect.Use(negroni.HandlerFunc(authM.inspectWithAuth))
	inspectRoute := vd.volumeInspectRoute()
	nInspect.UseHandlerFunc(serverRegisterRoute(inspectRoute.fn, preRouteCheckFn))
	router.Methods(inspectRoute.verb).Path(inspectRoute.path).Handler(nInspect)

	// Setup middleware for Enumerate
	nEnumerate := negroni.New()
	nEnumerate.Use(negroni.HandlerFunc(authM.enumerateWithAuth))
	enumerateRoute := vd.volumeEnumerateRoute()
	nEnumerate.UseHandlerFunc(serverRegisterRoute(enumerateRoute.fn, preRouteCheckFn))
	router.Methods(enumerateRoute.verb).Path(enumerateRoute.path).Handler(nEnumerate)

	routes := []*Route{vd.versionRoute()}
	routes = append(routes, vd.otherVolumeRoutes()...)
	routes = append(routes, vd.snapRoutes()...)
	routes = append(routes, vd.backupRoutes()...)
	routes = append(routes, vd.migrateRoutes()...)
	for _, v := range routes {
		router.Methods(v.verb).Path(v.path).HandlerFunc(serverRegisterRoute(v.fn, preRouteCheckFn))
	}
	return router, nil

}
