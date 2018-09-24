// Code generated by MockGen. DO NOT EDIT.
// Source: gx/ipfs/QmU129xU8dM79BgR97hu4fsiUDkTQrNHbzkiYfyrkNci8o/go-libp2p-transport (interfaces: Conn)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	go_libp2p_crypto "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	go_libp2p_peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	go_libp2p_transport "gx/ipfs/QmU129xU8dM79BgR97hu4fsiUDkTQrNHbzkiYfyrkNci8o/go-libp2p-transport"
	go_stream_muxer "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	go_multiaddr "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	reflect "reflect"
)

// MockTransportConn is a mock of Conn interface
type MockTransportConn struct {
	ctrl     *gomock.Controller
	recorder *MockTransportConnMockRecorder
}

// MockTransportConnMockRecorder is the mock recorder for MockTransportConn
type MockTransportConnMockRecorder struct {
	mock *MockTransportConn
}

// NewMockTransportConn creates a new mock instance
func NewMockTransportConn(ctrl *gomock.Controller) *MockTransportConn {
	mock := &MockTransportConn{ctrl: ctrl}
	mock.recorder = &MockTransportConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTransportConn) EXPECT() *MockTransportConnMockRecorder {
	return m.recorder
}

// AcceptStream mocks base method
func (m *MockTransportConn) AcceptStream() (go_stream_muxer.Stream, error) {
	ret := m.ctrl.Call(m, "AcceptStream")
	ret0, _ := ret[0].(go_stream_muxer.Stream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AcceptStream indicates an expected call of AcceptStream
func (mr *MockTransportConnMockRecorder) AcceptStream() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptStream", reflect.TypeOf((*MockTransportConn)(nil).AcceptStream))
}

// Close mocks base method
func (m *MockTransportConn) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockTransportConnMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockTransportConn)(nil).Close))
}

// IsClosed mocks base method
func (m *MockTransportConn) IsClosed() bool {
	ret := m.ctrl.Call(m, "IsClosed")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsClosed indicates an expected call of IsClosed
func (mr *MockTransportConnMockRecorder) IsClosed() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsClosed", reflect.TypeOf((*MockTransportConn)(nil).IsClosed))
}

// LocalMultiaddr mocks base method
func (m *MockTransportConn) LocalMultiaddr() go_multiaddr.Multiaddr {
	ret := m.ctrl.Call(m, "LocalMultiaddr")
	ret0, _ := ret[0].(go_multiaddr.Multiaddr)
	return ret0
}

// LocalMultiaddr indicates an expected call of LocalMultiaddr
func (mr *MockTransportConnMockRecorder) LocalMultiaddr() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalMultiaddr", reflect.TypeOf((*MockTransportConn)(nil).LocalMultiaddr))
}

// LocalPeer mocks base method
func (m *MockTransportConn) LocalPeer() go_libp2p_peer.ID {
	ret := m.ctrl.Call(m, "LocalPeer")
	ret0, _ := ret[0].(go_libp2p_peer.ID)
	return ret0
}

// LocalPeer indicates an expected call of LocalPeer
func (mr *MockTransportConnMockRecorder) LocalPeer() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalPeer", reflect.TypeOf((*MockTransportConn)(nil).LocalPeer))
}

// LocalPrivateKey mocks base method
func (m *MockTransportConn) LocalPrivateKey() go_libp2p_crypto.PrivKey {
	ret := m.ctrl.Call(m, "LocalPrivateKey")
	ret0, _ := ret[0].(go_libp2p_crypto.PrivKey)
	return ret0
}

// LocalPrivateKey indicates an expected call of LocalPrivateKey
func (mr *MockTransportConnMockRecorder) LocalPrivateKey() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LocalPrivateKey", reflect.TypeOf((*MockTransportConn)(nil).LocalPrivateKey))
}

// OpenStream mocks base method
func (m *MockTransportConn) OpenStream() (go_stream_muxer.Stream, error) {
	ret := m.ctrl.Call(m, "OpenStream")
	ret0, _ := ret[0].(go_stream_muxer.Stream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenStream indicates an expected call of OpenStream
func (mr *MockTransportConnMockRecorder) OpenStream() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenStream", reflect.TypeOf((*MockTransportConn)(nil).OpenStream))
}

// RemoteMultiaddr mocks base method
func (m *MockTransportConn) RemoteMultiaddr() go_multiaddr.Multiaddr {
	ret := m.ctrl.Call(m, "RemoteMultiaddr")
	ret0, _ := ret[0].(go_multiaddr.Multiaddr)
	return ret0
}

// RemoteMultiaddr indicates an expected call of RemoteMultiaddr
func (mr *MockTransportConnMockRecorder) RemoteMultiaddr() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteMultiaddr", reflect.TypeOf((*MockTransportConn)(nil).RemoteMultiaddr))
}

// RemotePeer mocks base method
func (m *MockTransportConn) RemotePeer() go_libp2p_peer.ID {
	ret := m.ctrl.Call(m, "RemotePeer")
	ret0, _ := ret[0].(go_libp2p_peer.ID)
	return ret0
}

// RemotePeer indicates an expected call of RemotePeer
func (mr *MockTransportConnMockRecorder) RemotePeer() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemotePeer", reflect.TypeOf((*MockTransportConn)(nil).RemotePeer))
}

// RemotePublicKey mocks base method
func (m *MockTransportConn) RemotePublicKey() go_libp2p_crypto.PubKey {
	ret := m.ctrl.Call(m, "RemotePublicKey")
	ret0, _ := ret[0].(go_libp2p_crypto.PubKey)
	return ret0
}

// RemotePublicKey indicates an expected call of RemotePublicKey
func (mr *MockTransportConnMockRecorder) RemotePublicKey() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemotePublicKey", reflect.TypeOf((*MockTransportConn)(nil).RemotePublicKey))
}

// Transport mocks base method
func (m *MockTransportConn) Transport() go_libp2p_transport.Transport {
	ret := m.ctrl.Call(m, "Transport")
	ret0, _ := ret[0].(go_libp2p_transport.Transport)
	return ret0
}

// Transport indicates an expected call of Transport
func (mr *MockTransportConnMockRecorder) Transport() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transport", reflect.TypeOf((*MockTransportConn)(nil).Transport))
}
