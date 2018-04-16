// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/core/protocol/indigo/store (interfaces: NetworkManager)

package mocknetworkmanager

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	store "github.com/stratumn/alice/pb/indigo/store"
	cs "github.com/stratumn/go-indigocore/cs"

	go_libp2p_host "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
)

// MockNetworkManager is a mock of NetworkManager interface
type MockNetworkManager struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkManagerMockRecorder
}

// MockNetworkManagerMockRecorder is the mock recorder for MockNetworkManager
type MockNetworkManagerMockRecorder struct {
	mock *MockNetworkManager
}

// NewMockNetworkManager creates a new mock instance
func NewMockNetworkManager(ctrl *gomock.Controller) *MockNetworkManager {
	mock := &MockNetworkManager{ctrl: ctrl}
	mock.recorder = &MockNetworkManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockNetworkManager) EXPECT() *MockNetworkManagerMockRecorder {
	return _m.recorder
}

// AddListener mocks base method
func (_m *MockNetworkManager) AddListener() <-chan *store.SignedLink {
	ret := _m.ctrl.Call(_m, "AddListener")
	ret0, _ := ret[0].(<-chan *store.SignedLink)
	return ret0
}

// AddListener indicates an expected call of AddListener
func (_mr *MockNetworkManagerMockRecorder) AddListener() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "AddListener", reflect.TypeOf((*MockNetworkManager)(nil).AddListener))
}

// Join mocks base method
func (_m *MockNetworkManager) Join(_param0 context.Context, _param1 string, _param2 go_libp2p_host.Host) error {
	ret := _m.ctrl.Call(_m, "Join", _param0, _param1, _param2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Join indicates an expected call of Join
func (_mr *MockNetworkManagerMockRecorder) Join(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Join", reflect.TypeOf((*MockNetworkManager)(nil).Join), arg0, arg1, arg2)
}

// Leave mocks base method
func (_m *MockNetworkManager) Leave(_param0 context.Context, _param1 string) error {
	ret := _m.ctrl.Call(_m, "Leave", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Leave indicates an expected call of Leave
func (_mr *MockNetworkManagerMockRecorder) Leave(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Leave", reflect.TypeOf((*MockNetworkManager)(nil).Leave), arg0, arg1)
}

// Listen mocks base method
func (_m *MockNetworkManager) Listen(_param0 context.Context) error {
	ret := _m.ctrl.Call(_m, "Listen", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Listen indicates an expected call of Listen
func (_mr *MockNetworkManagerMockRecorder) Listen(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Listen", reflect.TypeOf((*MockNetworkManager)(nil).Listen), arg0)
}

// Publish mocks base method
func (_m *MockNetworkManager) Publish(_param0 context.Context, _param1 *cs.Link) error {
	ret := _m.ctrl.Call(_m, "Publish", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish
func (_mr *MockNetworkManagerMockRecorder) Publish(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Publish", reflect.TypeOf((*MockNetworkManager)(nil).Publish), arg0, arg1)
}

// RemoveListener mocks base method
func (_m *MockNetworkManager) RemoveListener(_param0 <-chan *store.SignedLink) {
	_m.ctrl.Call(_m, "RemoveListener", _param0)
}

// RemoveListener indicates an expected call of RemoveListener
func (_mr *MockNetworkManagerMockRecorder) RemoveListener(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "RemoveListener", reflect.TypeOf((*MockNetworkManager)(nil).RemoveListener), arg0)
}
