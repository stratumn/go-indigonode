// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/app/storage/protocol/p2p (interfaces: Encoder)

// Package mockencoder is a generated GoMock package.
package mockencoder

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockEncoder is a mock of Encoder interface
type MockEncoder struct {
	ctrl     *gomock.Controller
	recorder *MockEncoderMockRecorder
}

// MockEncoderMockRecorder is the mock recorder for MockEncoder
type MockEncoderMockRecorder struct {
	mock *MockEncoder
}

// NewMockEncoder creates a new mock instance
func NewMockEncoder(ctrl *gomock.Controller) *MockEncoder {
	mock := &MockEncoder{ctrl: ctrl}
	mock.recorder = &MockEncoderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEncoder) EXPECT() *MockEncoderMockRecorder {
	return m.recorder
}

// Encode mocks base method
func (m *MockEncoder) Encode(arg0 interface{}) error {
	ret := m.ctrl.Call(m, "Encode", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Encode indicates an expected call of Encode
func (mr *MockEncoderMockRecorder) Encode(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encode", reflect.TypeOf((*MockEncoder)(nil).Encode), arg0)
}
