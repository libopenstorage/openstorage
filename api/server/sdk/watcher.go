package sdk

import (
	"context"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/pborman/uuid"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	volumeEventType = "Volume"
)

type WathcerServer struct {
	volumeServer *VolumeServer

	watchConnections map[string][]*watchConnection
	volChan          chan *api.Volume
	nodeChan         chan *api.Node
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
	name         string
	eventChannel chan interface{}
}

func (w *watchConnection) callBack(vol interface{}) {
	w.eventChannel <- vol
}

func (s *WathcerServer) registerWatcher(client *watchConnection, eventType string) {
	if s.watchConnections == nil {
		s.watchConnections = make(map[string][]*watchConnection)
	}
	s.watchConnections[eventType] = append(s.watchConnections[eventType], client)
}

func (s *WathcerServer) removeWatcher(name string, eventType string) {
	var newWatchers []*watchConnection
	for _, client := range s.watchConnections[eventType] {
		if client.name == name {
			continue
		}
		newWatchers = append(newWatchers, client)
	}
	s.watchConnections[eventType] = newWatchers
}

func (s *WathcerServer) startWatcher(ctx context.Context) error {
	go s.startVolumeWatcher(ctx)
	return nil
}

func (s *WathcerServer) startVolumeWatcher(ctx context.Context) error {
	for {
		if s.volumeServer.driver(ctx) == nil {
			continue
		}

		if s.watchConnections == nil {
			s.watchConnections = make(map[string][]*watchConnection)
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
			for _, client := range s.watchConnections[volumeEventType] {
				go client.callBack(vol)
			}
		}
	}
}

func (w *WathcerServer) volumeWatch(
	req *api.SdkVolumeWatchRequest,
	stream api.OpenStorageWatch_WatchServer,
) error {
	ctx := stream.Context()
	if w.volumeServer.cluster() == nil || w.volumeServer.driver(ctx) == nil {
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

	client := watchConnection{
		name:         uuid.New(),
		eventChannel: make(chan interface{}, 2),
	}

	w.registerWatcher(&client, volumeEventType)
	defer w.removeWatcher(client.name, volumeEventType)

	group, _ := errgroup.WithContext(ctx)
	errChan := make(chan error)

	// spawn err-group process.
	group.Go(func() error {
		if vols, err := w.volumeServer.driver(ctx).Enumerate(&api.VolumeLocator{}, nil); err != nil {
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

			for event := range client.eventChannel {
				var vol *api.Volume
				if vol, ok = event.(*api.Volume); !ok {
					continue
				}
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
