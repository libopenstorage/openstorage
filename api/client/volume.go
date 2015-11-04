package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

type volumeClient struct {
	c *Client
}

func newVolumeClient(c *Client) volume.VolumeDriver {
	return &volumeClient{c: c}
}

// String description of this driver.
func (v *volumeClient) String() string {
	return "VolumeDriver"
}

func (v *volumeClient) Type() api.DriverType {
	// Block drivers implement the superset.
	return api.Block
}

const (
	graphPath  = "/graph"
	volumePath = "/volumes"
	snapPath   = "/snapshot"
)

func (v *volumeClient) GraphDriverCreate(id, parent string) error {
	resp := ""
	err := v.c.Put().Resource(graphPath + "/create").Instance(id).Do().Unmarshal(&resp)
	if err != nil {
		return err
	}

	if resp != id {
		return fmt.Errorf("Invalid response: %v", resp)
	}

	return nil
}

func (v *volumeClient) GraphDriverRemove(id string) error {
	resp := ""
	err := v.c.Put().Resource(graphPath + "/remove").Instance(id).Do().Unmarshal(&resp)
	if err != nil {
		return err
	}

	if resp != id {
		return fmt.Errorf("Invalid response: %v", resp)
	}

	return nil
}

func (v *volumeClient) GraphDriverGet(id, mountLabel string) (string, error) {
	resp := ""
	err := v.c.Get().Resource(graphPath + "/inspect").Instance(id).Do().Unmarshal(&resp)
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (v *volumeClient) GraphDriverRelease(id string) error {
	resp := ""
	err := v.c.Put().Resource(graphPath + "/release").Instance(id).Do().Unmarshal(&resp)
	if err != nil {
		return err
	}

	if resp != id {
		return fmt.Errorf("Invalid response: %v", resp)
	}

	return nil
}

func (v *volumeClient) GraphDriverExists(id string) bool {
	resp := false
	v.c.Get().Resource(graphPath + "/exists").Instance(id).Do().Unmarshal(&resp)
	return resp
}

func (v *volumeClient) GraphDriverDiff(id, parent string) io.Writer {
	path := graphPath + "/diff?id=" + id + "&parent=" + parent
	resp := v.c.Get().Resource(path).Do()
	return bytes.NewBuffer(resp.body)
}

func (v *volumeClient) GraphDriverChanges(id, parent string) ([]api.GraphDriverChanges, error) {
	var changes []api.GraphDriverChanges
	err := v.c.Get().Resource(graphPath + "/changes").Instance(id).Do().Unmarshal(&changes)
	return changes, err
}

func (v *volumeClient) GraphDriverApplyDiff(id, parent string, diff io.Reader) (int, error) {
	resp := 0
	path := graphPath + "/diff?id=" + id + "&parent=" + parent

	b, err := ioutil.ReadAll(diff)
	if err != nil {
		return 0, err
	}

	err = v.c.Put().Resource(path).Instance(id).Body(b).Do().Unmarshal(&resp)
	if err != nil {
		return 0, err
	}
	return resp, nil
}

func (v *volumeClient) GraphDriverDiffSize(id, parent string) (int, error) {
	size := 0
	err := v.c.Get().Resource(graphPath + "/diffsize").Instance(id).Do().Unmarshal(&size)
	return size, err
}

// Create a new Vol for the specific volume spev.c.
// It returns a system generated VolumeID that uniquely identifies the volume
func (v *volumeClient) Create(locator api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	var response api.VolumeCreateResponse
	createReq := api.VolumeCreateRequest{
		Locator: locator,
		Source:  source,
		Spec:    spec,
	}
	err := v.c.Post().Resource(volumePath).Body(&createReq).Do().Unmarshal(&response)
	if err != nil {
		return api.VolumeID(""), err
	}
	if response.Error != "" {
		return api.VolumeID(""), errors.New(response.Error)
	}
	return response.ID, nil
}

// Status diagnostic information
func (v *volumeClient) Status() [][2]string {
	return [][2]string{}
}

// Inspect specified volumes.
// Errors ErrEnoEnt may be returned.
func (v *volumeClient) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	var vols []api.Volume

	if len(ids) == 0 {
		return nil, nil
	}
	req := v.c.Get().Resource(volumePath)

	for _, v := range ids {
		req.QueryOption(string(api.OptVolumeID), string(v))
	}
	err := req.Do().Unmarshal(&vols)
	if err != nil {
		return nil, err
	}
	return vols, nil
}

// Delete volume.
// Errors ErrEnoEnt, ErrVolHasSnaps may be returned.
func (v *volumeClient) Delete(volumeID api.VolumeID) error {
	var response api.VolumeResponse

	err := v.c.Delete().Resource(volumePath).Instance(string(volumeID)).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return nil
}

