package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"go.pedge.io/dlog"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/units"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume/drivers"
)

const (
	// VolumeDriver is the string returned in the handshake protocol.
	VolumeDriver = "VolumeDriver"
)

var (
	nameRegex       = regexp.MustCompile("name=([0-9A-Za-z]+),?")
	sizeRegex       = regexp.MustCompile("size=([0-9A-Za-z]+),?")
	fsRegex         = regexp.MustCompile("fs=([0-9A-Za-z]+),?")
	bsRegex         = regexp.MustCompile("bs=([0-9]+),?")
	haRegex         = regexp.MustCompile("ha=([0-9]+),?")
	cosRegex        = regexp.MustCompile("cos=([A-Za-z]+),?")
	sharedRegex     = regexp.MustCompile("shared=([A-Za-z]+),?")
	passphraseRegex = regexp.MustCompile("passphrase=([0-9A-Za-z_@./#&+-]+),?")
)

// Implementation of the Docker volumes plugin specification.
type driver struct {
	restBase
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

func newVolumePlugin(name string) restServer {
	return &driver{restBase{name: name, version: "0.3"}}
}

func (d *driver) String() string {
	return d.name
}

func volDriverPath(method string) string {
	return fmt.Sprintf("/%s.%s", VolumeDriver, method)
}

func (d *driver) volNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume: " + e.Error())
	d.logRequest(request, id).Warnln(http.StatusNotFound, " ", err.Error())
	return err
}

func (d *driver) volNotMounted(request string, id string) error {
	err := fmt.Errorf("volume not mounted")
	d.logRequest(request, id).Debugln(http.StatusNotFound, " ", err.Error())
	return err
}

func (d *driver) Routes() []*Route {
	return []*Route{
		&Route{verb: "POST", path: volDriverPath("Create"), fn: d.create},
		&Route{verb: "POST", path: volDriverPath("Remove"), fn: d.remove},
		&Route{verb: "POST", path: volDriverPath("Mount"), fn: d.mount},
		&Route{verb: "POST", path: volDriverPath("Path"), fn: d.path},
		&Route{verb: "POST", path: volDriverPath("List"), fn: d.list},
		&Route{verb: "POST", path: volDriverPath("Get"), fn: d.get},
		&Route{verb: "POST", path: volDriverPath("Unmount"), fn: d.unmount},
		&Route{verb: "POST", path: volDriverPath("Capabilities"), fn: d.capabilities},
		&Route{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
		&Route{verb: "GET", path: "/status", fn: d.status},
	}
}

func (d *driver) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) errorResponse(w http.ResponseWriter, err error) {
	json.NewEncoder(w).Encode(&volumeResponse{Err: err.Error()})
}

func (d *driver) volFromName(name string) (*api.Volume, error) {
	v, err := volumedrivers.Get(d.name)
	if err != nil {
		return nil, fmt.Errorf("Cannot locate volume driver for %s: %s", d.name, err.Error())
	}
	vols, err := v.Inspect([]string{name})
	if err == nil && len(vols) == 1 {
		return vols[0], nil
	}
	vols, err = v.Enumerate(&api.VolumeLocator{Name: name}, nil)
	if err == nil && len(vols) == 1 {
		return vols[0], nil
	}
	return nil, fmt.Errorf("Cannot locate volume %s", name)
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
	d.logRequest(method, request.Name).Debugf("ID: %v", request.ID)
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

func (d *driver) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintln("osd plugin", d.version))
}

func (d *driver) cosLevel(cos string) (uint32, error) {
	switch cos {
	case "high", "3":
		return uint32(api.CosType_HIGH), nil
	case "medium", "2":
		return uint32(api.CosType_MEDIUM), nil
	case "low", "1", "":
		return uint32(api.CosType_LOW), nil
	}
	return uint32(api.CosType_LOW),
		fmt.Errorf("Cos must be one of %q | %q | %q", "high", "medium", "low")

}

