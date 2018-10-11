/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

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
package sdk

import (
	"fmt"

	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
)

// VolumeServer is an implementation of the gRPC OpenStorageVolume interface
type VolumeServer struct {
	specHandler spec.SpecHandler
	driver      volume.VolumeDriver
	cluster     cluster.Cluster
}

// NewVolumeServer is a provider of VolumeServer
func NewVolumeServer(specHandler spec.SpecHandler, driver volume.VolumeDriver, cluster cluster.Cluster) *VolumeServer {
	return &VolumeServer{specHandler: specHandler, driver: driver, cluster: cluster}
}

// NewVolumeDriver is a provider of VolumeDriver.
// There could be other providers for volume driver but this is how it is being initialized in SDK,
// therefore, this provider func should be included in providers.ProviderSet
func NewVolumeDriver(driver DriverNameStr) (volume.VolumeDriver, error) {
	d, err := volumedrivers.Get(string(driver))
	if err != nil {
		return nil, fmt.Errorf("Unable to get driver %v info: %s", driver, err.Error())
	}
	return d, nil
}

// NewSpecHandler is a provider of SpecHandler
func NewSpecHandler() spec.SpecHandler {
	return spec.NewSpecHandler()
}
