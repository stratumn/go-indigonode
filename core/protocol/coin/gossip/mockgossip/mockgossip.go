// Code generated by MockGen. DO NOT EDIT.
// Source: gossip.go

// Package mockgossip is a generated GoMock package.
package mockgossip

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	coin "github.com/stratumn/alice/pb/coin"
)

// MockGossip is a mock of Gossip interface
type MockGossip struct {
	ctrl     *gomock.Controller
	recorder *MockGossipMockRecorder
}

// MockGossipMockRecorder is the mock recorder for MockGossip
type MockGossipMockRecorder struct {
	mock *MockGossip
}

// NewMockGossip creates a new mock instance
func NewMockGossip(ctrl *gomock.Controller) *MockGossip {
	mock := &MockGossip{ctrl: ctrl}
	mock.recorder = &MockGossipMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGossip) EXPECT() *MockGossipMockRecorder {
	return m.recorder
}

// ListenTx mocks base method
func (m *MockGossip) ListenTx(ctx context.Context, processTx func(*coin.Transaction) error) error {
	ret := m.ctrl.Call(m, "ListenTx", ctx, processTx)
	ret0, _ := ret[0].(error)
	return ret0
}

// ListenTx indicates an expected call of ListenTx
func (mr *MockGossipMockRecorder) ListenTx(ctx, processTx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenTx", reflect.TypeOf((*MockGossip)(nil).ListenTx), ctx, processTx)
}

// ListenBlock mocks base method
func (m *MockGossip) ListenBlock(ctx context.Context, processBlock func(*coin.Block) error) error {
	ret := m.ctrl.Call(m, "ListenBlock", ctx, processBlock)
	ret0, _ := ret[0].(error)
	return ret0
}

// ListenBlock indicates an expected call of ListenBlock
func (mr *MockGossipMockRecorder) ListenBlock(ctx, processBlock interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListenBlock", reflect.TypeOf((*MockGossip)(nil).ListenBlock), ctx, processBlock)
}

// PublishTx mocks base method
func (m *MockGossip) PublishTx(tx *coin.Transaction) error {
	ret := m.ctrl.Call(m, "PublishTx", tx)
	ret0, _ := ret[0].(error)
	return ret0
}

// PublishTx indicates an expected call of PublishTx
func (mr *MockGossipMockRecorder) PublishTx(tx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishTx", reflect.TypeOf((*MockGossip)(nil).PublishTx), tx)
}

// PublishBlock mocks base method
func (m *MockGossip) PublishBlock(block *coin.Block) error {
	ret := m.ctrl.Call(m, "PublishBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// PublishBlock indicates an expected call of PublishBlock
func (mr *MockGossipMockRecorder) PublishBlock(block interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishBlock", reflect.TypeOf((*MockGossip)(nil).PublishBlock), block)
}