func (d *driver) specFromOpts(
	Opts map[string]string,
) (*api.VolumeSpec, error) {
	spec := api.VolumeSpec{
		VolumeLabels: make(map[string]string),
		Format:       api.FSType_FS_TYPE_EXT4,
		HaLevel:      1,
	}

	for k, v := range Opts {
		switch k {
		case api.SpecEphemeral:
			spec.Ephemeral, _ = strconv.ParseBool(v)
		case api.SpecSize:
			if size, err := units.Parse(v); err != nil {
				return nil, err
			} else {
				spec.Size = uint64(size)
			}
		case api.SpecFilesystem:
			if value, err := api.FSTypeSimpleValueOf(v); err != nil {
				return nil, err
			} else {
				spec.Format = value
			}
		case api.SpecBlockSize:
			if blockSize, err := units.Parse(v); err != nil {
				return nil, err
			} else {
				spec.BlockSize = blockSize
			}
		case api.SpecHaLevel:
			haLevel, _ := strconv.ParseInt(v, 10, 64)
			spec.HaLevel = haLevel
		case api.SpecPriority:
			cos, _ := api.CosTypeSimpleValueOf(v)
			spec.Cos = cos
		case api.SpecDedupe:
			spec.Dedupe, _ = strconv.ParseBool(v)
		case api.SpecSnapshotInterval:
			snapshotInterval, _ := strconv.ParseUint(v, 10, 32)
			spec.SnapshotInterval = uint32(snapshotInterval)
		case api.SpecShared:
			if shared, err := strconv.ParseBool(v); err != nil {
				return nil, err
			} else {
				spec.Shared = shared
			}
		default:
			spec.VolumeLabels[k] = v
		}
	}
	return &spec, nil
}

func (d *driver) getVal(r *regexp.Regexp, str string) (bool, string) {
	found := r.FindString(str)
	if found == "" {
		return false, ""
	}

	submatches := r.FindStringSubmatch(str)
	if len(submatches) < 2 {
		return false, ""
	}

	val := submatches[1]

	return true, val
}

func (d *driver) specFromString(
	str string,
) (bool, *api.VolumeSpec, string) {
	name := ""
	spec := api.VolumeSpec{
		VolumeLabels: make(map[string]string),
		Format:       api.FSType_FS_TYPE_EXT4,
		HaLevel:      1,
	}

	ok, name := d.getVal(nameRegex, str)
	if !ok {
		return false, &spec, str
	}

	dlog.Infof("Parsing inline spec for volume %v", name)

	if ok, sz := d.getVal(sizeRegex, str); ok {
		size, _ := units.Parse(sz)
		spec.Size = uint64(size)
	}

	if ok, fs := d.getVal(fsRegex, str); ok {
		spec.Format, _ = api.FSTypeSimpleValueOf(fs)
	}

	if ok, bs := d.getVal(bsRegex, str); ok {
		val, _ := strconv.Atoi(bs)
		spec.BlockSize = int64(val)
	}

	if ok, ha := d.getVal(haRegex, str); ok {
		val, _ := strconv.Atoi(ha)
		spec.HaLevel = int64(val)
	}

	if ok, cos := d.getVal(cosRegex, str); ok {
		spec.Cos, _ = api.CosTypeSimpleValueOf(cos)
	}

	if ok, shared := d.getVal(sharedRegex, str); ok {
		if strings.EqualFold(shared, "true") {
			spec.Shared = true
		}
	}

	if ok, passphrase := d.getVal(passphraseRegex, str); ok {
		spec.Encrypted = true
		spec.Passphrase = passphrase
	}

	return true, &spec, name
}

func (d *driver) mountpath(request *mountRequest) string {
	return path.Join(volume.MountBase, request.Name)
}