// Snap specified volume. IO to the underlying volume should be quiesced before
// calling this function.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	var response api.SnapCreateResponse
	createReq := api.SnapCreateRequest{
		ID:       volumeID,
		Readonly: readonly,
		Locator:  locator,
	}
	err := v.c.Post().Resource(snapPath).Body(&createReq).Do().Unmarshal(&response)
	if err != nil {
		return api.BadVolumeID, err
	}
	if response.Error != "" {
		return api.BadVolumeID, errors.New(response.Error)
	}
	return response.ID, nil
}

// Stats for specified volume.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) Stats(volumeID api.VolumeID) (api.Stats, error) {
	var stats api.Stats
	err := v.c.Get().Resource(volumePath + "/stats").Instance(string(volumeID)).Do().Unmarshal(&stats)
	if err != nil {
		return api.Stats{}, err
	}
	return stats, nil
}

// Alerts on this volume.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) Alerts(volumeID api.VolumeID) (api.Alerts, error) {
	var alerts api.Alerts
	err := v.c.Get().Resource(volumePath + "/alerts").Instance(string(volumeID)).Do().Unmarshal(&alerts)
	if err != nil {
		return api.Alerts{}, err
	}
	return alerts, nil
}

// Shutdown and cleanup.
func (v *volumeClient) Shutdown() {
	return
}

// Enumerate volumes that map to the volumeLocator. Locator fields may be regexp.
// If locator fields are left blank, this will return all volumes.
func (v *volumeClient) Enumerate(locator api.VolumeLocator, labels api.Labels) ([]api.Volume, error) {
	var vols []api.Volume
	req := v.c.Get().Resource(volumePath)
	if locator.Name != "" {
		req.QueryOption(string(api.OptName), locator.Name)
	}
	if len(locator.VolumeLabels) != 0 {
		req.QueryOptionLabel(string(api.OptLabel), locator.VolumeLabels)
	}
	if len(labels) != 0 {
		req.QueryOptionLabel(string(api.OptConfigLabel), labels)
	}
	err := req.Do().Unmarshal(&vols)
	if err != nil {
		return nil, err
	}
	return vols, nil
}

// Enumerate snaps for specified volume
// Count indicates the number of snaps populated.
func (v *volumeClient) SnapEnumerate(ids []api.VolumeID, snapLabels api.Labels) ([]api.Volume, error) {
	var snaps []api.Volume

	req := v.c.Get().Resource(snapPath)
	for _, v := range ids {
		req.QueryOption(string(api.OptVolumeID), string(v))
	}
	if len(snapLabels) != 0 {
		req.QueryOptionLabel(string(api.OptConfigLabel), snapLabels)
	}
	err := req.Do().Unmarshal(&snaps)
	if err != nil {
		return nil, err
	}
	return snaps, nil
}

// Attach map device to the host.
// On success the devicePath specifies location where the device is exported
// Errors ErrEnoEnt, ErrVolAttached may be returned.
func (v *volumeClient) Attach(volumeID api.VolumeID) (string, error) {
	var response api.VolumeSetResponse

	req := api.VolumeSetRequest{
		Action: &api.VolumeStateAction{Attach: api.ParamOn},
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return "", err
	}
	if response.VolumeResponse.Error != "" {
		return "", errors.New(response.VolumeResponse.Error)
	}
	return response.DevicePath, nil
}

// Detach device from the host.
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Detach(volumeID api.VolumeID) error {
	var response api.VolumeSetResponse
	req := api.VolumeSetRequest{
		Action: &api.VolumeStateAction{Attach: api.ParamOff},
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.VolumeResponse.Error != "" {
		return errors.New(response.VolumeResponse.Error)
	}
	return nil
}

// Mount volume at specified path
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Mount(volumeID api.VolumeID, mountpath string) error {
	var response api.VolumeSetResponse
	req := api.VolumeSetRequest{
		Action: &api.VolumeStateAction{Mount: api.ParamOn, MountPath: mountpath},
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.VolumeResponse.Error != "" {
		return errors.New(response.VolumeResponse.Error)
	}
	return nil
}

// Unmount volume at specified path
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Unmount(volumeID api.VolumeID, mountpath string) error {
	var response api.VolumeSetResponse
	req := api.VolumeSetRequest{
		Action: &api.VolumeStateAction{Mount: api.ParamOff, MountPath: mountpath},
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.VolumeResponse.Error != "" {
		return errors.New(response.VolumeResponse.Error)
	}
	return nil
}

// Update volume
func (v *volumeClient) Set(volumeID api.VolumeID, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	var response api.VolumeSetResponse
	req := api.VolumeSetRequest{
		Locator: locator,
		Spec:    spec,
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.VolumeResponse.Error != "" {
		return errors.New(response.VolumeResponse.Error)
	}
	return nil
}
