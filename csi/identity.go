/*
Package csi is CSI driver interface for OSD
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
package csi

import (
	"fmt"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
)

const (
	// CSI 1.7 compatible
	csiDriverVersion = "1.7.0"

	// https://tools.ietf.org/html/rfc1035#section-2.3.1
	csiDriverNameFormat = "%s.openstorage.org"
)

// GetPluginCapabilities is a CSI API
func (s *OsdCsiServer) GetPluginCapabilities(
	ctx context.Context,
	req *csi.GetPluginCapabilitiesRequest,
) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			&csi.PluginCapability{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			&csi.PluginCapability{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
			&csi.PluginCapability{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_ONLINE,
					},
				},
			},
		},
	}, nil
}

// Probe is a CSI API
func (s *OsdCsiServer) Probe(
	ctx context.Context,
	req *csi.ProbeRequest,
) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}

// GetPluginInfo is a CSI API which returns the information about the plugin.
// This includes name, version, and any other OSD specific information
func (s *OsdCsiServer) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {

	var name string
	if len(s.csiDriverName) != 0 {
		name = s.csiDriverName
	} else {
		name = fmt.Sprintf(csiDriverNameFormat, s.driver.Name())
	}

	return &csi.GetPluginInfoResponse{
		Name:          name,
		VendorVersion: csiDriverVersion,

		// As OSD CSI Driver matures, add here more information
		Manifest: map[string]string{
			"driver": s.driver.Name(),
		},
	}, nil
}
