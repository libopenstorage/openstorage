package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/config"
	osecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/libopenstorage/openstorage/pkg/options"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// VolumeDriver is the string returned in the handshake protocol.
	VolumeDriver = "VolumeDriver"
)

// Implementation of the Docker volumes plugin specification.
type driver struct {
	restBase
	spec.SpecHandler

	sdkUds string
	conn   *grpc.ClientConn
	mu     sync.Mutex

	secretsStore *osecrets.Auth
}

type handshakeResp struct {
	Implements []string
}

type volumeRequest struct {
	Name string
	Opts map[string]string
}

type mountRequest struct {
	Name string
	ID   string
}

type volumeResponse struct {
	Err string
}

type volumePathResponse struct {
	Mountpoint string
	volumeResponse
}

type volumeInfo struct {
	Name       string
	Mountpoint string
}

type capabilities struct {
	Scope string
}

type capabilitiesResponse struct {
	Capabilities capabilities
}

func newVolumePlugin(name, sdkUds string, secretsStore *osecrets.Auth) restServer {
	d := &driver{
		restBase:     restBase{name: name, version: "0.3"},
		SpecHandler:  spec.NewSpecHandler(),
		sdkUds:       sdkUds,
		secretsStore: secretsStore,
	}
	return d
}

func (d *driver) String() string {
	return d.name
}

func volDriverPath(method string) string {
	return fmt.Sprintf("/%s.%s", VolumeDriver, method)
}

func (d *driver) volNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume: " + e.Error())
	if e == volume.ErrDriverInitializing {
		d.logRequest(request, id).Warnln(http.StatusInternalServerError, " ", err.Error())
	} else {
		d.logRequest(request, id).Warnln(http.StatusNotFound, " ", err.Error())
	}
	return err
}

func (d *driver) volNotMounted(request string, id string) error {
	err := fmt.Errorf("volume not mounted")
	d.logRequest(request, id).Debugln(http.StatusNotFound, " ", err.Error())
	return err
}

func (d *driver) Routes() []*Route {
	return []*Route{
		// secure handlers
		{verb: "POST", path: volDriverPath("Create"), fn: d.create},
		{verb: "POST", path: volDriverPath("Remove"), fn: d.remove},
		{verb: "POST", path: volDriverPath("Mount"), fn: d.mount},

		// insecure handlers
		{verb: "POST", path: volDriverPath("Path"), fn: d.path},
		{verb: "POST", path: volDriverPath("Get"), fn: d.get},
		{verb: "POST", path: volDriverPath("List"), fn: d.list},
		{verb: "POST", path: volDriverPath("Unmount"), fn: d.unmount},
		{verb: "POST", path: volDriverPath("Capabilities"), fn: d.capabilities},
		{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
		{verb: "GET", path: "/status", fn: d.status},
	}
}

func (d *driver) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) errorResponse(method string, w http.ResponseWriter, err error) {
	if err == volume.ErrDriverInitializing {
		d.sendError(method, "", w, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(&volumeResponse{Err: err.Error()})
	}
}

func (d *driver) volFromName(name string) (*api.Volume, error) {
	v, err := volumedrivers.Get(d.name)
	if err != nil {
		return nil, fmt.Errorf("Cannot locate volume driver for %s: %s", d.name, err.Error())
	}
	return util.VolumeFromName(v, name)
}

func (d *driver) volFromNameSdk(ctx context.Context, volumes api.OpenStorageVolumeClient, name string) (*api.Volume, error) {
	// get volume id
	volId, err := d.volIdFromName(ctx, volumes, name)
	if err != nil {
		return nil, err
	}

	// inspect for actual volume
	inspectResp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: volId,
	})
	if err != nil {
		return nil, err
	}
	return inspectResp.Volume, nil
}

func (d *driver) volFromIdSdk(ctx context.Context, volumes api.OpenStorageVolumeClient, volId string) (*api.Volume, error) {
	inspectResp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: volId,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot locate volume with id %s", volId)
	}
	return inspectResp.Volume, nil
}

func (d *driver) volIdFromName(ctx context.Context, volumes api.OpenStorageVolumeClient, name string) (string, error) {
	enumerateResp, err := volumes.EnumerateWithFilters(ctx, &api.SdkVolumeEnumerateWithFiltersRequest{
		Name: name,
	})
	if err != nil {
		return "", err
	} else if len(enumerateResp.VolumeIds) < 1 {
		return "", fmt.Errorf("Cannot locate volume with name %s", name)
	}

	return enumerateResp.VolumeIds[0], nil
}

