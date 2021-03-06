// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-node/app/chat/grpc (interfaces: Chat_GetHistoryServer)

// Package mockchat is a generated GoMock package.
package mockchat

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "github.com/stratumn/go-node/app/chat/grpc"
	metadata "google.golang.org/grpc/metadata"
	reflect "reflect"
)

// MockChat_GetHistoryServer is a mock of Chat_GetHistoryServer interface
type MockChat_GetHistoryServer struct {
	ctrl     *gomock.Controller
	recorder *MockChat_GetHistoryServerMockRecorder
}

// MockChat_GetHistoryServerMockRecorder is the mock recorder for MockChat_GetHistoryServer
type MockChat_GetHistoryServerMockRecorder struct {
	mock *MockChat_GetHistoryServer
}

// NewMockChat_GetHistoryServer creates a new mock instance
func NewMockChat_GetHistoryServer(ctrl *gomock.Controller) *MockChat_GetHistoryServer {
	mock := &MockChat_GetHistoryServer{ctrl: ctrl}
	mock.recorder = &MockChat_GetHistoryServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChat_GetHistoryServer) EXPECT() *MockChat_GetHistoryServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockChat_GetHistoryServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockChat_GetHistoryServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockChat_GetHistoryServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockChat_GetHistoryServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockChat_GetHistoryServer) Send(arg0 *grpc.DatedMessage) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockChat_GetHistoryServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockChat_GetHistoryServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockChat_GetHistoryServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockChat_GetHistoryServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockChat_GetHistoryServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockChat_GetHistoryServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockChat_GetHistoryServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockChat_GetHistoryServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockChat_GetHistoryServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockChat_GetHistoryServer)(nil).SetTrailer), arg0)
}