func (d *driver) create(w http.ResponseWriter, r *http.Request) {
	method := "create"
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	// Parse options from the name - If the scheduler was unable to pass in
	// the volume spec via the API, we allow the spec to be passed in
	// via the name in the format: "key=value;key=value;name=volname"
	specParsed, spec, name := d.specFromString(request.Name)

	d.logRequest(method, name).Infoln("")
	// If we fail to find the volume, create it.
	if _, err = d.volFromName(name); err != nil {
		v, err := volumedrivers.Get(d.name)
		if err != nil {
			d.errorResponse(w, err)
			return
		}

		if !specParsed {
			spec, err = d.specFromOpts(request.Opts)
			if err != nil {
				d.errorResponse(w, err)
				return
			}
		}

		if _, err := v.Create(
			&api.VolumeLocator{Name: name},
			nil,
			spec,
		); err != nil {
			d.errorResponse(w, err)
			return
		}
	}
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) remove(w http.ResponseWriter, r *http.Request) {
	method := "remove"
	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf("Cannot locate volume driver")
		d.errorResponse(w, err)
		return
	}
	if err = v.Delete(request.Name); err != nil {
		d.errorResponse(w, err)
		return
	}
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driver) mount(w http.ResponseWriter, r *http.Request) {
	var response volumePathResponse
	method := "mount"

	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf("Cannot locate volume driver")
		d.errorResponse(w, err)
		return
	}

	request, err := d.decodeMount(method, w, r)
	if err != nil {
		d.errorResponse(w, err)
		return
	}

	vol, err := d.volFromName(request.Name)
	if err != nil {
		d.errorResponse(w, err)
		return
	}

	// If this is a block driver, first attach the volume.
	if v.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		attachPath, err := v.Attach(vol.Id)
		if err != nil {
			if err == volume.ErrVolAttachedOnRemoteNode {
				d.logRequest(method, request.Name).Infof("Volume is attached on a remote node... will attempt to mount it.")
			} else {
				d.logRequest(method, request.Name).Warnf("Cannot attach volume: %v", err.Error())
				d.errorResponse(w, err)
				return
			}
		} else {
			d.logRequest(method, request.Name).Debugf("response %v", attachPath)
		}
	}

	// Now mount it.
	response.Mountpoint = d.mountpath(request)
	os.MkdirAll(response.Mountpoint, 0755)

	err = v.Mount(vol.Id, response.Mountpoint)
	if err != nil {
		d.logRequest(method, request.Name).Warnf("Cannot mount volume %v, %v",
			response.Mountpoint, err)
		d.errorResponse(w, err)
		return
	}

	d.logRequest(method, request.Name).Infof("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) path(w http.ResponseWriter, r *http.Request) {
	method := "path"
	var response volumePathResponse

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	vol, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(w, e)
		return
	}

	d.logRequest(method, request.Name).Debugf("")

	if len(vol.AttachPath) == 0 || len(vol.AttachPath) == 0 {
		e := d.volNotMounted(method, request.Name)
		d.errorResponse(w, e)
		return
	}
	response.Mountpoint = vol.AttachPath[0]
	response.Mountpoint = path.Join(response.Mountpoint, config.DataDir)
	d.logRequest(method, request.Name).Debugf("response %v", response.Mountpoint)
	json.NewEncoder(w).Encode(&response)
}

func (d *driver) list(w http.ResponseWriter, r *http.Request) {
	method := "list"

	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf("Cannot locate volume driver: %v", err.Error())
		d.errorResponse(w, err)
		return
	}

	vols, err := v.Enumerate(nil, nil)
	if err != nil {
		d.errorResponse(w, err)
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

func (d *driver) get(w http.ResponseWriter, r *http.Request) {
	method := "get"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}
	vol, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(w, e)
		return
	}

	volInfo := volumeInfo{Name: request.Name}
	if len(vol.AttachPath) > 0 || len(vol.AttachPath) > 0 {
		volInfo.Mountpoint = path.Join(vol.AttachPath[0], config.DataDir)
	}

	json.NewEncoder(w).Encode(map[string]volumeInfo{"Volume": volInfo})
}

func (d *driver) unmount(w http.ResponseWriter, r *http.Request) {
	method := "unmount"

	v, err := volumedrivers.Get(d.name)
	if err != nil {
		d.logRequest(method, "").Warnf("Cannot locate volume driver: %v", err.Error())
		d.errorResponse(w, err)
		return
	}

	request, err := d.decodeMount(method, w, r)
	if err != nil {
		return
	}

	vol, err := d.volFromName(request.Name)
	if err != nil {
		e := d.volNotFound(method, request.Name, err, w)
		d.errorResponse(w, e)
		return
	}

	mountpoint := d.mountpath(request)
	err = v.Unmount(vol.Id, mountpoint)
	if err != nil {
		d.logRequest(method, request.Name).Warnf("Cannot unmount volume %v, %v",
			mountpoint, err)
		d.errorResponse(w, err)
		return
	}

	if v.Type() == api.DriverType_DRIVER_TYPE_BLOCK {
		_ = v.Detach(vol.Id)
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