func (d *driver) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeRequest, error) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	d.logRequest(method, request.Name).Debugln("")
	return &request, nil
}

func (d *driver) decodeMount(method string, w http.ResponseWriter, r *http.Request) (*mountRequest, error) {
	var request mountRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	return &request, nil
}

func (d *driver) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{VolumeDriver},
	})
	if err != nil {
		d.sendError("handshake", "", w, "encode error", http.StatusInternalServerError)
		return
	}
	d.logRequest("handshake", "").Debugln("Handshake completed")
}

// attachTokenMount adds the user's token to the context auth metadata
func (d *driver) attachToken(ctx context.Context, request *volumeRequest) (context.Context, string, error) {
	token, err := d.parseTokenInput(request.Name, request.Opts)
	if err != nil {
		return ctx, "", err
	}

	return addTokenMetadata(ctx, token), "", nil
}

// attachTokenMount adds the user's token to the context auth metadata
func (d *driver) attachTokenMount(ctx context.Context, request *mountRequest) (context.Context, string, error) {
	token, err := d.parseTokenInput(request.Name, make(map[string]string))
	if err != nil {
		return ctx, "", err
	}

	return addTokenMetadata(ctx, token), token, nil
}

// parseTokenInput reads token input from the given name and opts.
// The following is the order of precedence for token in types:
//   1. token=<token> in name
//   2. token in opts
//   3. token_secret=<secret> in name
//   4. token_secret in opts
func (d *driver) parseTokenInput(name string, opts map[string]string) (string, error) {
	// get token from name
	tokenFromName, tokenInName := d.GetTokenFromString(name)
	if tokenInName {
		return tokenFromName, nil
	}

	// get token from opts
	tokenFromOpts := opts[api.Token]
	if tokenFromOpts != "" {
		return tokenFromOpts, nil
	}

	// get token secret
	tokenSecretContext, ok := d.GetTokenSecretContextFromString(name)
	if ok {
		token, err := d.secretTokenFromStore(tokenSecretContext)
		if err != nil {
			return "", err
		}

		return token, nil
	}

	// get token secret from opts
	tokenSecretFromOpts := opts[api.TokenSecret]
	if tokenSecretFromOpts != "" {
		return tokenSecretFromOpts, nil
	}

	return "", nil
}

// addTokenMetadata adds the token to a given context's metadata
func addTokenMetadata(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{
		"authorization": "bearer " + token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

// secretTokenFromStore pulls the token from the configured secret store for
// a given secret name and context.
func (d *driver) secretTokenFromStore(req *api.TokenSecretContext) (string, error) {
	if d.secretsStore == nil {
		return "", fmt.Errorf("A secret was passed in, but no secrets provider has been initialized")
	}

	token, err := d.secretsStore.GetToken(req)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *driver) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintln("osd plugin", d.version))
}

func (d *driver) mountpath(name string) string {
	return path.Join(volume.MountBase, name)
}

func (d *driver) getConn() (*grpc.ClientConn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn == nil {
		var err error
		d.conn, err = grpcserver.Connect(
			d.sdkUds,
			[]grpc.DialOption{grpc.WithInsecure()})
		if err != nil {
			return nil, fmt.Errorf("Failed to connect to gRPC handler: %v", err)
		}
	}
	return d.conn, nil
}

