/*
Package util provides utility functions for OSD servers and drivers.
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package util

import (
	"context"
	"fmt"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
)

// VolumeFromName returns the volume object associated with the specified name.
func VolumeFromName(v volume.VolumeDriver, name string) (*api.Volume, error) {
	vols, err := v.Inspect([]string{name})
	if err == nil && len(vols) == 1 {
		return vols[0], nil
	}
	vols, err = v.Enumerate(&api.VolumeLocator{Name: name}, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to locate volume %s. Error: %s", name, err.Error())
	} else if err == nil && len(vols) == 1 {
		return vols[0], nil
	}
	return nil, fmt.Errorf("Cannot locate volume with name %s", name)
}

// VolumeFromIdSdk uses the SDK to fetch the volume object associated with the specified id.
func VolumeFromIdSdk(ctx context.Context, volumes api.OpenStorageVolumeClient, id string) (*api.Volume, error) {
	inspectResp, err := volumes.Inspect(ctx, &api.SdkVolumeInspectRequest{
		VolumeId: id,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot locate volume with id %s", id)
	}
	return inspectResp.Volume, nil
}

// VolumeIdFromNameSdk uses the SDK to fetch the volume id associated with the specified name.
func VolumeIdFromNameSdk(ctx context.Context, volumes api.OpenStorageVolumeClient, name string) (string, error) {
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

// VolumeFromNameSdk uses the SDK to fetch the volume associated with a specified name.
func VolumeFromNameSdk(ctx context.Context, volumes api.OpenStorageVolumeClient, name string) (*api.Volume, error) {
	// get volume id
	volId, err := VolumeIdFromNameSdk(ctx, volumes, name)
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
