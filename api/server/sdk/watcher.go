package sdk

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WathcerServer struct {
	volumeServer *VolumeServer

	watchConnections []*watchConnection
	volChan          chan *api.Volume
}

// Watch streams a list of events back to clients. Events are resources that can typically fetch with openstorage API, such as Volumes,
// Nodes, Disks, etc.
//
// Implementation note: the flow of data starts from porx ETCD changes, then the data will be send to openstorage via a golang channel.
// Once the event object arrives at openstorage, it will be redistributed to a list of watch connections via another set of channels.
func (w *WathcerServer) Watch(req *api.SdkWatchRequest, stream api.OpenStorageWatch_WatchServer) error {
	go w.startWatcher(context.Background())
	if req.GetVolumeEvent() != nil {
		return w.volumeWatch(req.GetVolumeEvent(), stream)
	}
	return nil
}

type watchConnection struct {
	name    string
	volChan chan *api.Volume
}

func (w *watchConnection) callBack(vol *api.Volume) {
	w.volChan <- vol
}

func (s *WathcerServer) registerWatcher(client *watchConnection) {
	s.watchConnections = append(s.watchConnections, client)
}

func (s *WathcerServer) removeWatcher(name string) {
	defer close(s.volChan)

	var newWatchers []*watchConnection
	for _, client := range s.watchConnections {
		if client.name == name {
			continue
		}
		newWatchers = append(newWatchers, client)
	}
	s.watchConnections = newWatchers

}

func (s *WathcerServer) startWatcher(ctx context.Context) error {
	for {
		if s.volumeServer.driver(ctx) == nil {
			continue
		}

		volChan, err := s.volumeServer.driver(ctx).GetVolumeWatcher(&api.VolumeLocator{}, make(map[string]string))
		if err != nil {
			return status.Errorf(
				codes.Internal,
				"Failed to get volumes watcher: %v",
				err.Error())
		}
		s.volChan = volChan
		// volChan should be stuck here to wait for incoming events
		for vol := range s.volChan {
			for _, client := range s.watchConnections {
				go client.callBack(vol)
			}
		}

	}

}

func (s *WathcerServer) volumeWatch(
	req *api.SdkVolumeWatchRequest,
	stream api.OpenStorageWatch_WatchServer,
) error {
	ctx := stream.Context()
	if s.volumeServer.cluster() == nil || s.volumeServer.driver(ctx) == nil {
		return status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	// if input has deadline, ensure graceful exit within that deadline.
	deadline, ok := ctx.Deadline()
	var cancel context.CancelFunc
	if ok {
		// create a new context that will get done on deadline
		ctx, cancel = context.WithTimeout(ctx, deadline.Sub(time.Now()))
		defer cancel()
	}

	group, _ := errgroup.WithContext(ctx)
	errChan := make(chan error)

	client := watchConnection{
		name:    fmt.Sprintf("%v", rand.Intn(100)),
		volChan: make(chan *api.Volume, 10),
	}

	s.registerWatcher(&client)
	defer s.removeWatcher(client.name)

	// spawn err-group process.
	group.Go(func() error {
		if vols, err := s.volumeServer.driver(ctx).Enumerate(&api.VolumeLocator{}, nil); err != nil {
			return status.Errorf(
				codes.Internal,
				"Failed to enumerate volumes: %v",
				err.Error())
		} else {
			for _, vol := range vols {
				if !vol.IsPermitted(ctx, api.Ownership_Read) {
					continue
				}
				resp := convertVolumeToSdkReponse(vol)
				if err := stream.Send(resp); err != nil {
					return err
				}
			}

			for vol := range client.volChan {
				if !vol.IsPermitted(ctx, api.Ownership_Read) {
					continue
				}
				resp := convertVolumeToSdkReponse(vol)
				if err := stream.Send(resp); err != nil {
					return err
				}
			}

		}
		return nil
	})

	// wait for err-group processes to be done
	go func() {
		errChan <- group.Wait()
	}()

	// wait only as long as context deadline allows
	select {
	case err := <-errChan:
		if err != nil {
			return status.Errorf(codes.Internal, "error watching volume: %v", err)
		} else {
			return nil
		}
	case <-ctx.Done():
		return status.Error(codes.DeadlineExceeded,
			"Deadline is reached, server side func exiting")
	}

}

func convertVolumeToSdkReponse(vol *api.Volume) *api.SdkWatchResponse {
	resp := api.SdkVolumeWatchResponse{
		Volume: vol,
		Name:   vol.Locator.Name,
		Labels: vol.Locator.VolumeLabels,
	}
	volumeEventResponse := api.SdkWatchResponse_VolumeEvent{
		VolumeEvent: &resp,
	}
	return &api.SdkWatchResponse{
		EventType: &volumeEventResponse,
	}
}
