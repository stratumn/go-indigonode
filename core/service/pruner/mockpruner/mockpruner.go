// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/core/service/pruner (interfaces: Manager)

package mockpruner

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockManager is a mock of Manager interface
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockManager) EXPECT() *MockManagerMockRecorder {
	return _m.recorder
}

// Prune mocks base method
func (_m *MockManager) Prune() {
	_m.ctrl.Call(_m, "Prune")
}

// Prune indicates an expected call of Prune
func (_mr *MockManagerMockRecorder) Prune() *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Prune", reflect.TypeOf((*MockManager)(nil).Prune))
}
