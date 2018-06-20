// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/alice/core/streamutil (interfaces: Provider)

// Package mockstream is a generated GoMock package.
package mockstream

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	streamutil "github.com/stratumn/alice/core/streamutil"
	go_libp2p_host "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
	reflect "reflect"
)

// MockProvider is a mock of Provider interface
type MockProvider struct {
	ctrl     *gomock.Controller
	recorder *MockProviderMockRecorder
}

// MockProviderMockRecorder is the mock recorder for MockProvider
type MockProviderMockRecorder struct {
	mock *MockProvider
}

// NewMockProvider creates a new mock instance
func NewMockProvider(ctrl *gomock.Controller) *MockProvider {
	mock := &MockProvider{ctrl: ctrl}
	mock.recorder = &MockProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProvider) EXPECT() *MockProviderMockRecorder {
	return m.recorder
}

// NewStream mocks base method
func (m *MockProvider) NewStream(arg0 context.Context, arg1 go_libp2p_host.Host, arg2 ...streamutil.StreamOption) (streamutil.Stream, error) {
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NewStream", varargs...)
	ret0, _ := ret[0].(streamutil.Stream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewStream indicates an expected call of NewStream
func (mr *MockProviderMockRecorder) NewStream(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewStream", reflect.TypeOf((*MockProvider)(nil).NewStream), varargs...)
}