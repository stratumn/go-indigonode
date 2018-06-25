// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/app/raft/protocol/lib (interfaces: Config)

// Package mocklib is a generated GoMock package.
package mocklib

import (
	gomock "github.com/golang/mock/gomock"
	lib "github.com/stratumn/go-indigonode/app/raft/protocol/lib"
	reflect "reflect"
)

// MockConfig is a mock of Config interface
type MockConfig struct {
	ctrl     *gomock.Controller
	recorder *MockConfigMockRecorder
}

// MockConfigMockRecorder is the mock recorder for MockConfig
type MockConfigMockRecorder struct {
	mock *MockConfig
}

// NewMockConfig creates a new mock instance
func NewMockConfig(ctrl *gomock.Controller) *MockConfig {
	mock := &MockConfig{ctrl: ctrl}
	mock.recorder = &MockConfigMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConfig) EXPECT() *MockConfigMockRecorder {
	return m.recorder
}

// ElectionTick mocks base method
func (m *MockConfig) ElectionTick() int {
	ret := m.ctrl.Call(m, "ElectionTick")
	ret0, _ := ret[0].(int)
	return ret0
}

// ElectionTick indicates an expected call of ElectionTick
func (mr *MockConfigMockRecorder) ElectionTick() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ElectionTick", reflect.TypeOf((*MockConfig)(nil).ElectionTick))
}

// HeartbeatTick mocks base method
func (m *MockConfig) HeartbeatTick() int {
	ret := m.ctrl.Call(m, "HeartbeatTick")
	ret0, _ := ret[0].(int)
	return ret0
}

// HeartbeatTick indicates an expected call of HeartbeatTick
func (mr *MockConfigMockRecorder) HeartbeatTick() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HeartbeatTick", reflect.TypeOf((*MockConfig)(nil).HeartbeatTick))
}

// ID mocks base method
func (m *MockConfig) ID() uint64 {
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// ID indicates an expected call of ID
func (mr *MockConfigMockRecorder) ID() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockConfig)(nil).ID))
}

// MaxInflightMsgs mocks base method
func (m *MockConfig) MaxInflightMsgs() int {
	ret := m.ctrl.Call(m, "MaxInflightMsgs")
	ret0, _ := ret[0].(int)
	return ret0
}

// MaxInflightMsgs indicates an expected call of MaxInflightMsgs
func (mr *MockConfigMockRecorder) MaxInflightMsgs() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxInflightMsgs", reflect.TypeOf((*MockConfig)(nil).MaxInflightMsgs))
}

// MaxSizePerMsg mocks base method
func (m *MockConfig) MaxSizePerMsg() uint64 {
	ret := m.ctrl.Call(m, "MaxSizePerMsg")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// MaxSizePerMsg indicates an expected call of MaxSizePerMsg
func (mr *MockConfigMockRecorder) MaxSizePerMsg() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxSizePerMsg", reflect.TypeOf((*MockConfig)(nil).MaxSizePerMsg))
}

// Storage mocks base method
func (m *MockConfig) Storage() lib.Storage {
	ret := m.ctrl.Call(m, "Storage")
	ret0, _ := ret[0].(lib.Storage)
	return ret0
}

// Storage indicates an expected call of Storage
func (mr *MockConfigMockRecorder) Storage() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Storage", reflect.TypeOf((*MockConfig)(nil).Storage))
}
