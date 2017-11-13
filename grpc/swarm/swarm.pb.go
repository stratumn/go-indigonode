// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/stratumn/alice/grpc/swarm/swarm.proto

/*
Package swarm is a generated protocol buffer package.

It is generated from these files:
	github.com/stratumn/alice/grpc/swarm/swarm.proto

It has these top-level messages:
	LocalPeerReq
	PeersReq
	ConnectionsReq
	Peer
	Connection
*/
package swarm

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/grpc/ext"

import (
	context "context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// The local peer request message.
type LocalPeerReq struct {
}

func (m *LocalPeerReq) Reset()                    { *m = LocalPeerReq{} }
func (m *LocalPeerReq) String() string            { return proto.CompactTextString(m) }
func (*LocalPeerReq) ProtoMessage()               {}
func (*LocalPeerReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// The peers request message.
type PeersReq struct {
}

func (m *PeersReq) Reset()                    { *m = PeersReq{} }
func (m *PeersReq) String() string            { return proto.CompactTextString(m) }
func (*PeersReq) ProtoMessage()               {}
func (*PeersReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// The connections request message.
type ConnectionsReq struct {
	PeerId []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
}

func (m *ConnectionsReq) Reset()                    { *m = ConnectionsReq{} }
func (m *ConnectionsReq) String() string            { return proto.CompactTextString(m) }
func (*ConnectionsReq) ProtoMessage()               {}
func (*ConnectionsReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ConnectionsReq) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

// The peer message containing the ID of the peer.
type Peer struct {
	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *Peer) Reset()                    { *m = Peer{} }
func (m *Peer) String() string            { return proto.CompactTextString(m) }
func (*Peer) ProtoMessage()               {}
func (*Peer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Peer) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

// The connection message containing the peer ID and addresses.
type Connection struct {
	PeerId        []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	LocalAddress  []byte `protobuf:"bytes,2,opt,name=local_address,json=localAddress,proto3" json:"local_address,omitempty"`
	RemoteAddress []byte `protobuf:"bytes,3,opt,name=remote_address,json=remoteAddress,proto3" json:"remote_address,omitempty"`
}

func (m *Connection) Reset()                    { *m = Connection{} }
func (m *Connection) String() string            { return proto.CompactTextString(m) }
func (*Connection) ProtoMessage()               {}
func (*Connection) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Connection) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

func (m *Connection) GetLocalAddress() []byte {
	if m != nil {
		return m.LocalAddress
	}
	return nil
}

func (m *Connection) GetRemoteAddress() []byte {
	if m != nil {
		return m.RemoteAddress
	}
	return nil
}

func init() {
	proto.RegisterType((*LocalPeerReq)(nil), "stratumn.alice.grpc.swarm.LocalPeerReq")
	proto.RegisterType((*PeersReq)(nil), "stratumn.alice.grpc.swarm.PeersReq")
	proto.RegisterType((*ConnectionsReq)(nil), "stratumn.alice.grpc.swarm.ConnectionsReq")
	proto.RegisterType((*Peer)(nil), "stratumn.alice.grpc.swarm.Peer")
	proto.RegisterType((*Connection)(nil), "stratumn.alice.grpc.swarm.Connection")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Swarm service

type SwarmClient interface {
	// Returns the local peer.
	LocalPeer(ctx context.Context, in *LocalPeerReq, opts ...grpc.CallOption) (*Peer, error)
	// Streams the connected peers.
	Peers(ctx context.Context, in *PeersReq, opts ...grpc.CallOption) (Swarm_PeersClient, error)
	// Streams connections.
	Connections(ctx context.Context, in *ConnectionsReq, opts ...grpc.CallOption) (Swarm_ConnectionsClient, error)
}

type swarmClient struct {
	cc *grpc.ClientConn
}

func NewSwarmClient(cc *grpc.ClientConn) SwarmClient {
	return &swarmClient{cc}
}

func (c *swarmClient) LocalPeer(ctx context.Context, in *LocalPeerReq, opts ...grpc.CallOption) (*Peer, error) {
	out := new(Peer)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.swarm.Swarm/LocalPeer", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *swarmClient) Peers(ctx context.Context, in *PeersReq, opts ...grpc.CallOption) (Swarm_PeersClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Swarm_serviceDesc.Streams[0], c.cc, "/stratumn.alice.grpc.swarm.Swarm/Peers", opts...)
	if err != nil {
		return nil, err
	}
	x := &swarmPeersClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Swarm_PeersClient interface {
	Recv() (*Peer, error)
	grpc.ClientStream
}

type swarmPeersClient struct {
	grpc.ClientStream
}

func (x *swarmPeersClient) Recv() (*Peer, error) {
	m := new(Peer)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *swarmClient) Connections(ctx context.Context, in *ConnectionsReq, opts ...grpc.CallOption) (Swarm_ConnectionsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Swarm_serviceDesc.Streams[1], c.cc, "/stratumn.alice.grpc.swarm.Swarm/Connections", opts...)
	if err != nil {
		return nil, err
	}
	x := &swarmConnectionsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Swarm_ConnectionsClient interface {
	Recv() (*Connection, error)
	grpc.ClientStream
}

type swarmConnectionsClient struct {
	grpc.ClientStream
}

func (x *swarmConnectionsClient) Recv() (*Connection, error) {
	m := new(Connection)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Swarm service

type SwarmServer interface {
	// Returns the local peer.
	LocalPeer(context.Context, *LocalPeerReq) (*Peer, error)
	// Streams the connected peers.
	Peers(*PeersReq, Swarm_PeersServer) error
	// Streams connections.
	Connections(*ConnectionsReq, Swarm_ConnectionsServer) error
}

func RegisterSwarmServer(s *grpc.Server, srv SwarmServer) {
	s.RegisterService(&_Swarm_serviceDesc, srv)
}

func _Swarm_LocalPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocalPeerReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SwarmServer).LocalPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.swarm.Swarm/LocalPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SwarmServer).LocalPeer(ctx, req.(*LocalPeerReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Swarm_Peers_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PeersReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SwarmServer).Peers(m, &swarmPeersServer{stream})
}

type Swarm_PeersServer interface {
	Send(*Peer) error
	grpc.ServerStream
}

type swarmPeersServer struct {
	grpc.ServerStream
}

func (x *swarmPeersServer) Send(m *Peer) error {
	return x.ServerStream.SendMsg(m)
}

func _Swarm_Connections_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConnectionsReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SwarmServer).Connections(m, &swarmConnectionsServer{stream})
}

type Swarm_ConnectionsServer interface {
	Send(*Connection) error
	grpc.ServerStream
}

type swarmConnectionsServer struct {
	grpc.ServerStream
}

func (x *swarmConnectionsServer) Send(m *Connection) error {
	return x.ServerStream.SendMsg(m)
}

var _Swarm_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.alice.grpc.swarm.Swarm",
	HandlerType: (*SwarmServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LocalPeer",
			Handler:    _Swarm_LocalPeer_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Peers",
			Handler:       _Swarm_Peers_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Connections",
			Handler:       _Swarm_Connections_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/alice/grpc/swarm/swarm.proto",
}

func init() { proto.RegisterFile("github.com/stratumn/alice/grpc/swarm/swarm.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 400 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0x41, 0x8b, 0xda, 0x40,
	0x14, 0xc7, 0x49, 0xaa, 0xb6, 0x7d, 0x55, 0x0f, 0x43, 0x29, 0x31, 0x28, 0x1d, 0x52, 0x4a, 0xab,
	0x95, 0x44, 0xda, 0x63, 0x4f, 0xd5, 0x5e, 0x0a, 0x1e, 0x8a, 0xd2, 0x4b, 0x2f, 0x12, 0x93, 0xc1,
	0x0c, 0x24, 0x99, 0x38, 0x33, 0xe2, 0x7a, 0x5b, 0xdc, 0x6f, 0xb2, 0xdf, 0xc1, 0x93, 0x5f, 0xca,
	0xe3, 0x1e, 0x97, 0x99, 0xb8, 0x9a, 0x15, 0x56, 0x3d, 0x64, 0x18, 0xf2, 0x7e, 0xf9, 0xff, 0x1f,
	0xff, 0xf7, 0x02, 0xbd, 0x19, 0x95, 0xd1, 0x62, 0xea, 0x06, 0x2c, 0xf1, 0x84, 0xe4, 0xbe, 0x5c,
	0x24, 0xa9, 0xe7, 0xc7, 0x34, 0x20, 0xde, 0x8c, 0x67, 0x81, 0x27, 0x96, 0x3e, 0x4f, 0xf2, 0xd3,
	0xcd, 0x38, 0x93, 0x0c, 0x35, 0x9e, 0x30, 0x57, 0x63, 0xae, 0xc2, 0x5c, 0x0d, 0xd8, 0xdd, 0x0b,
	0x62, 0xe4, 0x46, 0xaa, 0x27, 0x17, 0x72, 0xea, 0x50, 0x1d, 0xb2, 0xc0, 0x8f, 0xff, 0x12, 0xc2,
	0x47, 0x64, 0xee, 0x00, 0xbc, 0x51, 0x57, 0xa1, 0xee, 0xff, 0xa0, 0x3e, 0x60, 0x69, 0x4a, 0x02,
	0x49, 0x59, 0xaa, 0xde, 0xa0, 0x01, 0xbc, 0xce, 0x08, 0xe1, 0x13, 0x1a, 0x5a, 0x06, 0x36, 0xbe,
	0x56, 0xfb, 0x9d, 0xfb, 0xad, 0xe5, 0x8c, 0x23, 0xb6, 0xc4, 0x2c, 0x8d, 0x57, 0x38, 0x38, 0xe2,
	0x58, 0x32, 0x2c, 0x23, 0x2a, 0xb0, 0xfa, 0x60, 0xb7, 0xb5, 0x8c, 0x51, 0x45, 0xdd, 0xfe, 0x84,
	0x4e, 0x13, 0x4a, 0xca, 0x02, 0xbd, 0x07, 0xf3, 0xa0, 0x53, 0xd2, 0x84, 0x49, 0x43, 0xe7, 0xce,
	0x00, 0x38, 0xba, 0xa2, 0xd6, 0xa9, 0x63, 0xa9, 0xa8, 0x85, 0xda, 0x50, 0x8b, 0x55, 0xfb, 0x13,
	0x3f, 0x0c, 0x39, 0x11, 0xc2, 0x32, 0x73, 0xe8, 0x41, 0x41, 0x55, 0x5d, 0xfa, 0x95, 0x57, 0xd0,
	0x37, 0xa8, 0x73, 0x92, 0x30, 0x49, 0x0e, 0xec, 0xab, 0x02, 0x5b, 0xcb, 0x6b, 0x7b, 0xf8, 0xfb,
	0xce, 0x84, 0xf2, 0x58, 0xc5, 0x89, 0xe6, 0xf0, 0xf6, 0x10, 0x10, 0xfa, 0xe2, 0xbe, 0x98, 0xbb,
	0x5b, 0x8c, 0xd1, 0xfe, 0x78, 0x06, 0x54, 0x8c, 0x63, 0xaf, 0x37, 0xd6, 0x87, 0xdf, 0x54, 0x64,
	0xb1, 0xbf, 0xc2, 0x32, 0x22, 0x58, 0xf7, 0xaa, 0xc3, 0x42, 0x73, 0x28, 0xeb, 0x19, 0xa0, 0x4f,
	0x17, 0x54, 0xc4, 0x55, 0x56, 0x78, 0xbd, 0xb1, 0x9a, 0x43, 0x2a, 0x24, 0xf6, 0xe3, 0x58, 0x7b,
	0xed, 0xc7, 0x44, 0x42, 0xed, 0x27, 0x7a, 0x06, 0xba, 0x35, 0xe0, 0x5d, 0x61, 0xd6, 0xa8, 0x7d,
	0x46, 0xf4, 0xf9, 0x4e, 0xd8, 0x9f, 0xaf, 0x42, 0x9d, 0xd6, 0x7a, 0x63, 0x35, 0x74, 0x17, 0x27,
	0x4b, 0xb2, 0x6f, 0xa1, 0xdf, 0xfd, 0xdf, 0xb9, 0xe6, 0x37, 0xf8, 0xa9, 0xcf, 0x69, 0x45, 0xaf,
	0xef, 0x8f, 0xc7, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3e, 0x04, 0xe0, 0x28, 0x3b, 0x03, 0x00, 0x00,
}