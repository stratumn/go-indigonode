// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/app/raft/protocol/lib (interfaces: Peer)

// Package mocklib is a generated GoMock package.
package mocklib

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockPeer is a mock of Peer interface
type MockPeer struct {
	ctrl     *gomock.Controller
	recorder *MockPeerMockRecorder
}

// MockPeerMockRecorder is the mock recorder for MockPeer
type MockPeerMockRecorder struct {
	mock *MockPeer
}

// NewMockPeer creates a new mock instance
func NewMockPeer(ctrl *gomock.Controller) *MockPeer {
	mock := &MockPeer{ctrl: ctrl}
	mock.recorder = &MockPeerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPeer) EXPECT() *MockPeerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockPeer) Context() []byte {
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockPeerMockRecorder) Context() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockPeer)(nil).Context))
}

// ID mocks base method
func (m *MockPeer) ID() uint64 {
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// ID indicates an expected call of ID
func (mr *MockPeerMockRecorder) ID() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockPeer)(nil).ID))
}
