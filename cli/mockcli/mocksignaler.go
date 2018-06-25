// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/cli (interfaces: Signaler)

// Package mockcli is a generated GoMock package.
package mockcli

import (
	gomock "github.com/golang/mock/gomock"
	os "os"
	reflect "reflect"
)

// MockSignaler is a mock of Signaler interface
type MockSignaler struct {
	ctrl     *gomock.Controller
	recorder *MockSignalerMockRecorder
}

// MockSignalerMockRecorder is the mock recorder for MockSignaler
type MockSignalerMockRecorder struct {
	mock *MockSignaler
}

// NewMockSignaler creates a new mock instance
func NewMockSignaler(ctrl *gomock.Controller) *MockSignaler {
	mock := &MockSignaler{ctrl: ctrl}
	mock.recorder = &MockSignalerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSignaler) EXPECT() *MockSignalerMockRecorder {
	return m.recorder
}

// Signal mocks base method
func (m *MockSignaler) Signal(arg0 os.Signal) error {
	ret := m.ctrl.Call(m, "Signal", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Signal indicates an expected call of Signal
func (mr *MockSignalerMockRecorder) Signal(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Signal", reflect.TypeOf((*MockSignaler)(nil).Signal), arg0)
}
