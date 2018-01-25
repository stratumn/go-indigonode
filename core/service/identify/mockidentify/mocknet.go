// Code generated by MockGen. DO NOT EDIT.
// Source: gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net (interfaces: Network)

// Package mockidentify is a generated GoMock package.
package mockidentify

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	go_libp2p_net "gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net"
	goprocess "gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess"
	go_multiaddr "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	go_libp2p_peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	go_libp2p_peerstore "gx/ipfs/QmeZVQzUrXqaszo24DAoHfGzcmCptN9JyngLkGAiEfk2x7/go-libp2p-peerstore"
	reflect "reflect"
)

// MockNetwork is a mock of Network interface
type MockNetwork struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkMockRecorder
}

// MockNetworkMockRecorder is the mock recorder for MockNetwork
type MockNetworkMockRecorder struct {
	mock *MockNetwork
}

// NewMockNetwork creates a new mock instance
func NewMockNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := &MockNetwork{ctrl: ctrl}
	mock.recorder = &MockNetworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockNetwork) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockNetworkMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockNetwork)(nil).Close))
}

// ClosePeer mocks base method
func (m *MockNetwork) ClosePeer(arg0 go_libp2p_peer.ID) error {
	ret := m.ctrl.Call(m, "ClosePeer", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ClosePeer indicates an expected call of ClosePeer
func (mr *MockNetworkMockRecorder) ClosePeer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClosePeer", reflect.TypeOf((*MockNetwork)(nil).ClosePeer), arg0)
}

// Connectedness mocks base method
func (m *MockNetwork) Connectedness(arg0 go_libp2p_peer.ID) go_libp2p_net.Connectedness {
	ret := m.ctrl.Call(m, "Connectedness", arg0)
	ret0, _ := ret[0].(go_libp2p_net.Connectedness)
	return ret0
}

// Connectedness indicates an expected call of Connectedness
func (mr *MockNetworkMockRecorder) Connectedness(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connectedness", reflect.TypeOf((*MockNetwork)(nil).Connectedness), arg0)
}

// Conns mocks base method
func (m *MockNetwork) Conns() []go_libp2p_net.Conn {
	ret := m.ctrl.Call(m, "Conns")
	ret0, _ := ret[0].([]go_libp2p_net.Conn)
	return ret0
}

// Conns indicates an expected call of Conns
func (mr *MockNetworkMockRecorder) Conns() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Conns", reflect.TypeOf((*MockNetwork)(nil).Conns))
}

// ConnsToPeer mocks base method
func (m *MockNetwork) ConnsToPeer(arg0 go_libp2p_peer.ID) []go_libp2p_net.Conn {
	ret := m.ctrl.Call(m, "ConnsToPeer", arg0)
	ret0, _ := ret[0].([]go_libp2p_net.Conn)
	return ret0
}

// ConnsToPeer indicates an expected call of ConnsToPeer
func (mr *MockNetworkMockRecorder) ConnsToPeer(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnsToPeer", reflect.TypeOf((*MockNetwork)(nil).ConnsToPeer), arg0)
}

