package client

import (
	"errors"
	"fmt"
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

const (
	volumePath = "/volume"
	snapPath   = "/snapshot"
)

// Create a new Vol for the specific volume spev.c.
// It returns a system generated VolumeID that uniquely identifies the volume
// If CreateOptions.FailIfExists is set and a volume matching the locator
// exists then this will fail with ErrEexist. Otherwise if a matching available
// volume is found then it is returned instead of creating a new volume.
func (v *volumeClient) Create(locator api.VolumeLocator,
	options *api.CreateOptions,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	var response api.VolumeCreateResponse
	createReq := api.VolumeCreateRequest{
		Locator: locator,
		Options: options,
		Spec:    spec,
	}
	err := v.c.Post().Resource(volumePath).Body(&createReq).Do().Unmarshal(&response)
	if err != nil {
		return api.VolumeID(""), err
	}
	if response.Status != "" {
		return api.VolumeID(""), errors.New(response.Status)
	}
	return response.ID, nil
}

// Inspect specified volumes.
// Errors ErrEnoEnt may be returned.
func (v *volumeClient) Inspect(ids []api.VolumeID) ([]api.Volume, error) {
	var vols []api.Volume

	if len(ids) == 0 {
		return nil, fmt.Errorf("No volume IDs provided")
	}
	req := v.c.Get().Resource(volumePath)

	for _, v := range ids {
		req.QueryOption(string(api.OptID), string(v))
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
	if response.Status != "" {
		return errors.New(response.Status)
	}
	return nil
}

// Snap specified volume. IO to the underlying volume should be quiesced before
// calling this function.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) Snapshot(volumeID api.VolumeID, labels api.Labels) (api.SnapID, error) {

	var response api.SnapCreateResponse
	createReq := api.SnapCreateRequest{
		ID:     volumeID,
		Labels: labels,
	}
	err := v.c.Post().Resource(snapPath).Body(&createReq).Do().Unmarshal(&response)
	if err != nil {
		return api.SnapID(""), err
	}
	if response.Status != "" {
		return api.SnapID(""), errors.New(response.Status)
	}
	return response.ID, nil
}

// SnapDelete snap specified by snapID.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) SnapDelete(snapID api.SnapID) error {
	var response api.VolumeResponse

	err := v.c.Delete().Resource(snapPath).Instance(string(snapID)).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.Status != "" {
		return errors.New(response.Status)
	}
	return nil
}

// SnapInspect provides details on this snapshot.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) SnapInspect(ids []api.SnapID) ([]api.VolumeSnap, error) {
	var snaps []api.VolumeSnap

	if len(ids) == 0 {
		return nil, fmt.Errorf("No snap IDs provided")
	}
	req := v.c.Get().Resource(snapPath)

	for _, v := range ids {
		req.QueryOption(string(api.OptID), string(v))
	}
	err := req.Do().Unmarshal(&snaps)
	if err != nil {
		return nil, err
	}
	return snaps, nil
}

// Stats for specified volume.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) Stats(volumeID api.VolumeID) (api.VolumeStats, error) {
	var stats api.VolumeStats
	err := v.c.Get().Resource(volumePath + "/stats").Instance(string(volumeID)).Do().Unmarshal(&stats)
	if err != nil {
		return api.VolumeStats{}, err
	}
	return stats, nil
}

// Alerts on this volume.
// Errors ErrEnoEnt may be returned
func (v *volumeClient) Alerts(volumeID api.VolumeID) (api.VolumeAlerts, error) {
	var alerts api.VolumeAlerts
	err := v.c.Get().Resource(volumePath + "/alerts").Instance(string(volumeID)).Do().Unmarshal(&alerts)
	if err != nil {
		return api.VolumeAlerts{}, err
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
		req.QueryOptionLabel(string(api.OptVolumeLabel), locator.VolumeLabels)
	}
	if len(labels) != 0 {
		req.QueryOptionLabel(string(api.OptConfigLabel), labels)
	}
	fmt.Println("Gotback ", vols)
	err := req.Do().Unmarshal(&vols)
	if err != nil {
		return nil, err
	}
	return vols, nil
}

// Enumerate snaps for specified volume
// Count indicates the number of snaps populated.
func (v *volumeClient) SnapEnumerate(locator api.VolumeLocator, labels api.Labels) ([]api.VolumeSnap, error) {
	var snaps []api.VolumeSnap
	req := v.c.Get().Resource(snapPath)
	if locator.Name != "" {
		req.QueryOption(string(api.OptName), locator.Name)
	}
	if len(locator.VolumeLabels) != 0 {
		req.QueryOptionLabel(string(api.OptVolumeLabel), locator.VolumeLabels)
	}
	if len(labels) != 0 {
		req.QueryOptionLabel(string(api.OptConfigLabel), labels)
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
	var response api.VolumeStateResponse

	req := api.VolumeStateRequest{
		Attach: true,
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return "", err
	}
	if response.Status != "" {
		return "", errors.New(response.Status)
	}
	return response.Path, nil
}

// Format volume according to spec provided in Create
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Format(volumeID api.VolumeID) error {
	var response api.VolumeStateResponse
	req := api.VolumeStateRequest{
		Format: true,
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.Status != "" {
		return errors.New(response.Status)
	}
	return nil
}

// Detach device from the host.
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Detach(volumeID api.VolumeID) error {
	var response api.VolumeStateResponse
	req := api.VolumeStateRequest{
		Attach: false,
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.Status != "" {
		return errors.New(response.Status)
	}
	return nil
}

// Mount volume at specified path
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Mount(volumeID api.VolumeID, mountpath string) error {
	var response api.VolumeStateResponse
	req := api.VolumeStateRequest{
		Attach: true,
		Mount:  true,
		Path:   mountpath,
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.Status != "" {
		return errors.New(response.Status)
	}
	return nil
}

// Unmount volume at specified path
// Errors ErrEnoEnt, ErrVolDetached may be returned.
func (v *volumeClient) Unmount(volumeID api.VolumeID, mountpath string) error {
	var response api.VolumeStateResponse
	req := api.VolumeStateRequest{
		Attach: true,
		Mount:  false,
	}
	err := v.c.Put().Resource(volumePath).Instance(string(volumeID)).Body(&req).Do().Unmarshal(&response)
	if err != nil {
		return err
	}
	if response.Status != "" {
		return errors.New(response.Status)
	}
	return nil
}
