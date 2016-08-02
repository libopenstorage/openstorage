package flexvolume

import (
	"time"

	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"github.com/golang/protobuf/ptypes/empty"
)

type apiServer struct {
	protorpclog.Logger
	client Client
}

func newAPIServer(client Client) *apiServer {
	return &apiServer{protorpclog.NewLogger("flexvolume.API"), client}
}

func (a *apiServer) Init(_ context.Context, _ *empty.Empty) (_ *empty.Empty, err error) {
	defer func(start time.Time) { a.Log(nil, nil, err, time.Since(start)) }(time.Now())
	return checkClientError(a.client.Init())
}

func (a *apiServer) Attach(_ context.Context, request *AttachRequest) (_ *empty.Empty, err error) {
	defer func(start time.Time) { a.Log(request, nil, err, time.Since(start)) }(time.Now())
	return checkClientError(a.client.Attach(request.JsonOptions))
}

func (a *apiServer) Detach(_ context.Context, request *DetachRequest) (_ *empty.Empty, err error) {
	defer func(start time.Time) { a.Log(request, nil, err, time.Since(start)) }(time.Now())
	return checkClientError(a.client.Detach(request.MountDevice))
}

func (a *apiServer) Mount(_ context.Context, request *MountRequest) (_ *empty.Empty, err error) {
	defer func(start time.Time) { a.Log(request, nil, err, time.Since(start)) }(time.Now())
	return checkClientError(a.client.Mount(request.TargetMountDir, request.MountDevice, request.JsonOptions))
}

func (a *apiServer) Unmount(_ context.Context, request *UnmountRequest) (_ *empty.Empty, err error) {
	defer func(start time.Time) { a.Log(request, nil, err, time.Since(start)) }(time.Now())
	return checkClientError(a.client.Unmount(request.MountDir))
}

func checkClientError(err error) (*empty.Empty, error) {
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}