func (d *driver) create(w http.ResponseWriter, r *http.Request) {
	method := "create"
	ctx := r.Context()
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	// attach token in context metadata
	ctx, _, err = d.attachToken(ctx, request)
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	// get spec for volume creation
	specParsed, spec, locator, source, name := d.SpecFromString(request.Name)
	d.logRequest(method, name).Infoln("")
	if !specParsed {
		spec, locator, source, err = d.SpecFromOpts(request.Opts)
		if err != nil {
			d.errorResponse(method, w, err)
			return
		}
	}

	// get grpc connection
	conn, err := d.getConn()
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	// clone if exists, create otherwise
	volumes := api.NewOpenStorageVolumeClient(conn)
	if source != nil && len(source.Parent) != 0 {
		// clone
		_, err = volumes.Clone(ctx, &api.SdkVolumeCloneRequest{
			Name:     name,
			ParentId: source.Parent,
		})
	} else {
		// create
		_, err = volumes.Create(ctx, &api.SdkVolumeCreateRequest{
			Name:   name,
			Spec:   spec,
			Labels: locator.GetVolumeLabels(),
		})
	}
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) remove(w http.ResponseWriter, r *http.Request) {
	method := "remove"
	ctx := r.Context()
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	// attach token in context metadata
	ctx, _, err = d.attachToken(ctx, request)
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	// get name for deletion
	_, _, _, _, name := d.SpecFromString(request.Name)
	d.logRequest(method, name).Infoln("")

	// get grpc connection
	conn, err := d.getConn()
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}
	volumes := api.NewOpenStorageVolumeClient(conn)

	// get volume id to delete
	volId, err := d.volIdFromName(ctx, volumes, name)
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	// delete volume
	_, err = volumes.Delete(ctx, &api.SdkVolumeDeleteRequest{
		VolumeId: volId,
	})
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) scaleUp(
	ctx context.Context,
	conn *grpc.ClientConn,
	method string,
	vd volume.VolumeDriver,
	inVol *api.Volume,
	allVols []*api.Volume,
	attachOptions map[string]string,
) (
	outVol *api.Volume,
	err error,
) {
	// Get an SDK volume client
	volumeclient := api.NewOpenStorageVolumeClient(conn)
	mountClient := api.NewOpenStorageMountAttachClient(conn)

	// Create new volume if existing volumes are not available.
	spec := inVol.Spec.Copy()
	spec.Scale = 1
	spec.ReplicaSet = nil
	volCount := len(allVols)
	for i := len(allVols); volCount < int(inVol.Spec.Scale); i++ {
		name := fmt.Sprintf("%s_%03d", inVol.Locator.Name, i)
		resp, err := volumeclient.Create(ctx, &api.SdkVolumeCreateRequest{
			Name: name,
			Spec: spec,
		})
		if err != nil {
			return nil, err
		}
		id := resp.GetVolumeId()
		if outVol, err = d.volFromName(id); err != nil {
			return nil, err
		}
		_, err = mountClient.Attach(ctx, &api.SdkVolumeAttachRequest{
			VolumeId:      outVol.Id,
			DriverOptions: attachOptions,
		})
		if err == nil {
			return outVol, nil
		}
		// If we fail to attach the volume, continue to look for a
		// free volume.
		volCount++
	}
	return nil, volume.ErrVolAttachedScale
}

func (d *driver) attachScale(
	ctx context.Context,
	conn *grpc.ClientConn,
	method string,
	vd volume.VolumeDriver,
	inVol *api.Volume,
	attachOptions map[string]string,
) (
	*api.Volume,
	error,
) {
	// Find a volume that has data local to this node.
	vols, err := vd.Enumerate(
		&api.VolumeLocator{
			Name: fmt.Sprintf("%s.*", inVol.Locator.Name),
			VolumeLabels: map[string]string{
				volume.LocationConstraint: volume.LocalNode,
			},
		},
		nil,
	)
	// Try to attach local volumes.
	if err == nil {
		for _, vol := range vols {
			if v, err := d.attachVol(ctx, conn, method, vd, vol, attachOptions); err == nil {
				return v, nil
			}
		}
	}
	// Create a new local volume if we fail to attach existing local volume
	// or if none exist.
	allVols, err := vd.Enumerate(
		&api.VolumeLocator{
			Name: fmt.Sprintf("%s.*", inVol.Locator.Name),
		},
		nil,
	)

	// Get an SDK mount/attach client
	mountClient := api.NewOpenStorageMountAttachClient(conn)

	// Try to attach existing volumes.
	for _, outVol := range allVols {
		_, err := mountClient.Attach(ctx, &api.SdkVolumeAttachRequest{
			VolumeId:      outVol.Id,
			DriverOptions: attachOptions,
		})
		if err == nil {
			return outVol, nil
		}
	}

	// Get an SDK volume client
	volumeclient := api.NewOpenStorageVolumeClient(conn)

	if len(allVols) < int(inVol.Spec.Scale) {
		name := fmt.Sprintf("%s_%03d", inVol.Locator.Name, len(allVols))
		spec := inVol.Spec.Copy()
		spec.ReplicaSet = &api.ReplicaSet{Nodes: []string{volume.LocalNode}}
		spec.Scale = 1
		resp, err := volumeclient.Create(ctx, &api.SdkVolumeCreateRequest{
			Name: name,
			Spec: spec,
		})
		if err != nil {
			return d.scaleUp(ctx, conn, method, vd, inVol, allVols, attachOptions)
		}
		id := resp.GetVolumeId()
		outVol, err := d.volFromName(id)
		if err != nil {
			return nil, err
		}
		_, err = mountClient.Attach(ctx, &api.SdkVolumeAttachRequest{
			VolumeId:      outVol.Id,
			DriverOptions: attachOptions,
		})
		if err == nil {
			return outVol, nil
		}
		// We failed to attach, scaleUp.
		allVols = append(allVols, outVol)
	}
	return d.scaleUp(ctx, conn, method, vd, inVol, allVols, attachOptions)
}

