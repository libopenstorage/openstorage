package sdk

import (
	"context"
	"sync"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	volumeEventType = "Volume"

	eventChannelSize = 100
	streamTimeout    = 2 * time.Second
)

type WatcherServer struct {
	volumeServer *VolumeServer

	watchConnections map[string][]*watchConnection
	volumeChannel    chan *api.Volume
	sync.RWMutex
}

// Watch streams a list of events back to clients. Events are resources that can typically fetch with openstorage API, such as Volumes,
// Nodes, Disks, etc.
//
// Implementation note: the flow of data starts from porx ETCD changes, then the data will be send to openstorage via a golang channel.
// Once the event object arrives at openstorage, it will be redistributed to a list of watch connections via another set of channels.
func (w *WatcherServer) Watch(req *api.SdkWatchRequest, stream api.OpenStorageWatch_WatchServer) error {
	if req.GetVolumeEvent() != nil {
		return w.volumeWatch(req.GetVolumeEvent(), stream)
	}
	return status.Errorf(codes.InvalidArgument, "invalid request type for watcher %v", req)
}

type watchConnection struct {
	name         string
	eventType    string
	eventChannel chan interface{}
}

func (w *watchConnection) callBack(eventData interface{}) {
	select {
	case w.eventChannel <- eventData:
		logrus.Debugf("successfully callback event for %v", w.name)
	default:
		logrus.Warnf("failed to send eventData for %v with event type %v", w.name, w.eventType)
	}

}

func (s *WatcherServer) registerWatcher(client *watchConnection, eventType string) {
	s.Lock()
	defer s.Unlock()
	if s.watchConnections == nil {
		s.watchConnections = make(map[string][]*watchConnection)
	}
	s.watchConnections[eventType] = append(s.watchConnections[eventType], client)
	logrus.Debugf("successfully register watcher %v", client.name)
}

func (s *WatcherServer) removeWatcher(name string, eventType string) {
	s.Lock()
	defer s.Unlock()
	var newWatchers []*watchConnection
	for _, client := range s.watchConnections[eventType] {
		if client.name == name {
			// clean up client go channel
			close(client.eventChannel)
			continue
		}
		newWatchers = append(newWatchers, client)
	}
	s.watchConnections[eventType] = newWatchers
}

func (s *WatcherServer) startWatcher(ctx context.Context, done chan bool) error {
	group, _ := errgroup.WithContext(ctx)
	errChan := make(chan error)
	group.Go(func() error {
		return s.startVolumeWatcher(ctx, done)
	})

	// wait for err-group processes to be done
	go func() {
		errChan <- group.Wait()
	}()

	// wait only as long as context deadline allows
	select {
	case err := <-errChan:
		if err != nil {
			return status.Errorf(codes.Internal, "error starting watcher: %v", err)
		} else {
			return nil
		}
	case <-ctx.Done():
		return status.Error(codes.DeadlineExceeded,
			"Deadline is reached, server side func exiting")
	}
}

func (s *WatcherServer) startVolumeWatcher(ctx context.Context, done chan bool) error {
	if s.watchConnections == nil {
		s.watchConnections = make(map[string][]*watchConnection)
	}

	// wait for driver to be initialized with an non-empty volume watcher
	for {
		if s.volumeServer.driver(ctx) == nil {
			continue
		}

		volumeChannel, err := s.volumeServer.driver(ctx).GetVolumeWatcher(&api.VolumeLocator{}, make(map[string]string))
		if err != nil {
			logrus.Warnf("error getting volume watcher %v", err)
		}
		if volumeChannel == nil {
			continue
		}
		s.volumeChannel = volumeChannel
		goto volumeWatch
	}
	// volumeChannel should be waiting for incoming events
volumeWatch:
	for {
		select {
		case vol := <-s.volumeChannel:
			s.RLock()
			for _, client := range s.watchConnections[volumeEventType] {
				go client.callBack(vol)
			}
			s.RUnlock()
		case <-done:
			logrus.Infof("exiting volume watcher\n")
			break volumeWatch
		}
	}

	return nil
}

func (w *WatcherServer) volumeWatch(
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
		ctx, cancel = context.WithTimeout(ctx, time.Until(deadline))
		defer cancel()
	}

	client := watchConnection{
		name:         uuid.New(),
		eventType:    volumeEventType,
		eventChannel: make(chan interface{}, eventChannelSize),
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
				resp := convertApiVolumeToSdkReponse(vol)
				if err := stream.Send(resp); err != nil {
					return err
				}
			}

			for event := range client.eventChannel {

				// create a new context that will return error if execution took more than streamTimeout
				timeoutCtx, timeoutCancelled := context.WithTimeout(ctx, streamTimeout)
				defer timeoutCancelled()
				var vol *api.Volume
				if vol, ok = event.(*api.Volume); !ok {
					logrus.Warnf("error converting event to Volume Type for event %v", event)
					continue
				}
				if !vol.IsPermitted(ctx, api.Ownership_Read) {
					continue
				}

				resp := convertApiVolumeToSdkReponse(vol)
				err := stream.Send(resp)

				if err != nil {
					logrus.Warnf("error sending stream: %v", err)
					return err
				}
				if timeoutCtx.Err() != nil {
					logrus.Warnf("context error: %v", timeoutCtx.Err())
					return timeoutCtx.Err()
				}
				timeoutCancelled()
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

func convertApiVolumeToSdkReponse(vol *api.Volume) *api.SdkWatchResponse {
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
