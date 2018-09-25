// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-node/app/raft/protocol/lib (interfaces: Lib)

// Package mocklib is a generated GoMock package.
package mocklib

import (
	gomock "github.com/golang/mock/gomock"
	lib "github.com/stratumn/go-node/app/raft/protocol/lib"
	reflect "reflect"
)

// MockLib is a mock of Lib interface
type MockLib struct {
	ctrl     *gomock.Controller
	recorder *MockLibMockRecorder
}

// MockLibMockRecorder is the mock recorder for MockLib
type MockLibMockRecorder struct {
	mock *MockLib
}

// NewMockLib creates a new mock instance
func NewMockLib(ctrl *gomock.Controller) *MockLib {
	mock := &MockLib{ctrl: ctrl}
	mock.recorder = &MockLibMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLib) EXPECT() *MockLibMockRecorder {
	return m.recorder
}

// NewConfChange mocks base method
func (m *MockLib) NewConfChange(arg0 uint64, arg1 lib.ConfChangeType, arg2 uint64, arg3 []byte) lib.ConfChange {
	ret := m.ctrl.Call(m, "NewConfChange", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(lib.ConfChange)
	return ret0
}

// NewConfChange indicates an expected call of NewConfChange
func (mr *MockLibMockRecorder) NewConfChange(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewConfChange", reflect.TypeOf((*MockLib)(nil).NewConfChange), arg0, arg1, arg2, arg3)
}

// NewConfig mocks base method
func (m *MockLib) NewConfig(arg0 uint64, arg1, arg2 int, arg3 lib.Storage, arg4 uint64, arg5 int) lib.Config {
	ret := m.ctrl.Call(m, "NewConfig", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(lib.Config)
	return ret0
}

// NewConfig indicates an expected call of NewConfig
func (mr *MockLibMockRecorder) NewConfig(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewConfig", reflect.TypeOf((*MockLib)(nil).NewConfig), arg0, arg1, arg2, arg3, arg4, arg5)
}

// NewMemoryStorage mocks base method
func (m *MockLib) NewMemoryStorage() lib.Storage {
	ret := m.ctrl.Call(m, "NewMemoryStorage")
	ret0, _ := ret[0].(lib.Storage)
	return ret0
}

// NewMemoryStorage indicates an expected call of NewMemoryStorage
func (mr *MockLibMockRecorder) NewMemoryStorage() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewMemoryStorage", reflect.TypeOf((*MockLib)(nil).NewMemoryStorage))
}

// NewPeer mocks base method
func (m *MockLib) NewPeer(arg0 uint64, arg1 []byte) lib.Peer {
	ret := m.ctrl.Call(m, "NewPeer", arg0, arg1)
	ret0, _ := ret[0].(lib.Peer)
	return ret0
}

// NewPeer indicates an expected call of NewPeer
func (mr *MockLibMockRecorder) NewPeer(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewPeer", reflect.TypeOf((*MockLib)(nil).NewPeer), arg0, arg1)
}

// StartNode mocks base method
func (m *MockLib) StartNode(arg0 lib.Config, arg1 []lib.Peer) lib.Node {
	ret := m.ctrl.Call(m, "StartNode", arg0, arg1)
	ret0, _ := ret[0].(lib.Node)
	return ret0
}

// StartNode indicates an expected call of StartNode
func (mr *MockLibMockRecorder) StartNode(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartNode", reflect.TypeOf((*MockLib)(nil).StartNode), arg0, arg1)
}
