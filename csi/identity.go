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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
)

const (
	csiDriverVersion = "0.1.0"
	csiDriverName    = "csi-osd"
)

var (
	csiVersion = &csi.Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
	}
)

// GetSupportedVersions is a CSI API which returns the supported CSI version
func (s *OsdCsiServer) GetSupportedVersions(
	context.Context,
	*csi.GetSupportedVersionsRequest) (*csi.GetSupportedVersionsResponse, error) {
	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Result_{
			Result: &csi.GetSupportedVersionsResponse_Result{
				SupportedVersions: []*csi.Version{
					csiVersion,
				},
			},
		},
	}, nil
}

// GetPluginInfo is a CSI API which returns the information about the plugin.
// This includes name, version, and any other OSD specific information
func (s *OsdCsiServer) GetPluginInfo(
	context.Context,
	*csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Reply: &csi.GetPluginInfoResponse_Result_{
			Result: &csi.GetPluginInfoResponse_Result{
				Name:          csiDriverName,
				VendorVersion: csiDriverVersion,
				Manifest: map[string]string{
					"driver": s.driver.Name(),
				},
			},
		},
	}, nil
}
