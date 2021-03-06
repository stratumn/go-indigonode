// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-node/core/app/host/grpc (interfaces: Host_AddressesServer,Host_ConnectServer,Host_PeerAddressesServer,Host_ClearPeerAddressesServer)

// Package mockgrpc is a generated GoMock package.
package mockgrpc

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "github.com/stratumn/go-node/core/app/host/grpc"
	metadata "google.golang.org/grpc/metadata"
	reflect "reflect"
)

// MockHost_AddressesServer is a mock of Host_AddressesServer interface
type MockHost_AddressesServer struct {
	ctrl     *gomock.Controller
	recorder *MockHost_AddressesServerMockRecorder
}

// MockHost_AddressesServerMockRecorder is the mock recorder for MockHost_AddressesServer
type MockHost_AddressesServerMockRecorder struct {
	mock *MockHost_AddressesServer
}

// NewMockHost_AddressesServer creates a new mock instance
func NewMockHost_AddressesServer(ctrl *gomock.Controller) *MockHost_AddressesServer {
	mock := &MockHost_AddressesServer{ctrl: ctrl}
	mock.recorder = &MockHost_AddressesServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHost_AddressesServer) EXPECT() *MockHost_AddressesServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockHost_AddressesServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockHost_AddressesServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockHost_AddressesServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockHost_AddressesServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockHost_AddressesServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockHost_AddressesServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockHost_AddressesServer) Send(arg0 *grpc.Address) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockHost_AddressesServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockHost_AddressesServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockHost_AddressesServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockHost_AddressesServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockHost_AddressesServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockHost_AddressesServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockHost_AddressesServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockHost_AddressesServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockHost_AddressesServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockHost_AddressesServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockHost_AddressesServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockHost_AddressesServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockHost_AddressesServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockHost_AddressesServer)(nil).SetTrailer), arg0)
}

// MockHost_ConnectServer is a mock of Host_ConnectServer interface
type MockHost_ConnectServer struct {
	ctrl     *gomock.Controller
	recorder *MockHost_ConnectServerMockRecorder
}

// MockHost_ConnectServerMockRecorder is the mock recorder for MockHost_ConnectServer
type MockHost_ConnectServerMockRecorder struct {
	mock *MockHost_ConnectServer
}

// NewMockHost_ConnectServer creates a new mock instance
func NewMockHost_ConnectServer(ctrl *gomock.Controller) *MockHost_ConnectServer {
	mock := &MockHost_ConnectServer{ctrl: ctrl}
	mock.recorder = &MockHost_ConnectServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHost_ConnectServer) EXPECT() *MockHost_ConnectServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockHost_ConnectServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockHost_ConnectServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockHost_ConnectServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockHost_ConnectServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockHost_ConnectServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockHost_ConnectServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockHost_ConnectServer) Send(arg0 *grpc.Connection) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockHost_ConnectServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockHost_ConnectServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockHost_ConnectServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockHost_ConnectServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockHost_ConnectServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockHost_ConnectServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockHost_ConnectServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockHost_ConnectServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockHost_ConnectServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockHost_ConnectServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockHost_ConnectServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockHost_ConnectServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockHost_ConnectServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockHost_ConnectServer)(nil).SetTrailer), arg0)
}

// MockHost_PeerAddressesServer is a mock of Host_PeerAddressesServer interface
type MockHost_PeerAddressesServer struct {
	ctrl     *gomock.Controller
	recorder *MockHost_PeerAddressesServerMockRecorder
}

// MockHost_PeerAddressesServerMockRecorder is the mock recorder for MockHost_PeerAddressesServer
type MockHost_PeerAddressesServerMockRecorder struct {
	mock *MockHost_PeerAddressesServer
}

// NewMockHost_PeerAddressesServer creates a new mock instance
func NewMockHost_PeerAddressesServer(ctrl *gomock.Controller) *MockHost_PeerAddressesServer {
	mock := &MockHost_PeerAddressesServer{ctrl: ctrl}
	mock.recorder = &MockHost_PeerAddressesServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHost_PeerAddressesServer) EXPECT() *MockHost_PeerAddressesServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockHost_PeerAddressesServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockHost_PeerAddressesServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockHost_PeerAddressesServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockHost_PeerAddressesServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockHost_PeerAddressesServer) Send(arg0 *grpc.Address) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockHost_PeerAddressesServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockHost_PeerAddressesServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockHost_PeerAddressesServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockHost_PeerAddressesServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockHost_PeerAddressesServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockHost_PeerAddressesServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockHost_PeerAddressesServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockHost_PeerAddressesServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockHost_PeerAddressesServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockHost_PeerAddressesServer)(nil).SetTrailer), arg0)
}

// MockHost_ClearPeerAddressesServer is a mock of Host_ClearPeerAddressesServer interface
type MockHost_ClearPeerAddressesServer struct {
	ctrl     *gomock.Controller
	recorder *MockHost_ClearPeerAddressesServerMockRecorder
}

// MockHost_ClearPeerAddressesServerMockRecorder is the mock recorder for MockHost_ClearPeerAddressesServer
type MockHost_ClearPeerAddressesServerMockRecorder struct {
	mock *MockHost_ClearPeerAddressesServer
}

// NewMockHost_ClearPeerAddressesServer creates a new mock instance
func NewMockHost_ClearPeerAddressesServer(ctrl *gomock.Controller) *MockHost_ClearPeerAddressesServer {
	mock := &MockHost_ClearPeerAddressesServer{ctrl: ctrl}
	mock.recorder = &MockHost_ClearPeerAddressesServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHost_ClearPeerAddressesServer) EXPECT() *MockHost_ClearPeerAddressesServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockHost_ClearPeerAddressesServer) Context() context.Context {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockHost_ClearPeerAddressesServer) RecvMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockHost_ClearPeerAddressesServer) Send(arg0 *grpc.Address) error {
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockHost_ClearPeerAddressesServer) SendHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockHost_ClearPeerAddressesServer) SendMsg(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockHost_ClearPeerAddressesServer) SetHeader(arg0 metadata.MD) error {
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockHost_ClearPeerAddressesServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockHost_ClearPeerAddressesServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockHost_ClearPeerAddressesServer)(nil).SetTrailer), arg0)
}