// DialPeer mocks base method
func (m *MockNetwork) DialPeer(arg0 context.Context, arg1 go_libp2p_peer.ID) (go_libp2p_net.Conn, error) {
	ret := m.ctrl.Call(m, "DialPeer", arg0, arg1)
	ret0, _ := ret[0].(go_libp2p_net.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DialPeer indicates an expected call of DialPeer
func (mr *MockNetworkMockRecorder) DialPeer(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DialPeer", reflect.TypeOf((*MockNetwork)(nil).DialPeer), arg0, arg1)
}

// InterfaceListenAddresses mocks base method
func (m *MockNetwork) InterfaceListenAddresses() ([]go_multiaddr.Multiaddr, error) {
	ret := m.ctrl.Call(m, "InterfaceListenAddresses")
	ret0, _ := ret[0].([]go_multiaddr.Multiaddr)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InterfaceListenAddresses indicates an expected call of InterfaceListenAddresses
func (mr *MockNetworkMockRecorder) InterfaceListenAddresses() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InterfaceListenAddresses", reflect.TypeOf((*MockNetwork)(nil).InterfaceListenAddresses))
}

// Listen mocks base method
func (m *MockNetwork) Listen(arg0 ...go_multiaddr.Multiaddr) error {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Listen", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Listen indicates an expected call of Listen
func (mr *MockNetworkMockRecorder) Listen(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Listen", reflect.TypeOf((*MockNetwork)(nil).Listen), arg0...)
}

// ListenAddresses mocks base method
func (m *MockNetwork) ListenAddresses() []go_multiaddr.Multiaddr {
	ret := m.ctrl.Call(m, "ListenAddresses")
	ret0, _ := ret[0].([]go_multiaddr.Multiaddr)
	return ret0
}

// ListenAddresses indicates an expected call of ListenAddresses
func (mr *MockNetworkMockRecorder) ListenAddresses() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenAddresses", reflect.TypeOf((*MockNetwork)(nil).ListenAddresses))
}

// LocalPeer mocks base method
func (m *MockNetwork) LocalPeer() go_libp2p_peer.ID {
	ret := m.ctrl.Call(m, "LocalPeer")
	ret0, _ := ret[0].(go_libp2p_peer.ID)
	return ret0
}

// LocalPeer indicates an expected call of LocalPeer
func (mr *MockNetworkMockRecorder) LocalPeer() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalPeer", reflect.TypeOf((*MockNetwork)(nil).LocalPeer))
}

// NewStream mocks base method
func (m *MockNetwork) NewStream(arg0 context.Context, arg1 go_libp2p_peer.ID) (go_libp2p_net.Stream, error) {
	ret := m.ctrl.Call(m, "NewStream", arg0, arg1)
	ret0, _ := ret[0].(go_libp2p_net.Stream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewStream indicates an expected call of NewStream
func (mr *MockNetworkMockRecorder) NewStream(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewStream", reflect.TypeOf((*MockNetwork)(nil).NewStream), arg0, arg1)
}

// Notify mocks base method
func (m *MockNetwork) Notify(arg0 go_libp2p_net.Notifiee) {
	m.ctrl.Call(m, "Notify", arg0)
}

// Notify indicates an expected call of Notify
func (mr *MockNetworkMockRecorder) Notify(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockNetwork)(nil).Notify), arg0)
}

// Peers mocks base method
func (m *MockNetwork) Peers() []go_libp2p_peer.ID {
	ret := m.ctrl.Call(m, "Peers")
	ret0, _ := ret[0].([]go_libp2p_peer.ID)
	return ret0
}

// Peers indicates an expected call of Peers
func (mr *MockNetworkMockRecorder) Peers() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Peers", reflect.TypeOf((*MockNetwork)(nil).Peers))
}

// Peerstore mocks base method
func (m *MockNetwork) Peerstore() go_libp2p_peerstore.Peerstore {
	ret := m.ctrl.Call(m, "Peerstore")
	ret0, _ := ret[0].(go_libp2p_peerstore.Peerstore)
	return ret0
}

// Peerstore indicates an expected call of Peerstore
func (mr *MockNetworkMockRecorder) Peerstore() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Peerstore", reflect.TypeOf((*MockNetwork)(nil).Peerstore))
}

// Process mocks base method
func (m *MockNetwork) Process() goprocess.Process {
	ret := m.ctrl.Call(m, "Process")
	ret0, _ := ret[0].(goprocess.Process)
	return ret0
}

// Process indicates an expected call of Process
func (mr *MockNetworkMockRecorder) Process() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockNetwork)(nil).Process))
}

// SetConnHandler mocks base method
func (m *MockNetwork) SetConnHandler(arg0 go_libp2p_net.ConnHandler) {
	m.ctrl.Call(m, "SetConnHandler", arg0)
}

// SetConnHandler indicates an expected call of SetConnHandler
func (mr *MockNetworkMockRecorder) SetConnHandler(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetConnHandler", reflect.TypeOf((*MockNetwork)(nil).SetConnHandler), arg0)
}

// SetStreamHandler mocks base method
func (m *MockNetwork) SetStreamHandler(arg0 go_libp2p_net.StreamHandler) {
	m.ctrl.Call(m, "SetStreamHandler", arg0)
}

// SetStreamHandler indicates an expected call of SetStreamHandler
func (mr *MockNetworkMockRecorder) SetStreamHandler(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStreamHandler", reflect.TypeOf((*MockNetwork)(nil).SetStreamHandler), arg0)
}

// StopNotify mocks base method
func (m *MockNetwork) StopNotify(arg0 go_libp2p_net.Notifiee) {
	m.ctrl.Call(m, "StopNotify", arg0)
}

// StopNotify indicates an expected call of StopNotify
func (mr *MockNetworkMockRecorder) StopNotify(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopNotify", reflect.TypeOf((*MockNetwork)(nil).StopNotify), arg0)
}
