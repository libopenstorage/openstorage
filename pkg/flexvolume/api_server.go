package flexvolume

import (
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
)

type apiServer struct {
	client Client
}

func newAPIServer(client Client) *apiServer {
	return &apiServer{client}
}

func (a *apiServer) Init(_ context.Context, _ *google_protobuf.Empty) (*google_protobuf.Empty, error) {
	return checkClientError(a.client.Init())
}

func (a *apiServer) Attach(_ context.Context, request *AttachRequest) (*google_protobuf.Empty, error) {
	jsonOptions, err := bytesToJSONOptions(request.JsonOptions)
	if err != nil {
		return nil, err
	}
	return checkClientError(a.client.Attach(jsonOptions))
}

func (a *apiServer) Detach(_ context.Context, request *DetachRequest) (*google_protobuf.Empty, error) {
	return checkClientError(a.client.Detach(request.MountDevice))
}

func (a *apiServer) Mount(_ context.Context, request *MountRequest) (*google_protobuf.Empty, error) {
	jsonOptions, err := bytesToJSONOptions(request.JsonOptions)
	if err != nil {
		return nil, err
	}
	return checkClientError(a.client.Mount(request.TargetMountDir, request.MountDevice, jsonOptions))
}

func (a *apiServer) Unmount(_ context.Context, request *UnmountRequest) (*google_protobuf.Empty, error) {
	return checkClientError(a.client.Unmount(request.MountDir))
}

func checkClientError(err error) (*google_protobuf.Empty, error) {
	if err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}
