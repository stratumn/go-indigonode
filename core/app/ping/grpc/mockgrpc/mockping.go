// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-node/core/app/ping/grpc (interfaces: Ping_PingServer)

// Package mockgrpc is a generated GoMock package.
package mockgrpc

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "github.com/stratumn/go-node/core/app/ping/grpc"
	metadata "google.golang.org/grpc/metadata"
	reflect "reflect"
)

// MockPing_PingServer is a mock of Ping_PingServer interface
type MockPing_PingServer struct {
	ctrl     *gomock.Controller
	recorder *MockPing_PingServerMockRecorder
}

// MockPing_PingServerMockRecorder is the mock recorder for MockPing_PingServer
type MockPing_PingServerMockRecorder struct {
	mock *MockPing_PingServer
}

// NewMockPing_PingServer creates a new mock instance
func NewMockPing_PingServer(ctrl *gomock.Controller) *MockPing_PingServer {
	mock := &MockPing_PingServer{ctrl: ctrl}
	mock.recorder = &MockPing_PingServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPing_PingServer) EXPECT() *MockPing_PingServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockPing_PingServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockPing_PingServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockPing_PingServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockPing_PingServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockPing_PingServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockPing_PingServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockPing_PingServer) Send(arg0 *grpc.Response) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockPing_PingServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockPing_PingServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockPing_PingServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockPing_PingServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockPing_PingServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockPing_PingServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockPing_PingServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockPing_PingServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockPing_PingServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockPing_PingServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockPing_PingServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockPing_PingServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockPing_PingServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockPing_PingServer)(nil).SetTrailer), arg0)
}
