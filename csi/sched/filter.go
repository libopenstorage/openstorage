package sched

import (
	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

type Filter interface {
	PreVolumeCreate(req *csi.CreateVolumeRequest) (*csi.CreateVolumeRequest, error)
}
