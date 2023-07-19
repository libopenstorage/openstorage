// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/libopenstorage/openstorage/api (interfaces: OpenStorageWatchServer,OpenStorageWatchClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	api "github.com/libopenstorage/openstorage/api"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// MockOpenStorageWatchServer is a mock of OpenStorageWatchServer interface
type MockOpenStorageWatchServer struct {
	ctrl     *gomock.Controller
	recorder *MockOpenStorageWatchServerMockRecorder
}

// MockOpenStorageWatchServerMockRecorder is the mock recorder for MockOpenStorageWatchServer
type MockOpenStorageWatchServerMockRecorder struct {
	mock *MockOpenStorageWatchServer
}

// NewMockOpenStorageWatchServer creates a new mock instance
func NewMockOpenStorageWatchServer(ctrl *gomock.Controller) *MockOpenStorageWatchServer {
	mock := &MockOpenStorageWatchServer{ctrl: ctrl}
	mock.recorder = &MockOpenStorageWatchServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOpenStorageWatchServer) EXPECT() *MockOpenStorageWatchServerMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockOpenStorageWatchServer) Watch(arg0 *api.SdkWatchRequest, arg1 api.OpenStorageWatch_WatchServer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Watch indicates an expected call of Watch
func (mr *MockOpenStorageWatchServerMockRecorder) Watch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockOpenStorageWatchServer)(nil).Watch), arg0, arg1)
}

// MockOpenStorageWatchClient is a mock of OpenStorageWatchClient interface
type MockOpenStorageWatchClient struct {
	ctrl     *gomock.Controller
	recorder *MockOpenStorageWatchClientMockRecorder
}

// MockOpenStorageWatchClientMockRecorder is the mock recorder for MockOpenStorageWatchClient
type MockOpenStorageWatchClientMockRecorder struct {
	mock *MockOpenStorageWatchClient
}

// NewMockOpenStorageWatchClient creates a new mock instance
func NewMockOpenStorageWatchClient(ctrl *gomock.Controller) *MockOpenStorageWatchClient {
	mock := &MockOpenStorageWatchClient{ctrl: ctrl}
	mock.recorder = &MockOpenStorageWatchClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockOpenStorageWatchClient) EXPECT() *MockOpenStorageWatchClientMockRecorder {
	return m.recorder
}

// Watch mocks base method
func (m *MockOpenStorageWatchClient) Watch(arg0 context.Context, arg1 *api.SdkWatchRequest, arg2 ...grpc.CallOption) (api.OpenStorageWatch_WatchClient, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Watch", varargs...)
	ret0, _ := ret[0].(api.OpenStorageWatch_WatchClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Watch indicates an expected call of Watch
func (mr *MockOpenStorageWatchClientMockRecorder) Watch(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockOpenStorageWatchClient)(nil).Watch), varargs...)
}