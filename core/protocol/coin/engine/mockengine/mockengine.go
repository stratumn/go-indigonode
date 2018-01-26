// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/core/protocol/coin/engine (interfaces: Engine)

// Package mockengine is a generated GoMock package.
package mockengine

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	chain "github.com/stratumn/alice/core/protocol/coin/chain"
	state "github.com/stratumn/alice/core/protocol/coin/state"
	coin "github.com/stratumn/alice/pb/coin"
	reflect "reflect"
)

// MockEngine is a mock of Engine interface
type MockEngine struct {
	ctrl     *gomock.Controller
	recorder *MockEngineMockRecorder
}

// MockEngineMockRecorder is the mock recorder for MockEngine
type MockEngineMockRecorder struct {
	mock *MockEngine
}

// NewMockEngine creates a new mock instance
func NewMockEngine(ctrl *gomock.Controller) *MockEngine {
	mock := &MockEngine{ctrl: ctrl}
	mock.recorder = &MockEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEngine) EXPECT() *MockEngineMockRecorder {
	return m.recorder
}

// Finalize mocks base method
func (m *MockEngine) Finalize(arg0 chain.Reader, arg1 *coin.Header, arg2 *state.State, arg3 []*coin.Transaction) (*coin.Block, error) {
	ret := m.ctrl.Call(m, "Finalize", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*coin.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Finalize indicates an expected call of Finalize
func (mr *MockEngineMockRecorder) Finalize(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Finalize", reflect.TypeOf((*MockEngine)(nil).Finalize), arg0, arg1, arg2, arg3)
}

// Prepare mocks base method
func (m *MockEngine) Prepare(arg0 chain.Reader, arg1 *coin.Header) error {
	ret := m.ctrl.Call(m, "Prepare", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Prepare indicates an expected call of Prepare
func (mr *MockEngineMockRecorder) Prepare(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prepare", reflect.TypeOf((*MockEngine)(nil).Prepare), arg0, arg1)
}

// VerifyHeader mocks base method
func (m *MockEngine) VerifyHeader(arg0 chain.Reader, arg1 *coin.Header) error {
	ret := m.ctrl.Call(m, "VerifyHeader", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyHeader indicates an expected call of VerifyHeader
func (mr *MockEngineMockRecorder) VerifyHeader(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyHeader", reflect.TypeOf((*MockEngine)(nil).VerifyHeader), arg0, arg1)
}

// VerifyHeaders mocks base method
func (m *MockEngine) VerifyHeaders(arg0 context.Context, arg1 chain.Reader, arg2 []*coin.Header) <-chan error {
	ret := m.ctrl.Call(m, "VerifyHeaders", arg0, arg1, arg2)
	ret0, _ := ret[0].(<-chan error)
	return ret0
}

// VerifyHeaders indicates an expected call of VerifyHeaders
func (mr *MockEngineMockRecorder) VerifyHeaders(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyHeaders", reflect.TypeOf((*MockEngine)(nil).VerifyHeaders), arg0, arg1, arg2)
}