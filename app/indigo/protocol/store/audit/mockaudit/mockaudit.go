// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit (interfaces: Store)

// Package mockaudit is a generated GoMock package.
package mockaudit

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	cs "github.com/stratumn/go-indigocore/cs"
	audit "github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	go_libp2p_peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	reflect "reflect"
)

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// AddSegment mocks base method
func (m *MockStore) AddSegment(arg0 context.Context, arg1 *cs.Segment) error {
	ret := m.ctrl.Call(m, "AddSegment", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddSegment indicates an expected call of AddSegment
func (mr *MockStoreMockRecorder) AddSegment(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSegment", reflect.TypeOf((*MockStore)(nil).AddSegment), arg0, arg1)
}

// GetByPeer mocks base method
func (m *MockStore) GetByPeer(arg0 context.Context, arg1 go_libp2p_peer.ID, arg2 audit.Pagination) (cs.SegmentSlice, error) {
	ret := m.ctrl.Call(m, "GetByPeer", arg0, arg1, arg2)
	ret0, _ := ret[0].(cs.SegmentSlice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByPeer indicates an expected call of GetByPeer
func (mr *MockStoreMockRecorder) GetByPeer(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPeer", reflect.TypeOf((*MockStore)(nil).GetByPeer), arg0, arg1, arg2)
}
