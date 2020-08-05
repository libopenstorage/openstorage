// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/libopenstorage/openstorage/api (interfaces: OpenStoragePoolServer,OpenStoragePoolClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	api "github.com/libopenstorage/openstorage/api"
	grpc "google.golang.org/grpc"
)

// MockOpenStoragePoolServer is a mock of OpenStoragePoolServer interface.
type MockOpenStoragePoolServer struct {
	ctrl     *gomock.Controller
	recorder *MockOpenStoragePoolServerMockRecorder
}

// MockOpenStoragePoolServerMockRecorder is the mock recorder for MockOpenStoragePoolServer.
type MockOpenStoragePoolServerMockRecorder struct {
	mock *MockOpenStoragePoolServer
}

// NewMockOpenStoragePoolServer creates a new mock instance.
func NewMockOpenStoragePoolServer(ctrl *gomock.Controller) *MockOpenStoragePoolServer {
	mock := &MockOpenStoragePoolServer{ctrl: ctrl}
	mock.recorder = &MockOpenStoragePoolServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOpenStoragePoolServer) EXPECT() *MockOpenStoragePoolServerMockRecorder {
	return m.recorder
}

// EnumerateRebalanceJobs mocks base method.
func (m *MockOpenStoragePoolServer) EnumerateRebalanceJobs(arg0 context.Context, arg1 *api.SdkEnumerateRebalanceJobsRequest) (*api.SdkEnumerateRebalanceJobsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnumerateRebalanceJobs", arg0, arg1)
	ret0, _ := ret[0].(*api.SdkEnumerateRebalanceJobsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnumerateRebalanceJobs indicates an expected call of EnumerateRebalanceJobs.
func (mr *MockOpenStoragePoolServerMockRecorder) EnumerateRebalanceJobs(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnumerateRebalanceJobs", reflect.TypeOf((*MockOpenStoragePoolServer)(nil).EnumerateRebalanceJobs), arg0, arg1)
}

// GetRebalanceJobStatus mocks base method.
func (m *MockOpenStoragePoolServer) GetRebalanceJobStatus(arg0 context.Context, arg1 *api.SdkGetRebalanceJobStatusRequest) (*api.SdkGetRebalanceJobStatusResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRebalanceJobStatus", arg0, arg1)
	ret0, _ := ret[0].(*api.SdkGetRebalanceJobStatusResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRebalanceJobStatus indicates an expected call of GetRebalanceJobStatus.
func (mr *MockOpenStoragePoolServerMockRecorder) GetRebalanceJobStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRebalanceJobStatus", reflect.TypeOf((*MockOpenStoragePoolServer)(nil).GetRebalanceJobStatus), arg0, arg1)
}

// Rebalance mocks base method.
func (m *MockOpenStoragePoolServer) Rebalance(arg0 context.Context, arg1 *api.SdkStorageRebalanceRequest) (*api.SdkStorageRebalanceResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rebalance", arg0, arg1)
	ret0, _ := ret[0].(*api.SdkStorageRebalanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Rebalance indicates an expected call of Rebalance.
func (mr *MockOpenStoragePoolServerMockRecorder) Rebalance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rebalance", reflect.TypeOf((*MockOpenStoragePoolServer)(nil).Rebalance), arg0, arg1)
}

// Resize mocks base method.
func (m *MockOpenStoragePoolServer) Resize(arg0 context.Context, arg1 *api.SdkStoragePoolResizeRequest) (*api.SdkStoragePoolResizeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resize", arg0, arg1)
	ret0, _ := ret[0].(*api.SdkStoragePoolResizeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Resize indicates an expected call of Resize.
func (mr *MockOpenStoragePoolServerMockRecorder) Resize(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resize", reflect.TypeOf((*MockOpenStoragePoolServer)(nil).Resize), arg0, arg1)
}

// UpdateRebalanceJobState mocks base method.
func (m *MockOpenStoragePoolServer) UpdateRebalanceJobState(arg0 context.Context, arg1 *api.SdkUpdateRebalanceJobRequest) (*api.SdkUpdateRebalanceJobResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateRebalanceJobState", arg0, arg1)
	ret0, _ := ret[0].(*api.SdkUpdateRebalanceJobResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateRebalanceJobState indicates an expected call of UpdateRebalanceJobState.
func (mr *MockOpenStoragePoolServerMockRecorder) UpdateRebalanceJobState(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRebalanceJobState", reflect.TypeOf((*MockOpenStoragePoolServer)(nil).UpdateRebalanceJobState), arg0, arg1)
}

// MockOpenStoragePoolClient is a mock of OpenStoragePoolClient interface.
type MockOpenStoragePoolClient struct {
	ctrl     *gomock.Controller
	recorder *MockOpenStoragePoolClientMockRecorder
}

// MockOpenStoragePoolClientMockRecorder is the mock recorder for MockOpenStoragePoolClient.
type MockOpenStoragePoolClientMockRecorder struct {
	mock *MockOpenStoragePoolClient
}

// NewMockOpenStoragePoolClient creates a new mock instance.
func NewMockOpenStoragePoolClient(ctrl *gomock.Controller) *MockOpenStoragePoolClient {
	mock := &MockOpenStoragePoolClient{ctrl: ctrl}
	mock.recorder = &MockOpenStoragePoolClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOpenStoragePoolClient) EXPECT() *MockOpenStoragePoolClientMockRecorder {
	return m.recorder
}

// EnumerateRebalanceJobs mocks base method.
func (m *MockOpenStoragePoolClient) EnumerateRebalanceJobs(arg0 context.Context, arg1 *api.SdkEnumerateRebalanceJobsRequest, arg2 ...grpc.CallOption) (*api.SdkEnumerateRebalanceJobsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EnumerateRebalanceJobs", varargs...)
	ret0, _ := ret[0].(*api.SdkEnumerateRebalanceJobsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnumerateRebalanceJobs indicates an expected call of EnumerateRebalanceJobs.
func (mr *MockOpenStoragePoolClientMockRecorder) EnumerateRebalanceJobs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnumerateRebalanceJobs", reflect.TypeOf((*MockOpenStoragePoolClient)(nil).EnumerateRebalanceJobs), varargs...)
}

// GetRebalanceJobStatus mocks base method.
func (m *MockOpenStoragePoolClient) GetRebalanceJobStatus(arg0 context.Context, arg1 *api.SdkGetRebalanceJobStatusRequest, arg2 ...grpc.CallOption) (*api.SdkGetRebalanceJobStatusResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRebalanceJobStatus", varargs...)
	ret0, _ := ret[0].(*api.SdkGetRebalanceJobStatusResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRebalanceJobStatus indicates an expected call of GetRebalanceJobStatus.
func (mr *MockOpenStoragePoolClientMockRecorder) GetRebalanceJobStatus(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRebalanceJobStatus", reflect.TypeOf((*MockOpenStoragePoolClient)(nil).GetRebalanceJobStatus), varargs...)
}

// Rebalance mocks base method.
func (m *MockOpenStoragePoolClient) Rebalance(arg0 context.Context, arg1 *api.SdkStorageRebalanceRequest, arg2 ...grpc.CallOption) (*api.SdkStorageRebalanceResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Rebalance", varargs...)
	ret0, _ := ret[0].(*api.SdkStorageRebalanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Rebalance indicates an expected call of Rebalance.
func (mr *MockOpenStoragePoolClientMockRecorder) Rebalance(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rebalance", reflect.TypeOf((*MockOpenStoragePoolClient)(nil).Rebalance), varargs...)
}

// Resize mocks base method.
func (m *MockOpenStoragePoolClient) Resize(arg0 context.Context, arg1 *api.SdkStoragePoolResizeRequest, arg2 ...grpc.CallOption) (*api.SdkStoragePoolResizeResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Resize", varargs...)
	ret0, _ := ret[0].(*api.SdkStoragePoolResizeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Resize indicates an expected call of Resize.
func (mr *MockOpenStoragePoolClientMockRecorder) Resize(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resize", reflect.TypeOf((*MockOpenStoragePoolClient)(nil).Resize), varargs...)
}

// UpdateRebalanceJobState mocks base method.
func (m *MockOpenStoragePoolClient) UpdateRebalanceJobState(arg0 context.Context, arg1 *api.SdkUpdateRebalanceJobRequest, arg2 ...grpc.CallOption) (*api.SdkUpdateRebalanceJobResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateRebalanceJobState", varargs...)
	ret0, _ := ret[0].(*api.SdkUpdateRebalanceJobResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateRebalanceJobState indicates an expected call of UpdateRebalanceJobState.
func (mr *MockOpenStoragePoolClientMockRecorder) UpdateRebalanceJobState(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRebalanceJobState", reflect.TypeOf((*MockOpenStoragePoolClient)(nil).UpdateRebalanceJobState), varargs...)
}
