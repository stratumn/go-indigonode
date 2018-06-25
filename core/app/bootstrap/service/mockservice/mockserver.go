// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/core/app/bootstrap/grpc (interfaces: Bootstrap_ListServer)

// Package mockservice is a generated GoMock package.
package mockservice

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	pb "github.com/stratumn/go-indigonode/core/app/bootstrap/pb"
	metadata "google.golang.org/grpc/metadata"
	reflect "reflect"
)

// MockBootstrap_ListServer is a mock of Bootstrap_ListServer interface
type MockBootstrap_ListServer struct {
	ctrl     *gomock.Controller
	recorder *MockBootstrap_ListServerMockRecorder
}

// MockBootstrap_ListServerMockRecorder is the mock recorder for MockBootstrap_ListServer
type MockBootstrap_ListServerMockRecorder struct {
	mock *MockBootstrap_ListServer
}

// NewMockBootstrap_ListServer creates a new mock instance
func NewMockBootstrap_ListServer(ctrl *gomock.Controller) *MockBootstrap_ListServer {
	mock := &MockBootstrap_ListServer{ctrl: ctrl}
	mock.recorder = &MockBootstrap_ListServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBootstrap_ListServer) EXPECT() *MockBootstrap_ListServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBootstrap_ListServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBootstrap_ListServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBootstrap_ListServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBootstrap_ListServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBootstrap_ListServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBootstrap_ListServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBootstrap_ListServer) Send(arg0 *pb.UpdateProposal) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBootstrap_ListServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBootstrap_ListServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBootstrap_ListServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBootstrap_ListServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBootstrap_ListServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBootstrap_ListServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBootstrap_ListServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBootstrap_ListServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBootstrap_ListServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBootstrap_ListServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBootstrap_ListServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBootstrap_ListServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBootstrap_ListServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBootstrap_ListServer)(nil).SetTrailer), arg0)
}
