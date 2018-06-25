// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/stratumn/go-indigonode/app/coin/protocol/p2p (interfaces: P2P)

// Package mockp2p is a generated GoMock package.
package mockp2p

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	pb "github.com/stratumn/go-indigonode/app/coin/pb"
	chain "github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	p2p "github.com/stratumn/go-indigonode/app/coin/protocol/p2p"
	go_libp2p_peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	reflect "reflect"
)

// MockP2P is a mock of P2P interface
type MockP2P struct {
	ctrl     *gomock.Controller
	recorder *MockP2PMockRecorder
}

// MockP2PMockRecorder is the mock recorder for MockP2P
type MockP2PMockRecorder struct {
	mock *MockP2P
}

// NewMockP2P creates a new mock instance
func NewMockP2P(ctrl *gomock.Controller) *MockP2P {
	mock := &MockP2P{ctrl: ctrl}
	mock.recorder = &MockP2PMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockP2P) EXPECT() *MockP2PMockRecorder {
	return m.recorder
}

// RequestBlockByHash mocks base method
func (m *MockP2P) RequestBlockByHash(arg0 context.Context, arg1 go_libp2p_peer.ID, arg2 []byte) (*pb.Block, error) {
	ret := m.ctrl.Call(m, "RequestBlockByHash", arg0, arg1, arg2)
	ret0, _ := ret[0].(*pb.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestBlockByHash indicates an expected call of RequestBlockByHash
func (mr *MockP2PMockRecorder) RequestBlockByHash(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestBlockByHash", reflect.TypeOf((*MockP2P)(nil).RequestBlockByHash), arg0, arg1, arg2)
}

// RequestBlocksByNumber mocks base method
func (m *MockP2P) RequestBlocksByNumber(arg0 context.Context, arg1 go_libp2p_peer.ID, arg2, arg3 uint64) ([]*pb.Block, error) {
	ret := m.ctrl.Call(m, "RequestBlocksByNumber", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*pb.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestBlocksByNumber indicates an expected call of RequestBlocksByNumber
func (mr *MockP2PMockRecorder) RequestBlocksByNumber(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestBlocksByNumber", reflect.TypeOf((*MockP2P)(nil).RequestBlocksByNumber), arg0, arg1, arg2, arg3)
}

// RequestHeaderByHash mocks base method
func (m *MockP2P) RequestHeaderByHash(arg0 context.Context, arg1 go_libp2p_peer.ID, arg2 []byte) (*pb.Header, error) {
	ret := m.ctrl.Call(m, "RequestHeaderByHash", arg0, arg1, arg2)
	ret0, _ := ret[0].(*pb.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestHeaderByHash indicates an expected call of RequestHeaderByHash
func (mr *MockP2PMockRecorder) RequestHeaderByHash(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestHeaderByHash", reflect.TypeOf((*MockP2P)(nil).RequestHeaderByHash), arg0, arg1, arg2)
}

// RequestHeadersByNumber mocks base method
func (m *MockP2P) RequestHeadersByNumber(arg0 context.Context, arg1 go_libp2p_peer.ID, arg2, arg3 uint64) ([]*pb.Header, error) {
	ret := m.ctrl.Call(m, "RequestHeadersByNumber", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*pb.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestHeadersByNumber indicates an expected call of RequestHeadersByNumber
func (mr *MockP2PMockRecorder) RequestHeadersByNumber(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestHeadersByNumber", reflect.TypeOf((*MockP2P)(nil).RequestHeadersByNumber), arg0, arg1, arg2, arg3)
}

// RespondBlockByHash mocks base method
func (m *MockP2P) RespondBlockByHash(arg0 context.Context, arg1 *pb.BlockRequest, arg2 p2p.Encoder, arg3 chain.Reader) error {
	ret := m.ctrl.Call(m, "RespondBlockByHash", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondBlockByHash indicates an expected call of RespondBlockByHash
func (mr *MockP2PMockRecorder) RespondBlockByHash(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondBlockByHash", reflect.TypeOf((*MockP2P)(nil).RespondBlockByHash), arg0, arg1, arg2, arg3)
}

// RespondBlocksByNumber mocks base method
func (m *MockP2P) RespondBlocksByNumber(arg0 context.Context, arg1 *pb.BlocksRequest, arg2 p2p.Encoder, arg3 chain.Reader) error {
	ret := m.ctrl.Call(m, "RespondBlocksByNumber", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondBlocksByNumber indicates an expected call of RespondBlocksByNumber
func (mr *MockP2PMockRecorder) RespondBlocksByNumber(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondBlocksByNumber", reflect.TypeOf((*MockP2P)(nil).RespondBlocksByNumber), arg0, arg1, arg2, arg3)
}

// RespondHeaderByHash mocks base method
func (m *MockP2P) RespondHeaderByHash(arg0 context.Context, arg1 *pb.HeaderRequest, arg2 p2p.Encoder, arg3 chain.Reader) error {
	ret := m.ctrl.Call(m, "RespondHeaderByHash", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondHeaderByHash indicates an expected call of RespondHeaderByHash
func (mr *MockP2PMockRecorder) RespondHeaderByHash(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondHeaderByHash", reflect.TypeOf((*MockP2P)(nil).RespondHeaderByHash), arg0, arg1, arg2, arg3)
}

// RespondHeadersByNumber mocks base method
func (m *MockP2P) RespondHeadersByNumber(arg0 context.Context, arg1 *pb.HeadersRequest, arg2 p2p.Encoder, arg3 chain.Reader) error {
	ret := m.ctrl.Call(m, "RespondHeadersByNumber", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondHeadersByNumber indicates an expected call of RespondHeadersByNumber
func (mr *MockP2PMockRecorder) RespondHeadersByNumber(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondHeadersByNumber", reflect.TypeOf((*MockP2P)(nil).RespondHeadersByNumber), arg0, arg1, arg2, arg3)
}