func (d *driver) attachVol(
	ctx context.Context,
	conn *grpc.ClientConn,
	method string,
	vd volume.VolumeDriver,
	vol *api.Volume,
	attachOptions map[string]string,
) (
	outVolume *api.Volume,
	err error,
) {
	mountClient := api.NewOpenStorageMountAttachClient(conn)
	resp, err := mountClient.Attach(ctx, &api.SdkVolumeAttachRequest{
		VolumeId:      vol.Id,
		DriverOptions: attachOptions,
	})
	attachPath := resp.GetDevicePath()

	if serverError, ok := status.FromError(err); ok {
		switch serverError.Code() {
		case codes.OK:
			d.logRequest(method, vol.Locator.Name).Debugf(
				"response %v", attachPath)
			return vol, nil
		case codes.AlreadyExists:
			d.logRequest(method, vol.Locator.Name).Infof(
				"Mount volume attached on remote node.")
			return vol, err
		default:
			d.logRequest(method, vol.Locator.Name).Warnf(
				"Cannot attach volume: %v", err.Error())
			return vol, err
		}
	} else {
		d.logRequest(method, vol.Locator.Name).Warnf(
			"Cannot attach volume: %v", err.Error())
		return vol, err
	}
}

func (d *driver) attachOptionsFromSpec(
	spec *api.VolumeSpec,
) map[string]string {
	if spec.Passphrase != "" {
		opts := make(map[string]string)
		opts[options.OptionsSecret] = spec.Passphrase
		return opts
	}
	return nil
}

func (d *driver) mount(w http.ResponseWriter, r *http.Request) {
	var response volumePathResponse
	ctx := r.Context()
	method := "mount"
	request, err := d.decodeMount(method, w, r)
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	// get spec and name from request
	_, spec, _, _, name := d.SpecFromString(request.Name)
	attachOptions := d.attachOptionsFromSpec(spec)

	// attach token in context metadata
	ctx, _, err = d.attachTokenMount(ctx, request)
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	// get grpc connection
	conn, err := d.getConn()
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	//get volume to mount
	mountClient := api.NewOpenStorageMountAttachClient(conn)
	volumeClient := api.NewOpenStorageVolumeClient(conn)
	vol, err := d.volFromNameSdk(ctx, volumeClient, name)
	if err != nil {
		d.sendError(method, "", w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get access to local driver
	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf("Cannot locate volume driver")
		d.errorResponse(method, w, err)
		return
	}

	// If a scaled volume is already mounted, check if it can be unmounted and
	// detached. If not return an error.
	mountpoint := d.mountpath(name)
	if vol.Spec.Scale > 1 {
		id := v.MountedAt(mountpoint)
		if len(id) != 0 {
			err = v.Unmount(id, mountpoint, nil)
			if err != nil {
				d.logRequest(method, "").Warnf("Error unmounting scaled volume: %v", err)
				err = fmt.Errorf("Cannot remount scaled volume(%v)."+
					" Volume %v is mounted at %v", name, id, mountpoint)
				d.errorResponse(method, w, err)
				return
			}

			if v.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
				err = v.Detach(id, nil)
				if err != nil {
					d.logRequest(method, "").Warnf("Error detaching scaled volume: %v", err)

					_, mountErr := mountClient.Mount(ctx, &api.SdkVolumeMountRequest{
						VolumeId:  id,
						MountPath: mountpoint,
					})
					if mountErr != nil {
						d.logRequest(method, "").Warnf("Error remounting scaled volume: %v", mountErr.Error())
					}
					err = fmt.Errorf("Cannot remount scaled volume(%v)."+
						" Volume %v is mounted at %v", name, id, mountpoint)
					d.logRequest(method, "").Warnf(err.Error())
					d.errorResponse(method, w, err)
					return
				}
			}
		}
	}

	// If this is a block driver, first attach the volume.
	if v.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		// If volume is scaled up, a new volume is created and
		// vol will change.
		if vol.Scaled() {
			vol, err = d.attachScale(ctx, conn, method, v, vol, attachOptions)
		} else {
			vol, err = d.attachVol(ctx, conn, method, v, vol, attachOptions)
		}
		if err != nil {
			d.errorResponse(method, w, err)
			return
		}
	}

	// get and prepare mountpath
	response.Mountpoint = mountpoint
	os.MkdirAll(mountpoint, 0755)

	// mount volume
	_, err = mountClient.Mount(ctx, &api.SdkVolumeMountRequest{
		VolumeId:  vol.Id,
		MountPath: mountpoint,
		Options:   options.NewVolumeAttachOptions(attachOptions),
	})
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	json.NewEncoder(w).Encode(&response)
}

