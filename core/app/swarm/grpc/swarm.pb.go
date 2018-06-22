// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/alice/core/app/swarm/grpc/swarm.proto

/*
	Package grpc is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/alice/core/app/swarm/grpc/swarm.proto

	It has these top-level messages:
		LocalPeerReq
		PeersReq
		ConnectionsReq
		Peer
		Connection
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/grpc/ext"

import context "context"
import grpc1 "google.golang.org/grpc"

import io "io"

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
func (*LocalPeerReq) Descriptor() ([]byte, []int) { return fileDescriptorSwarm, []int{0} }

// The peers request message.
type PeersReq struct {
}

func (m *PeersReq) Reset()                    { *m = PeersReq{} }
func (m *PeersReq) String() string            { return proto.CompactTextString(m) }
func (*PeersReq) ProtoMessage()               {}
func (*PeersReq) Descriptor() ([]byte, []int) { return fileDescriptorSwarm, []int{1} }

// The connections request message.
type ConnectionsReq struct {
	PeerId []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
}

func (m *ConnectionsReq) Reset()                    { *m = ConnectionsReq{} }
func (m *ConnectionsReq) String() string            { return proto.CompactTextString(m) }
func (*ConnectionsReq) ProtoMessage()               {}
func (*ConnectionsReq) Descriptor() ([]byte, []int) { return fileDescriptorSwarm, []int{2} }

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
func (*Peer) Descriptor() ([]byte, []int) { return fileDescriptorSwarm, []int{3} }

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
func (*Connection) Descriptor() ([]byte, []int) { return fileDescriptorSwarm, []int{4} }

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
	proto.RegisterType((*LocalPeerReq)(nil), "stratumn.alice.core.app.swarm.grpc.LocalPeerReq")
	proto.RegisterType((*PeersReq)(nil), "stratumn.alice.core.app.swarm.grpc.PeersReq")
	proto.RegisterType((*ConnectionsReq)(nil), "stratumn.alice.core.app.swarm.grpc.ConnectionsReq")
	proto.RegisterType((*Peer)(nil), "stratumn.alice.core.app.swarm.grpc.Peer")
	proto.RegisterType((*Connection)(nil), "stratumn.alice.core.app.swarm.grpc.Connection")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for Swarm service

type SwarmClient interface {
	// Returns the local peer.
	LocalPeer(ctx context.Context, in *LocalPeerReq, opts ...grpc1.CallOption) (*Peer, error)
	// Streams the connected peers.
	Peers(ctx context.Context, in *PeersReq, opts ...grpc1.CallOption) (Swarm_PeersClient, error)
	// Streams connections.
	Connections(ctx context.Context, in *ConnectionsReq, opts ...grpc1.CallOption) (Swarm_ConnectionsClient, error)
}

type swarmClient struct {
	cc *grpc1.ClientConn
}

func NewSwarmClient(cc *grpc1.ClientConn) SwarmClient {
	return &swarmClient{cc}
}

func (c *swarmClient) LocalPeer(ctx context.Context, in *LocalPeerReq, opts ...grpc1.CallOption) (*Peer, error) {
	out := new(Peer)
	err := grpc1.Invoke(ctx, "/stratumn.alice.core.app.swarm.grpc.Swarm/LocalPeer", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *swarmClient) Peers(ctx context.Context, in *PeersReq, opts ...grpc1.CallOption) (Swarm_PeersClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Swarm_serviceDesc.Streams[0], c.cc, "/stratumn.alice.core.app.swarm.grpc.Swarm/Peers", opts...)
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
	grpc1.ClientStream
}

type swarmPeersClient struct {
	grpc1.ClientStream
}

func (x *swarmPeersClient) Recv() (*Peer, error) {
	m := new(Peer)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *swarmClient) Connections(ctx context.Context, in *ConnectionsReq, opts ...grpc1.CallOption) (Swarm_ConnectionsClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Swarm_serviceDesc.Streams[1], c.cc, "/stratumn.alice.core.app.swarm.grpc.Swarm/Connections", opts...)
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
	grpc1.ClientStream
}

type swarmConnectionsClient struct {
	grpc1.ClientStream
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

func RegisterSwarmServer(s *grpc1.Server, srv SwarmServer) {
	s.RegisterService(&_Swarm_serviceDesc, srv)
}

func _Swarm_LocalPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocalPeerReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SwarmServer).LocalPeer(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.core.app.swarm.grpc.Swarm/LocalPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SwarmServer).LocalPeer(ctx, req.(*LocalPeerReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Swarm_Peers_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(PeersReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SwarmServer).Peers(m, &swarmPeersServer{stream})
}

type Swarm_PeersServer interface {
	Send(*Peer) error
	grpc1.ServerStream
}

type swarmPeersServer struct {
	grpc1.ServerStream
}

func (x *swarmPeersServer) Send(m *Peer) error {
	return x.ServerStream.SendMsg(m)
}

func _Swarm_Connections_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(ConnectionsReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SwarmServer).Connections(m, &swarmConnectionsServer{stream})
}

type Swarm_ConnectionsServer interface {
	Send(*Connection) error
	grpc1.ServerStream
}

type swarmConnectionsServer struct {
	grpc1.ServerStream
}

func (x *swarmConnectionsServer) Send(m *Connection) error {
	return x.ServerStream.SendMsg(m)
}

var _Swarm_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "stratumn.alice.core.app.swarm.grpc.Swarm",
	HandlerType: (*SwarmServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "LocalPeer",
			Handler:    _Swarm_LocalPeer_Handler,
		},
	},
	Streams: []grpc1.StreamDesc{
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
	Metadata: "github.com/stratumn/alice/core/app/swarm/grpc/swarm.proto",
}

func (m *LocalPeerReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LocalPeerReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *PeersReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeersReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *ConnectionsReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConnectionsReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSwarm(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	return i, nil
}

func (m *Peer) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Peer) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Id) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSwarm(dAtA, i, uint64(len(m.Id)))
		i += copy(dAtA[i:], m.Id)
	}
	return i, nil
}

func (m *Connection) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Connection) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSwarm(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if len(m.LocalAddress) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintSwarm(dAtA, i, uint64(len(m.LocalAddress)))
		i += copy(dAtA[i:], m.LocalAddress)
	}
	if len(m.RemoteAddress) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintSwarm(dAtA, i, uint64(len(m.RemoteAddress)))
		i += copy(dAtA[i:], m.RemoteAddress)
	}
	return i, nil
}

func encodeVarintSwarm(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *LocalPeerReq) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *PeersReq) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *ConnectionsReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovSwarm(uint64(l))
	}
	return n
}

func (m *Peer) Size() (n int) {
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovSwarm(uint64(l))
	}
	return n
}

func (m *Connection) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovSwarm(uint64(l))
	}
	l = len(m.LocalAddress)
	if l > 0 {
		n += 1 + l + sovSwarm(uint64(l))
	}
	l = len(m.RemoteAddress)
	if l > 0 {
		n += 1 + l + sovSwarm(uint64(l))
	}
	return n
}

func sovSwarm(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozSwarm(x uint64) (n int) {
	return sovSwarm(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *LocalPeerReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSwarm
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LocalPeerReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LocalPeerReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSwarm(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSwarm
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PeersReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSwarm
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PeersReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeersReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSwarm(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSwarm
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ConnectionsReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSwarm
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ConnectionsReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ConnectionsReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSwarm
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PeerId = append(m.PeerId[:0], dAtA[iNdEx:postIndex]...)
			if m.PeerId == nil {
				m.PeerId = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSwarm(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSwarm
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Peer) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSwarm
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Peer: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Peer: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSwarm
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = append(m.Id[:0], dAtA[iNdEx:postIndex]...)
			if m.Id == nil {
				m.Id = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSwarm(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSwarm
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Connection) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSwarm
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Connection: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Connection: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSwarm
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PeerId = append(m.PeerId[:0], dAtA[iNdEx:postIndex]...)
			if m.PeerId == nil {
				m.PeerId = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LocalAddress", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSwarm
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LocalAddress = append(m.LocalAddress[:0], dAtA[iNdEx:postIndex]...)
			if m.LocalAddress == nil {
				m.LocalAddress = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemoteAddress", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSwarm
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RemoteAddress = append(m.RemoteAddress[:0], dAtA[iNdEx:postIndex]...)
			if m.RemoteAddress == nil {
				m.RemoteAddress = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSwarm(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSwarm
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipSwarm(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSwarm
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSwarm
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthSwarm
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowSwarm
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipSwarm(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthSwarm = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSwarm   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/alice/core/app/swarm/grpc/swarm.proto", fileDescriptorSwarm)
}

var fileDescriptorSwarm = []byte{
	// 439 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xb1, 0x8e, 0x13, 0x31,
	0x10, 0x86, 0x71, 0x2e, 0x77, 0x80, 0xc9, 0xa5, 0xb0, 0x10, 0x5a, 0x56, 0x77, 0x2b, 0xcb, 0xd5,
	0x01, 0x27, 0xef, 0x29, 0x54, 0x88, 0x8a, 0xbb, 0x6b, 0x90, 0x52, 0xa0, 0x44, 0x34, 0x34, 0x91,
	0xb3, 0x6b, 0x65, 0x2d, 0xed, 0xae, 0x8d, 0xed, 0x28, 0xa4, 0x4d, 0xde, 0x80, 0x8a, 0x9a, 0x8a,
	0x17, 0x48, 0x95, 0x17, 0xa0, 0xe4, 0x11, 0x50, 0x78, 0x01, 0x4a, 0x4a, 0x64, 0x6f, 0x48, 0x96,
	0x48, 0x88, 0xa4, 0x48, 0x34, 0xda, 0xf9, 0xe6, 0x9f, 0x5f, 0x33, 0x63, 0xf8, 0x62, 0x24, 0x6c,
	0x36, 0x1e, 0xd2, 0x44, 0x16, 0xb1, 0xb1, 0x9a, 0xd9, 0x71, 0x51, 0xc6, 0x2c, 0x17, 0x09, 0x8f,
	0x13, 0xa9, 0x79, 0xcc, 0x94, 0x8a, 0xcd, 0x84, 0xe9, 0x22, 0x1e, 0x69, 0x95, 0x54, 0x21, 0x55,
	0x5a, 0x5a, 0x89, 0xc8, 0x1f, 0x9e, 0x7a, 0x9e, 0x3a, 0x9e, 0x32, 0xa5, 0x68, 0x05, 0x39, 0x3e,
	0xbc, 0xfc, 0xb7, 0xbc, 0xd7, 0xe3, 0x1f, 0xac, 0xfb, 0x55, 0x8a, 0xa4, 0x0d, 0x5b, 0x5d, 0x99,
	0xb0, 0xfc, 0x0d, 0xe7, 0xba, 0xc7, 0xdf, 0x13, 0x08, 0xef, 0xb9, 0xd0, 0xb8, 0xf8, 0x2d, 0x6c,
	0xdf, 0xc8, 0xb2, 0xe4, 0x89, 0x15, 0xb2, 0x74, 0x5f, 0xd0, 0x0d, 0xbc, 0xab, 0x38, 0xd7, 0x03,
	0x91, 0x06, 0x00, 0x83, 0x8b, 0xd6, 0xf5, 0xd3, 0xcf, 0xcb, 0x80, 0xf4, 0x33, 0x39, 0xc1, 0xb2,
	0xcc, 0xa7, 0x38, 0xd9, 0xe2, 0xd8, 0x4a, 0x6c, 0x33, 0x61, 0xb0, 0x2b, 0xf8, 0xb9, 0x0c, 0x40,
	0xef, 0xc4, 0x45, 0xaf, 0x53, 0x72, 0x06, 0x9b, 0xae, 0x05, 0x7a, 0x08, 0x1b, 0x1b, 0x9d, 0xa6,
	0x27, 0x1a, 0x22, 0x25, 0x73, 0x00, 0xe1, 0xb6, 0x2b, 0x3a, 0xdf, 0xed, 0xd8, 0xac, 0x6b, 0xa1,
	0x27, 0xf0, 0x34, 0x77, 0xf6, 0x07, 0x2c, 0x4d, 0x35, 0x37, 0x26, 0x68, 0x54, 0xd0, 0x2f, 0x07,
	0xb5, 0x7c, 0xea, 0x55, 0x95, 0x41, 0xcf, 0x60, 0x5b, 0xf3, 0x42, 0x5a, 0xbe, 0x61, 0x8f, 0x6a,
	0xec, 0x69, 0x95, 0x5b, 0xc3, 0x9d, 0x2f, 0x47, 0xf0, 0xb8, 0xef, 0x66, 0x8a, 0xe6, 0x00, 0xde,
	0xdf, 0x4c, 0x08, 0x5d, 0xd1, 0xff, 0x6f, 0x80, 0xd6, 0x07, 0x1a, 0x5e, 0xec, 0x53, 0xe1, 0x60,
	0x12, 0xce, 0x16, 0xc1, 0xa3, 0x5b, 0x61, 0x54, 0xce, 0xa6, 0xd8, 0x66, 0x1c, 0x7b, 0xfb, 0x7e,
	0x7e, 0xce, 0xc5, 0xb1, 0xdf, 0x0b, 0xba, 0xdc, 0x57, 0xcf, 0x1c, 0xd6, 0x1d, 0xcf, 0x16, 0xc1,
	0x59, 0x57, 0x18, 0x8b, 0x59, 0x9e, 0xfb, 0xf6, 0xeb, 0x65, 0xf2, 0xd4, 0x5b, 0x30, 0x57, 0x00,
	0x7d, 0x04, 0xf0, 0x41, 0xed, 0x22, 0x50, 0x67, 0x1f, 0xf5, 0xbf, 0x4f, 0x28, 0xa4, 0x87, 0xd5,
	0x90, 0xf3, 0xd9, 0x22, 0x78, 0xec, 0x7d, 0xed, 0x1c, 0xd7, 0xda, 0xd4, 0xf5, 0xed, 0xd7, 0x55,
	0x04, 0xbe, 0xad, 0x22, 0xf0, 0x7d, 0x15, 0x81, 0x4f, 0x3f, 0xa2, 0x3b, 0xef, 0x3a, 0x07, 0x3d,
	0xb0, 0x97, 0xee, 0x6f, 0x78, 0xe2, 0x9f, 0xc3, 0xf3, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x43,
	0x92, 0x3f, 0x66, 0x9d, 0x03, 0x00, 0x00,
}