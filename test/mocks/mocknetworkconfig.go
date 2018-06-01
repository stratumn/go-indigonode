// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/core/protector (interfaces: NetworkConfig)

package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	protector "github.com/stratumn/alice/pb/protector"
	go_multicodec "gx/ipfs/QmRDePEiL4Yupq5EkcK3L3ko3iMgYaqUdLu7xc1kqs7dnV/go-multicodec"
	go_multiaddr "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	go_libp2p_peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	reflect "reflect"
)

// MockNetworkConfig is a mock of NetworkConfig interface
type MockNetworkConfig struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkConfigMockRecorder
}

// MockNetworkConfigMockRecorder is the mock recorder for MockNetworkConfig
type MockNetworkConfigMockRecorder struct {
	mock *MockNetworkConfig
}

// NewMockNetworkConfig creates a new mock instance
func NewMockNetworkConfig(ctrl *gomock.Controller) *MockNetworkConfig {
	mock := &MockNetworkConfig{ctrl: ctrl}
	mock.recorder = &MockNetworkConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockNetworkConfig) EXPECT() *MockNetworkConfigMockRecorder {
	return _m.recorder
}

// AddPeer mocks base method
func (_m *MockNetworkConfig) AddPeer(_param0 context.Context, _param1 go_libp2p_peer.ID, _param2 []go_multiaddr.Multiaddr) error {
	ret := _m.ctrl.Call(_m, "AddPeer", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddPeer indicates an expected call of AddPeer
func (_mr *MockNetworkConfigMockRecorder) AddPeer(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "AddPeer", reflect.TypeOf((*MockNetworkConfig)(nil).AddPeer), arg0, arg1, arg2)
}

// AllowedPeers mocks base method
func (_m *MockNetworkConfig) AllowedPeers(_param0 context.Context) []go_libp2p_peer.ID {
	ret := _m.ctrl.Call(_m, "AllowedPeers", _param0)
	ret0, _ := ret[0].([]go_libp2p_peer.ID)
	return ret0
}

// AllowedPeers indicates an expected call of AllowedPeers
func (_mr *MockNetworkConfigMockRecorder) AllowedPeers(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "AllowedPeers", reflect.TypeOf((*MockNetworkConfig)(nil).AllowedPeers), arg0)
}

// Encode mocks base method
func (_m *MockNetworkConfig) Encode(_param0 go_multicodec.Encoder) error {
	ret := _m.ctrl.Call(_m, "Encode", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Encode indicates an expected call of Encode
func (_mr *MockNetworkConfigMockRecorder) Encode(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Encode", reflect.TypeOf((*MockNetworkConfig)(nil).Encode), arg0)
}

// IsAllowed mocks base method
func (_m *MockNetworkConfig) IsAllowed(_param0 context.Context, _param1 go_libp2p_peer.ID) bool {
	ret := _m.ctrl.Call(_m, "IsAllowed", _param0, _param1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAllowed indicates an expected call of IsAllowed
func (_mr *MockNetworkConfigMockRecorder) IsAllowed(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "IsAllowed", reflect.TypeOf((*MockNetworkConfig)(nil).IsAllowed), arg0, arg1)
}

// NetworkState mocks base method
func (_m *MockNetworkConfig) NetworkState(_param0 context.Context) protector.NetworkState {
	ret := _m.ctrl.Call(_m, "NetworkState", _param0)
	ret0, _ := ret[0].(protector.NetworkState)
	return ret0
}

// NetworkState indicates an expected call of NetworkState
func (_mr *MockNetworkConfigMockRecorder) NetworkState(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "NetworkState", reflect.TypeOf((*MockNetworkConfig)(nil).NetworkState), arg0)
}

// RemovePeer mocks base method
func (_m *MockNetworkConfig) RemovePeer(_param0 context.Context, _param1 go_libp2p_peer.ID) error {
	ret := _m.ctrl.Call(_m, "RemovePeer", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemovePeer indicates an expected call of RemovePeer
func (_mr *MockNetworkConfigMockRecorder) RemovePeer(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "RemovePeer", reflect.TypeOf((*MockNetworkConfig)(nil).RemovePeer), arg0, arg1)
}

// Reset mocks base method
func (_m *MockNetworkConfig) Reset(_param0 context.Context, _param1 *protector.NetworkConfig) error {
	ret := _m.ctrl.Call(_m, "Reset", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Reset indicates an expected call of Reset
func (_mr *MockNetworkConfigMockRecorder) Reset(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Reset", reflect.TypeOf((*MockNetworkConfig)(nil).Reset), arg0, arg1)
}

// SetNetworkState mocks base method
func (_m *MockNetworkConfig) SetNetworkState(_param0 context.Context, _param1 protector.NetworkState) error {
	ret := _m.ctrl.Call(_m, "SetNetworkState", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetNetworkState indicates an expected call of SetNetworkState
func (_mr *MockNetworkConfigMockRecorder) SetNetworkState(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "SetNetworkState", reflect.TypeOf((*MockNetworkConfig)(nil).SetNetworkState), arg0, arg1)
}