func (d *driver) list(w http.ResponseWriter, r *http.Request) {
	method := "list"

	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf("Cannot locate volume driver: %v", err.Error())
		d.errorResponse(method, w, err)
		return
	}

	vols, err := v.Enumerate(nil, nil)
	if err != nil {
		d.errorResponse(method, w, err)
		return
	}

	volInfo := make([]volumeInfo, len(vols))
	for i, v := range vols {
		volInfo[i].Name = v.Locator.Name
		if len(v.AttachPath) > 0 || len(v.AttachPath) > 0 {
			volInfo[i].Mountpoint = path.Join(v.AttachPath[0], config.DataDir)
		}
	}
	json.NewEncoder(w).Encode(map[string][]volumeInfo{"Volumes": volInfo})
}

func (d *driver) path(w http.ResponseWriter, r *http.Request) {
	method := "path"
	var response volumePathResponse

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	_, _, _, _, name := d.SpecFromString(request.Name)
	vol, err := d.volFromName(name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(method, w, e)
		return
	}

	d.logRequest(method, name).Debugf("")
	if len(vol.AttachPath) == 0 || len(vol.AttachPath) == 0 {
		e := d.volNotMounted(method, name)
		d.errorResponse(method, w, e)
		return
	}
	response.Mountpoint = vol.AttachPath[0]
	response.Mountpoint = path.Join(response.Mountpoint, config.DataDir)
	d.logRequest(method, request.Name).Debugf("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) get(w http.ResponseWriter, r *http.Request) {
	method := "get"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	parsed, _, _, _, name := d.SpecFromString(request.Name)
	returnName := ""
	if parsed {
		returnName = request.Name
	} else {
		returnName = name
	}
	vol, err := d.volFromName(name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(method, w, e)
		return
	}

	volInfo := volumeInfo{Name: returnName}
	if len(vol.AttachPath) > 0 || len(vol.AttachPath) > 0 {
		volInfo.Mountpoint = path.Join(vol.AttachPath[0], config.DataDir)
	}

	json.NewEncoder(w).Encode(map[string]volumeInfo{"Volume": volInfo})
}

func (d *driver) unmount(w http.ResponseWriter, r *http.Request) {
	method := "unmount"

	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf(
			"Cannot locate volume driver: %v",
			err.Error())
		d.errorResponse(method, w, err)
		return
	}

	request, err := d.decodeMount(method, w, r)
	if err != nil {
		return
	}

	_, _, _, _, name := d.SpecFromString(request.Name)
	vol, err := d.volFromName(name)
	if err != nil {
		e := d.volNotFound(method, name, err, w)
		d.errorResponse(method, w, e)
		return
	}

	mountpoint := d.mountpath(name)
	id := vol.Id
	if vol.Spec.Scale > 1 {
		id = v.MountedAt(mountpoint)
		if len(id) == 0 {
			err := fmt.Errorf("Failed to find volume mapping for %v",
				mountpoint)
			d.logRequest(method, request.Name).Warnf(
				"Cannot unmount volume %v, %v",
				mountpoint, err)
			d.errorResponse(method, w, err)
			return
		}
	}

	opts := make(map[string]string)
	opts[options.OptionsDeleteAfterUnmount] = "true"

	err = v.Unmount(id, mountpoint, opts)
	if err != nil {
		d.logRequest(method, request.Name).Warnf(
			"Cannot unmount volume %v, %v",
			mountpoint, err)
		d.errorResponse(method, w, err)
		return
	}

	if v.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		_ = v.Detach(id, nil)
	}
	d.emptyResponse(w)
}

func (d *driver) capabilities(w http.ResponseWriter, r *http.Request) {
	method := "capabilities"
	var response capabilitiesResponse

	response.Capabilities.Scope = "global"
	d.logRequest(method, "").Infof("response %v", response.Capabilities.Scope)
	json.NewEncoder(w).Encode(&response)
}
