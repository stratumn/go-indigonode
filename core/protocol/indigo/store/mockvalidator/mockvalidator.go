// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigocore/validator (interfaces: Validator)

// Package mockvalidator is a generated GoMock package.
package mockvalidator

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	cs "github.com/stratumn/go-indigocore/cs"
	store "github.com/stratumn/go-indigocore/store"
	types "github.com/stratumn/go-indigocore/types"
	reflect "reflect"
)

// MockValidator is a mock of Validator interface
type MockValidator struct {
	ctrl     *gomock.Controller
	recorder *MockValidatorMockRecorder
}

// MockValidatorMockRecorder is the mock recorder for MockValidator
type MockValidatorMockRecorder struct {
	mock *MockValidator
}

// NewMockValidator creates a new mock instance
func NewMockValidator(ctrl *gomock.Controller) *MockValidator {
	mock := &MockValidator{ctrl: ctrl}
	mock.recorder = &MockValidatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockValidator) EXPECT() *MockValidatorMockRecorder {
	return m.recorder
}

// Hash mocks base method
func (m *MockValidator) Hash() (*types.Bytes32, error) {
	ret := m.ctrl.Call(m, "Hash")
	ret0, _ := ret[0].(*types.Bytes32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Hash indicates an expected call of Hash
func (mr *MockValidatorMockRecorder) Hash() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Hash", reflect.TypeOf((*MockValidator)(nil).Hash))
}

// ShouldValidate mocks base method
func (m *MockValidator) ShouldValidate(arg0 *cs.Link) bool {
	ret := m.ctrl.Call(m, "ShouldValidate", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ShouldValidate indicates an expected call of ShouldValidate
func (mr *MockValidatorMockRecorder) ShouldValidate(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShouldValidate", reflect.TypeOf((*MockValidator)(nil).ShouldValidate), arg0)
}

// Validate mocks base method
func (m *MockValidator) Validate(arg0 context.Context, arg1 store.SegmentReader, arg2 *cs.Link) error {
	ret := m.ctrl.Call(m, "Validate", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockValidatorMockRecorder) Validate(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockValidator)(nil).Validate), arg0, arg1, arg2)
}
