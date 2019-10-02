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
	"context"
	"time"

	"github.com/libopenstorage/openstorage/alerts"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/proto/time"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	alertChunkSize = 128
)

// alertsServer implements api.OpenStorageAlertsServer.
// In order to use this server implementation just have
// alertsServer pointer properly instantiated with a valid
// alerts.filterDeleter.
type alertsServer struct {
	server serverAccessor
}

func (s *alertsServer) alert() alerts.FilterDeleter {
	return s.server.alert()
}

// NewAlertsServer provides an instance of alerts server interface.
func NewAlertsServer(filterDeleter alerts.FilterDeleter) api.OpenStorageAlertsServer {
	return &alertsServer{
		server: &sdkGrpcServer{alertHandler: filterDeleter},
	}
}

func getOpts(opts []*api.SdkAlertsOption) []alerts.Option {
	var options []alerts.Option

	for _, opt := range opts {
		switch opt.GetOpt().(type) {
		case *api.SdkAlertsOption_MinSeverityType:
			options = append(options,
				alerts.NewMinSeverityOption(opt.GetMinSeverityType()))
		case *api.SdkAlertsOption_IsCleared:
			options = append(options,
				alerts.NewFlagCheckOption(opt.GetIsCleared()))
		case *api.SdkAlertsOption_TimeSpan:
			options = append(options,
				alerts.NewTimeSpanOption(
					prototime.TimestampToTime(opt.GetTimeSpan().GetStartTime()),
					prototime.TimestampToTime(opt.GetTimeSpan().GetEndTime())))
		case *api.SdkAlertsOption_CountSpan:
			options = append(options,
				alerts.NewCountSpanOption(
					opt.GetCountSpan().GetMinCount(),
					opt.GetCountSpan().GetMaxCount()))
		}
	}

	return options
}

func getFilters(queries []*api.SdkAlertsQuery) []alerts.Filter {
	var filters []alerts.Filter

	// range over all queries
	for _, x := range queries {
		switch x.GetQuery().(type) {
		case *api.SdkAlertsQuery_ResourceTypeQuery:
			q := x.GetResourceTypeQuery()
			filters = append(filters,
				alerts.NewResourceTypeFilter(
					q.ResourceType,
					getOpts(x.GetOpts())...))
		case *api.SdkAlertsQuery_AlertTypeQuery:
			q := x.GetAlertTypeQuery()
			filters = append(filters,
				alerts.NewAlertTypeFilter(
					q.AlertType,
					q.ResourceType,
					getOpts(x.GetOpts())...))
		case *api.SdkAlertsQuery_ResourceIdQuery:
			q := x.GetResourceIdQuery()
			filters = append(filters,
				alerts.NewResourceIDFilter(
					q.ResourceId,
					q.AlertType,
					q.ResourceType,
					getOpts(x.GetOpts())...))
		}
	}

	return filters
}

// EnumerateWithFilters implements api.OpenStorageAlertsServer for alertsServer.
// Input context should ideally have a deadline, in which case, a
// graceful exit is ensured within that deadline.
func (g *alertsServer) EnumerateWithFilters(request *api.SdkAlertsEnumerateWithFiltersRequest, stream api.OpenStorageAlerts_EnumerateWithFiltersServer) error {
	ctx := stream.Context()

	if g.alert() == nil {
		return status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	queries := request.GetQueries()
	if queries == nil {
		return status.Error(codes.InvalidArgument, "Must provide at least one query")
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

	filters := getFilters(queries)

	// spawn err-group process.
	group.Go(func() error {
		if out, err := g.alert().Enumerate(filters...); err != nil {
			return err
		} else {
			for i := 0; ; i++ {
				start := i * alertChunkSize
				stop := (i + 1) * alertChunkSize

				if start >= len(out) {
					break
				}

				if stop > len(out) {
					stop = len(out)
				}

				resp := new(api.SdkAlertsEnumerateWithFiltersResponse)
				resp.Alerts = append(resp.Alerts, out[start:stop]...)
				if err := stream.Send(resp); err != nil {
					return err
				}

				if stop == len(out) {
					break
				}
			}
			return nil
		}
	})

	// wait for err-group processes to be done
	go func() {
		errChan <- group.Wait()
	}()

	// wait only as long as context deadline allows
	select {
	case err := <-errChan:
		if err != nil {
			return status.Errorf(codes.Internal, "error enumerating alerts: %v", err)
		} else {
			return nil
		}
	case <-ctx.Done():
		return status.Error(codes.DeadlineExceeded,
			"Deadline is reached, server side func exiting")
	}
}

// Delete implements api.OpenStorageAlertsServer for alertsServer.
// Input context should ideally have a deadline, in which case, a
// graceful exit is ensured within that deadline.
func (g *alertsServer) Delete(ctx context.Context,
	request *api.SdkAlertsDeleteRequest) (*api.SdkAlertsDeleteResponse, error) {
	if g.alert() == nil {
		return nil, status.Error(codes.Unavailable, "Resource has not been initialized")
	}

	queries := request.GetQueries()
	if queries == nil {
		return nil, status.Error(codes.InvalidArgument, "Must provide at least one query")
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

	resp := new(api.SdkAlertsDeleteResponse)

	filters := getFilters(queries)

	// spawn err-group process.
	group.Go(func() error {
		return g.alert().Delete(filters...)
	})

	// wait for err-group processes to be done
	go func() {
		errChan <- group.Wait()
	}()

	// wait only as long as context deadline allows
	select {
	case err := <-errChan:
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error deleting alerts: %v", err)
		} else {
			return resp, nil
		}
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded,
			"Deadline is reached, server side func exiting")
	}
}
