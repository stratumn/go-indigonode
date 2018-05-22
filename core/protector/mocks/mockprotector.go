// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/core/protector (interfaces: Protector)

package mocks

import (
	gomock "github.com/golang/mock/gomock"
	protector "github.com/stratumn/alice/core/protector"
	go_libp2p_transport "gx/ipfs/QmPUHzTLPZFYqv8WqcBTuMFYTgeom4uHHEaxzk7bd5GYZB/go-libp2p-transport"
	go_multiaddr "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	go_libp2p_peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	reflect "reflect"
)

// MockProtector is a mock of Protector interface
type MockProtector struct {
	ctrl     *gomock.Controller
	recorder *MockProtectorMockRecorder
}

// MockProtectorMockRecorder is the mock recorder for MockProtector
type MockProtectorMockRecorder struct {
	mock *MockProtector
}

// NewMockProtector creates a new mock instance
func NewMockProtector(ctrl *gomock.Controller) *MockProtector {
	mock := &MockProtector{ctrl: ctrl}
	mock.recorder = &MockProtectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockProtector) EXPECT() *MockProtectorMockRecorder {
	return _m.recorder
}

// AllowedAddrs mocks base method
func (_m *MockProtector) AllowedAddrs() []go_multiaddr.Multiaddr {
	ret := _m.ctrl.Call(_m, "AllowedAddrs")
	ret0, _ := ret[0].([]go_multiaddr.Multiaddr)
	return ret0
}

// AllowedAddrs indicates an expected call of AllowedAddrs
func (_mr *MockProtectorMockRecorder) AllowedAddrs() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "AllowedAddrs", reflect.TypeOf((*MockProtector)(nil).AllowedAddrs))
}

// AllowedPeers mocks base method
func (_m *MockProtector) AllowedPeers() []go_libp2p_peer.ID {
	ret := _m.ctrl.Call(_m, "AllowedPeers")
	ret0, _ := ret[0].([]go_libp2p_peer.ID)
	return ret0
}

// AllowedPeers indicates an expected call of AllowedPeers
func (_mr *MockProtectorMockRecorder) AllowedPeers() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "AllowedPeers", reflect.TypeOf((*MockProtector)(nil).AllowedPeers))
}

// Fingerprint mocks base method
func (_m *MockProtector) Fingerprint() []byte {
	ret := _m.ctrl.Call(_m, "Fingerprint")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Fingerprint indicates an expected call of Fingerprint
func (_mr *MockProtectorMockRecorder) Fingerprint() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Fingerprint", reflect.TypeOf((*MockProtector)(nil).Fingerprint))
}

// ListenForUpdates mocks base method
func (_m *MockProtector) ListenForUpdates(_param0 <-chan protector.NetworkUpdate) {
	_m.ctrl.Call(_m, "ListenForUpdates", _param0)
}

// ListenForUpdates indicates an expected call of ListenForUpdates
func (_mr *MockProtectorMockRecorder) ListenForUpdates(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "ListenForUpdates", reflect.TypeOf((*MockProtector)(nil).ListenForUpdates), arg0)
}

// Protect mocks base method
func (_m *MockProtector) Protect(_param0 go_libp2p_transport.Conn) (go_libp2p_transport.Conn, error) {
	ret := _m.ctrl.Call(_m, "Protect", _param0)
	ret0, _ := ret[0].(go_libp2p_transport.Conn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Protect indicates an expected call of Protect
func (_mr *MockProtectorMockRecorder) Protect(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Protect", reflect.TypeOf((*MockProtector)(nil).Protect), arg0)
}
