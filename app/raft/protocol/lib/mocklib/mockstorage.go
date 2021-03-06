// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-node/app/raft/protocol/lib (interfaces: Storage)

// Package mocklib is a generated GoMock package.
package mocklib

import (
	gomock "github.com/golang/mock/gomock"
	lib "github.com/stratumn/go-node/app/raft/protocol/lib"
	reflect "reflect"
)

// MockStorage is a mock of Storage interface
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// Append mocks base method
func (m *MockStorage) Append(arg0 []lib.Entry) error {
	ret := m.ctrl.Call(m, "Append", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Append indicates an expected call of Append
func (mr *MockStorageMockRecorder) Append(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Append", reflect.TypeOf((*MockStorage)(nil).Append), arg0)
}

// Entries mocks base method
func (m *MockStorage) Entries(arg0, arg1, arg2 uint64) ([]lib.Entry, error) {
	ret := m.ctrl.Call(m, "Entries", arg0, arg1, arg2)
	ret0, _ := ret[0].([]lib.Entry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Entries indicates an expected call of Entries
func (mr *MockStorageMockRecorder) Entries(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Entries", reflect.TypeOf((*MockStorage)(nil).Entries), arg0, arg1, arg2)
}

// FirstIndex mocks base method
func (m *MockStorage) FirstIndex() (uint64, error) {
	ret := m.ctrl.Call(m, "FirstIndex")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FirstIndex indicates an expected call of FirstIndex
func (mr *MockStorageMockRecorder) FirstIndex() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FirstIndex", reflect.TypeOf((*MockStorage)(nil).FirstIndex))
}

// InitialState mocks base method
func (m *MockStorage) InitialState() (lib.HardState, lib.ConfState, error) {
	ret := m.ctrl.Call(m, "InitialState")
	ret0, _ := ret[0].(lib.HardState)
	ret1, _ := ret[1].(lib.ConfState)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// InitialState indicates an expected call of InitialState
func (mr *MockStorageMockRecorder) InitialState() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitialState", reflect.TypeOf((*MockStorage)(nil).InitialState))
}

// LastIndex mocks base method
func (m *MockStorage) LastIndex() (uint64, error) {
	ret := m.ctrl.Call(m, "LastIndex")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastIndex indicates an expected call of LastIndex
func (mr *MockStorageMockRecorder) LastIndex() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastIndex", reflect.TypeOf((*MockStorage)(nil).LastIndex))
}

// Snapshot mocks base method
func (m *MockStorage) Snapshot() (lib.Snapshot, error) {
	ret := m.ctrl.Call(m, "Snapshot")
	ret0, _ := ret[0].(lib.Snapshot)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Snapshot indicates an expected call of Snapshot
func (mr *MockStorageMockRecorder) Snapshot() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Snapshot", reflect.TypeOf((*MockStorage)(nil).Snapshot))
}

// Term mocks base method
func (m *MockStorage) Term(arg0 uint64) (uint64, error) {
	ret := m.ctrl.Call(m, "Term", arg0)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Term indicates an expected call of Term
func (mr *MockStorageMockRecorder) Term(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Term", reflect.TypeOf((*MockStorage)(nil).Term), arg0)
}
