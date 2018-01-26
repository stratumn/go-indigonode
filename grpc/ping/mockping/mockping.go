// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/grpc/ping (interfaces: Ping_PingServer)

package mockping

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	ping "github.com/stratumn/alice/grpc/ping"
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
func (_m *MockPing_PingServer) EXPECT() *MockPing_PingServerMockRecorder {
	return _m.recorder
}

// Context mocks base method
func (_m *MockPing_PingServer) Context() context.Context {
	ret := _m.ctrl.Call(_m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (_mr *MockPing_PingServerMockRecorder) Context() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Context", reflect.TypeOf((*MockPing_PingServer)(nil).Context))
}

// RecvMsg mocks base method
func (_m *MockPing_PingServer) RecvMsg(_param0 interface{}) error {
	ret := _m.ctrl.Call(_m, "RecvMsg", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (_mr *MockPing_PingServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "RecvMsg", reflect.TypeOf((*MockPing_PingServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (_m *MockPing_PingServer) Send(_param0 *ping.Response) error {
	ret := _m.ctrl.Call(_m, "Send", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (_mr *MockPing_PingServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Send", reflect.TypeOf((*MockPing_PingServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (_m *MockPing_PingServer) SendHeader(_param0 metadata.MD) error {
	ret := _m.ctrl.Call(_m, "SendHeader", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (_mr *MockPing_PingServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SendHeader", reflect.TypeOf((*MockPing_PingServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (_m *MockPing_PingServer) SendMsg(_param0 interface{}) error {
	ret := _m.ctrl.Call(_m, "SendMsg", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (_mr *MockPing_PingServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SendMsg", reflect.TypeOf((*MockPing_PingServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (_m *MockPing_PingServer) SetHeader(_param0 metadata.MD) error {
	ret := _m.ctrl.Call(_m, "SetHeader", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (_mr *MockPing_PingServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SetHeader", reflect.TypeOf((*MockPing_PingServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (_m *MockPing_PingServer) SetTrailer(_param0 metadata.MD) {
	_m.ctrl.Call(_m, "SetTrailer", _param0)
}

// SetTrailer indicates an expected call of SetTrailer
func (_mr *MockPing_PingServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SetTrailer", reflect.TypeOf((*MockPing_PingServer)(nil).SetTrailer), arg0)
}
