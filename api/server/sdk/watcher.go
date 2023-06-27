package sdk

import "github.com/libopenstorage/openstorage/api"

type WathcerServer struct {
	volumeServer *VolumeServer
}

func (w *WathcerServer) Watch(req *api.SdkWatchRequest, stream api.OpenStorageWatch_WatchServer) error {
	if req.GetVolumeEvent() != nil {
		return w.volumeServer.Watch(req.GetVolumeEvent(), stream)
	}
	return nil

}
