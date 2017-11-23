// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/grpc/swarm (interfaces: Swarm_PeersServer,Swarm_ConnectionsServer)

// Package mockswarm is a generated GoMock package.
package mockswarm

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	swarm "github.com/stratumn/alice/grpc/swarm"
	metadata "google.golang.org/grpc/metadata"
)

// MockSwarm_PeersServer is a mock of Swarm_PeersServer interface
type MockSwarm_PeersServer struct {
	ctrl     *gomock.Controller
	recorder *MockSwarm_PeersServerMockRecorder
}

// MockSwarm_PeersServerMockRecorder is the mock recorder for MockSwarm_PeersServer
type MockSwarm_PeersServerMockRecorder struct {
	mock *MockSwarm_PeersServer
}

// NewMockSwarm_PeersServer creates a new mock instance
func NewMockSwarm_PeersServer(ctrl *gomock.Controller) *MockSwarm_PeersServer {
	mock := &MockSwarm_PeersServer{ctrl: ctrl}
	mock.recorder = &MockSwarm_PeersServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSwarm_PeersServer) EXPECT() *MockSwarm_PeersServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockSwarm_PeersServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockSwarm_PeersServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockSwarm_PeersServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockSwarm_PeersServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockSwarm_PeersServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockSwarm_PeersServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockSwarm_PeersServer) Send(arg0 *swarm.Peer) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockSwarm_PeersServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSwarm_PeersServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockSwarm_PeersServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockSwarm_PeersServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockSwarm_PeersServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockSwarm_PeersServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockSwarm_PeersServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockSwarm_PeersServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockSwarm_PeersServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockSwarm_PeersServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockSwarm_PeersServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockSwarm_PeersServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockSwarm_PeersServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockSwarm_PeersServer)(nil).SetTrailer), arg0)
}

// MockSwarm_ConnectionsServer is a mock of Swarm_ConnectionsServer interface
type MockSwarm_ConnectionsServer struct {
	ctrl     *gomock.Controller
	recorder *MockSwarm_ConnectionsServerMockRecorder
}

// MockSwarm_ConnectionsServerMockRecorder is the mock recorder for MockSwarm_ConnectionsServer
type MockSwarm_ConnectionsServerMockRecorder struct {
	mock *MockSwarm_ConnectionsServer
}

// NewMockSwarm_ConnectionsServer creates a new mock instance
func NewMockSwarm_ConnectionsServer(ctrl *gomock.Controller) *MockSwarm_ConnectionsServer {
	mock := &MockSwarm_ConnectionsServer{ctrl: ctrl}
	mock.recorder = &MockSwarm_ConnectionsServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSwarm_ConnectionsServer) EXPECT() *MockSwarm_ConnectionsServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockSwarm_ConnectionsServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockSwarm_ConnectionsServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockSwarm_ConnectionsServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockSwarm_ConnectionsServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockSwarm_ConnectionsServer) Send(arg0 *swarm.Connection) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockSwarm_ConnectionsServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockSwarm_ConnectionsServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockSwarm_ConnectionsServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockSwarm_ConnectionsServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockSwarm_ConnectionsServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockSwarm_ConnectionsServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockSwarm_ConnectionsServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockSwarm_ConnectionsServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockSwarm_ConnectionsServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockSwarm_ConnectionsServer)(nil).SetTrailer), arg0)
}
