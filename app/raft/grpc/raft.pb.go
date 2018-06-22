// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/alice/app/raft/grpc/raft.proto

/*
	Package grpc is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/alice/app/raft/grpc/raft.proto

	It has these top-level messages:
		Empty
		Peer
		PeerID
		StatusInfo
		Proposal
		Entry
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/cli/grpc/ext"

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

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptorRaft, []int{0} }

type Peer struct {
	Id      uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Address []byte `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *Peer) Reset()                    { *m = Peer{} }
func (m *Peer) String() string            { return proto.CompactTextString(m) }
func (*Peer) ProtoMessage()               {}
func (*Peer) Descriptor() ([]byte, []int) { return fileDescriptorRaft, []int{1} }

func (m *Peer) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Peer) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

type PeerID struct {
	Address []byte `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *PeerID) Reset()                    { *m = PeerID{} }
func (m *PeerID) String() string            { return proto.CompactTextString(m) }
func (*PeerID) ProtoMessage()               {}
func (*PeerID) Descriptor() ([]byte, []int) { return fileDescriptorRaft, []int{2} }

func (m *PeerID) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

type StatusInfo struct {
	Running bool   `protobuf:"varint,1,opt,name=running,proto3" json:"running,omitempty"`
	Id      uint64 `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *StatusInfo) Reset()                    { *m = StatusInfo{} }
func (m *StatusInfo) String() string            { return proto.CompactTextString(m) }
func (*StatusInfo) ProtoMessage()               {}
func (*StatusInfo) Descriptor() ([]byte, []int) { return fileDescriptorRaft, []int{3} }

func (m *StatusInfo) GetRunning() bool {
	if m != nil {
		return m.Running
	}
	return false
}

func (m *StatusInfo) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type Proposal struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Proposal) Reset()                    { *m = Proposal{} }
func (m *Proposal) String() string            { return proto.CompactTextString(m) }
func (*Proposal) ProtoMessage()               {}
func (*Proposal) Descriptor() ([]byte, []int) { return fileDescriptorRaft, []int{4} }

func (m *Proposal) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Entry struct {
	Index uint64 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Data  []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Entry) Reset()                    { *m = Entry{} }
func (m *Entry) String() string            { return proto.CompactTextString(m) }
func (*Entry) ProtoMessage()               {}
func (*Entry) Descriptor() ([]byte, []int) { return fileDescriptorRaft, []int{5} }

func (m *Entry) GetIndex() uint64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *Entry) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Empty)(nil), "stratumn.alice.app.raft.grpc.Empty")
	proto.RegisterType((*Peer)(nil), "stratumn.alice.app.raft.grpc.Peer")
	proto.RegisterType((*PeerID)(nil), "stratumn.alice.app.raft.grpc.PeerID")
	proto.RegisterType((*StatusInfo)(nil), "stratumn.alice.app.raft.grpc.StatusInfo")
	proto.RegisterType((*Proposal)(nil), "stratumn.alice.app.raft.grpc.Proposal")
	proto.RegisterType((*Entry)(nil), "stratumn.alice.app.raft.grpc.Entry")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for Raft service

type RaftClient interface {
	Start(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (*Empty, error)
	Stop(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (*Empty, error)
	Status(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (*StatusInfo, error)
	Peers(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (Raft_PeersClient, error)
	Discover(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (Raft_DiscoverClient, error)
	Invite(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (*Empty, error)
	Join(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (*Empty, error)
	Expel(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (*Empty, error)
	Propose(ctx context.Context, in *Proposal, opts ...grpc1.CallOption) (*Empty, error)
	Log(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (Raft_LogClient, error)
}

type raftClient struct {
	cc *grpc1.ClientConn
}

func NewRaftClient(cc *grpc1.ClientConn) RaftClient {
	return &raftClient{cc}
}

func (c *raftClient) Start(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Start", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Stop(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Stop", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Status(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (*StatusInfo, error) {
	out := new(StatusInfo)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Status", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Peers(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (Raft_PeersClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Raft_serviceDesc.Streams[0], c.cc, "/stratumn.alice.app.raft.grpc.Raft/Peers", opts...)
	if err != nil {
		return nil, err
	}
	x := &raftPeersClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Raft_PeersClient interface {
	Recv() (*Peer, error)
	grpc1.ClientStream
}

type raftPeersClient struct {
	grpc1.ClientStream
}

func (x *raftPeersClient) Recv() (*Peer, error) {
	m := new(Peer)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *raftClient) Discover(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (Raft_DiscoverClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Raft_serviceDesc.Streams[1], c.cc, "/stratumn.alice.app.raft.grpc.Raft/Discover", opts...)
	if err != nil {
		return nil, err
	}
	x := &raftDiscoverClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Raft_DiscoverClient interface {
	Recv() (*Peer, error)
	grpc1.ClientStream
}

type raftDiscoverClient struct {
	grpc1.ClientStream
}

func (x *raftDiscoverClient) Recv() (*Peer, error) {
	m := new(Peer)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *raftClient) Invite(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Invite", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Join(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Join", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Expel(ctx context.Context, in *PeerID, opts ...grpc1.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Expel", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Propose(ctx context.Context, in *Proposal, opts ...grpc1.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc1.Invoke(ctx, "/stratumn.alice.app.raft.grpc.Raft/Propose", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *raftClient) Log(ctx context.Context, in *Empty, opts ...grpc1.CallOption) (Raft_LogClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Raft_serviceDesc.Streams[2], c.cc, "/stratumn.alice.app.raft.grpc.Raft/Log", opts...)
	if err != nil {
		return nil, err
	}
	x := &raftLogClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Raft_LogClient interface {
	Recv() (*Entry, error)
	grpc1.ClientStream
}

type raftLogClient struct {
	grpc1.ClientStream
}

func (x *raftLogClient) Recv() (*Entry, error) {
	m := new(Entry)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Raft service

type RaftServer interface {
	Start(context.Context, *Empty) (*Empty, error)
	Stop(context.Context, *Empty) (*Empty, error)
	Status(context.Context, *Empty) (*StatusInfo, error)
	Peers(*Empty, Raft_PeersServer) error
	Discover(*PeerID, Raft_DiscoverServer) error
	Invite(context.Context, *PeerID) (*Empty, error)
	Join(context.Context, *PeerID) (*Empty, error)
	Expel(context.Context, *PeerID) (*Empty, error)
	Propose(context.Context, *Proposal) (*Empty, error)
	Log(*Empty, Raft_LogServer) error
}

func RegisterRaftServer(s *grpc1.Server, srv RaftServer) {
	s.RegisterService(&_Raft_serviceDesc, srv)
}

func _Raft_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Start(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Start(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Stop(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Stop(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Status(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Status(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Peers_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RaftServer).Peers(m, &raftPeersServer{stream})
}

type Raft_PeersServer interface {
	Send(*Peer) error
	grpc1.ServerStream
}

type raftPeersServer struct {
	grpc1.ServerStream
}

func (x *raftPeersServer) Send(m *Peer) error {
	return x.ServerStream.SendMsg(m)
}

func _Raft_Discover_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(PeerID)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RaftServer).Discover(m, &raftDiscoverServer{stream})
}

type Raft_DiscoverServer interface {
	Send(*Peer) error
	grpc1.ServerStream
}

type raftDiscoverServer struct {
	grpc1.ServerStream
}

func (x *raftDiscoverServer) Send(m *Peer) error {
	return x.ServerStream.SendMsg(m)
}

func _Raft_Invite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Invite(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Invite",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Invite(ctx, req.(*PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Join(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Join",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Join(ctx, req.(*PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Expel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Expel(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Expel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Expel(ctx, req.(*PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Propose_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(Proposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftServer).Propose(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.app.raft.grpc.Raft/Propose",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftServer).Propose(ctx, req.(*Proposal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Raft_Log_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RaftServer).Log(m, &raftLogServer{stream})
}

type Raft_LogServer interface {
	Send(*Entry) error
	grpc1.ServerStream
}

type raftLogServer struct {
	grpc1.ServerStream
}

func (x *raftLogServer) Send(m *Entry) error {
	return x.ServerStream.SendMsg(m)
}

var _Raft_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "stratumn.alice.app.raft.grpc.Raft",
	HandlerType: (*RaftServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _Raft_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Raft_Stop_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _Raft_Status_Handler,
		},
		{
			MethodName: "Invite",
			Handler:    _Raft_Invite_Handler,
		},
		{
			MethodName: "Join",
			Handler:    _Raft_Join_Handler,
		},
		{
			MethodName: "Expel",
			Handler:    _Raft_Expel_Handler,
		},
		{
			MethodName: "Propose",
			Handler:    _Raft_Propose_Handler,
		},
	},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "Peers",
			Handler:       _Raft_Peers_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Discover",
			Handler:       _Raft_Discover_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Log",
			Handler:       _Raft_Log_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/alice/app/raft/grpc/raft.proto",
}

func (m *Empty) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Empty) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
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
	if m.Id != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintRaft(dAtA, i, uint64(m.Id))
	}
	if len(m.Address) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintRaft(dAtA, i, uint64(len(m.Address)))
		i += copy(dAtA[i:], m.Address)
	}
	return i, nil
}

func (m *PeerID) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeerID) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintRaft(dAtA, i, uint64(len(m.Address)))
		i += copy(dAtA[i:], m.Address)
	}
	return i, nil
}

func (m *StatusInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StatusInfo) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Running {
		dAtA[i] = 0x8
		i++
		if m.Running {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.Id != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintRaft(dAtA, i, uint64(m.Id))
	}
	return i, nil
}

func (m *Proposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Proposal) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintRaft(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *Entry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Entry) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Index != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintRaft(dAtA, i, uint64(m.Index))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintRaft(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func encodeVarintRaft(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Empty) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *Peer) Size() (n int) {
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovRaft(uint64(m.Id))
	}
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovRaft(uint64(l))
	}
	return n
}

func (m *PeerID) Size() (n int) {
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovRaft(uint64(l))
	}
	return n
}

func (m *StatusInfo) Size() (n int) {
	var l int
	_ = l
	if m.Running {
		n += 2
	}
	if m.Id != 0 {
		n += 1 + sovRaft(uint64(m.Id))
	}
	return n
}

func (m *Proposal) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovRaft(uint64(l))
	}
	return n
}

func (m *Entry) Size() (n int) {
	var l int
	_ = l
	if m.Index != 0 {
		n += 1 + sovRaft(uint64(m.Index))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovRaft(uint64(l))
	}
	return n
}

func sovRaft(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozRaft(x uint64) (n int) {
	return sovRaft(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Empty) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRaft
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
			return fmt.Errorf("proto: Empty: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Empty: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipRaft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRaft
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
				return ErrIntOverflowRaft
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
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
				return ErrInvalidLengthRaft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = append(m.Address[:0], dAtA[iNdEx:postIndex]...)
			if m.Address == nil {
				m.Address = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRaft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRaft
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
func (m *PeerID) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRaft
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
			return fmt.Errorf("proto: PeerID: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeerID: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
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
				return ErrInvalidLengthRaft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = append(m.Address[:0], dAtA[iNdEx:postIndex]...)
			if m.Address == nil {
				m.Address = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRaft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRaft
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
func (m *StatusInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRaft
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
			return fmt.Errorf("proto: StatusInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StatusInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Running", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Running = bool(v != 0)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipRaft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRaft
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
func (m *Proposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRaft
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
			return fmt.Errorf("proto: Proposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Proposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
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
				return ErrInvalidLengthRaft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRaft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRaft
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
func (m *Entry) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRaft
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
			return fmt.Errorf("proto: Entry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Entry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			m.Index = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Index |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRaft
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
				return ErrInvalidLengthRaft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRaft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthRaft
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
func skipRaft(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRaft
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
					return 0, ErrIntOverflowRaft
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
					return 0, ErrIntOverflowRaft
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
				return 0, ErrInvalidLengthRaft
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowRaft
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
				next, err := skipRaft(dAtA[start:])
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
	ErrInvalidLengthRaft = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRaft   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/alice/app/raft/grpc/raft.proto", fileDescriptorRaft)
}

var fileDescriptorRaft = []byte{
	// 652 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x95, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0x86, 0x3f, 0xe7, 0xb3, 0x93, 0x30, 0x54, 0x42, 0x0c, 0x6a, 0x35, 0x2d, 0x6d, 0x18, 0x5c,
	0x7e, 0x8a, 0x54, 0x9c, 0xa8, 0x2c, 0x58, 0x20, 0x21, 0x51, 0xa5, 0x20, 0x23, 0x84, 0x22, 0x77,
	0xc7, 0x6e, 0x1a, 0x4f, 0xdc, 0x91, 0x9c, 0x19, 0x33, 0x3e, 0xa9, 0xd2, 0xad, 0xaf, 0x82, 0x75,
	0x57, 0xdc, 0x80, 0x57, 0xb9, 0x01, 0x96, 0x5c, 0x02, 0x2a, 0x17, 0x00, 0x97, 0x80, 0xc6, 0x3f,
	0x94, 0x4a, 0xd0, 0x7a, 0x91, 0x45, 0x22, 0xeb, 0xcc, 0x7b, 0xde, 0xe7, 0xcc, 0x9c, 0xe3, 0x31,
	0xda, 0x8b, 0x04, 0x1c, 0xcf, 0x8e, 0xbc, 0xb1, 0x9a, 0xf6, 0x53, 0xd0, 0x0c, 0x66, 0x53, 0xd9,
	0x67, 0xb1, 0x18, 0xf3, 0x3e, 0x4b, 0x92, 0xbe, 0x66, 0x13, 0xe8, 0x47, 0x3a, 0x19, 0x17, 0x4f,
	0x5e, 0xa2, 0x15, 0x28, 0xbc, 0x59, 0x0b, 0xbd, 0x42, 0xe8, 0xb1, 0x24, 0xf1, 0x8a, 0x65, 0x23,
	0xdc, 0x18, 0xfc, 0xdb, 0x71, 0x1c, 0x8b, 0xd2, 0x8c, 0xcf, 0xc1, 0xfc, 0x4a, 0x3f, 0xb7, 0x83,
	0x9c, 0x83, 0x69, 0x02, 0xa7, 0xee, 0x7b, 0x64, 0x8f, 0x38, 0xd7, 0xf8, 0x2e, 0x6a, 0x89, 0x90,
	0x58, 0xd4, 0xda, 0xb1, 0xf7, 0x6f, 0x9e, 0x2d, 0x48, 0x27, 0x60, 0x13, 0xa0, 0xfe, 0x30, 0x68,
	0x89, 0x10, 0x3f, 0x41, 0x1d, 0x16, 0x86, 0x9a, 0xa7, 0x29, 0x69, 0x51, 0x6b, 0x67, 0x65, 0xff,
	0x96, 0x51, 0x98, 0x3c, 0xea, 0x0f, 0x7f, 0x2e, 0x88, 0x15, 0xd4, 0xeb, 0xee, 0x73, 0xd4, 0x36,
	0x71, 0x7f, 0x88, 0x9f, 0x5e, 0x24, 0x59, 0x45, 0xd2, 0x9d, 0x3f, 0x92, 0x3e, 0x2f, 0x88, 0x75,
	0x39, 0x71, 0x84, 0xd0, 0x21, 0x30, 0x98, 0xa5, 0xbe, 0x9c, 0x28, 0xfc, 0x10, 0x75, 0xf4, 0x4c,
	0x4a, 0x21, 0xa3, 0x22, 0xb9, 0x5b, 0xd5, 0x54, 0x86, 0x82, 0x7a, 0xad, 0xaa, 0xba, 0xf5, 0xd7,
	0xaa, 0xdd, 0x5d, 0xd4, 0x1d, 0x69, 0x95, 0xa8, 0x94, 0xc5, 0x98, 0x22, 0x3b, 0x64, 0xc0, 0xaa,
	0x4a, 0x56, 0xce, 0x16, 0xc4, 0x1e, 0x32, 0x60, 0xa6, 0x8c, 0xa0, 0x58, 0x71, 0x5f, 0x23, 0xe7,
	0x40, 0x82, 0x3e, 0xc5, 0xf7, 0x90, 0x23, 0x64, 0xc8, 0xe7, 0xd5, 0x61, 0xdc, 0x38, 0x5b, 0x10,
	0xc7, 0x37, 0x81, 0xa0, 0x8c, 0xe3, 0xcd, 0xca, 0xab, 0x3c, 0x8a, 0x6e, 0xed, 0x55, 0xfa, 0xec,
	0xfd, 0xe8, 0x22, 0xdb, 0x54, 0x81, 0x01, 0x39, 0x87, 0xc0, 0x34, 0xe0, 0x6d, 0xef, 0xaa, 0xe6,
	0x79, 0x45, 0x1f, 0x36, 0x9a, 0x88, 0xdc, 0xfb, 0x59, 0x4e, 0xb6, 0x0a, 0x53, 0x9a, 0x0a, 0x19,
	0xc5, 0x9c, 0x4a, 0x15, 0x72, 0x6a, 0x64, 0x34, 0xe5, 0xfa, 0x84, 0x6b, 0x7c, 0x8a, 0xec, 0x43,
	0x50, 0xc9, 0x12, 0xa1, 0x8f, 0xb3, 0x9c, 0x6c, 0x1b, 0x4f, 0x0a, 0xc7, 0xbc, 0xc2, 0x50, 0x26,
	0x43, 0xaa, 0xf9, 0x54, 0x9d, 0xf0, 0x22, 0x6a, 0x76, 0x8e, 0x25, 0x6a, 0x97, 0x1d, 0x6c, 0x06,
	0xdf, 0xb9, 0x5a, 0x74, 0x31, 0x0c, 0xee, 0x6a, 0x96, 0x93, 0xdb, 0x6f, 0x78, 0xbd, 0x47, 0x9a,
	0x96, 0x14, 0x81, 0x1c, 0x33, 0x4d, 0x0d, 0x71, 0xee, 0xd5, 0x22, 0xe3, 0x74, 0x01, 0x1a, 0xc7,
	0xb3, 0x14, 0xb8, 0xa6, 0x89, 0xf1, 0x1f, 0x58, 0xf8, 0x23, 0xea, 0x0e, 0x45, 0x3a, 0x56, 0xe6,
	0x84, 0x1f, 0x5c, 0x6f, 0xe4, 0x0f, 0x1b, 0xe1, 0xd6, 0xb3, 0x9c, 0xac, 0xd6, 0xbe, 0xb4, 0x18,
	0xdd, 0x0a, 0x3c, 0xb0, 0xb0, 0x46, 0x6d, 0x5f, 0x9e, 0x08, 0xe0, 0x0d, 0x81, 0x8d, 0x7a, 0xb9,
	0x95, 0xe5, 0x64, 0xfd, 0x55, 0x18, 0x52, 0x56, 0x4e, 0x0e, 0xa8, 0xa2, 0x7d, 0x15, 0x15, 0x4f,
	0x91, 0xfd, 0x56, 0x09, 0xb9, 0x4c, 0xe2, 0x46, 0x96, 0x93, 0x35, 0x63, 0x4a, 0xf9, 0x5c, 0xa4,
	0x02, 0x84, 0x8c, 0x7e, 0xe3, 0x00, 0x39, 0x07, 0xf3, 0x84, 0xc7, 0xcb, 0xe4, 0xd1, 0x2c, 0x27,
	0x9b, 0x41, 0x39, 0x99, 0xa6, 0x7b, 0x74, 0xa2, 0xd5, 0xf4, 0xd2, 0x26, 0x43, 0xd4, 0x29, 0xaf,
	0x05, 0x8e, 0x1f, 0x5d, 0xc3, 0xad, 0x6e, 0x8f, 0x66, 0xe4, 0x95, 0x2c, 0x27, 0xdd, 0xd1, 0x0c,
	0xca, 0x97, 0x41, 0xa0, 0xff, 0xdf, 0xa9, 0x68, 0x39, 0xaf, 0xa1, 0xb9, 0x96, 0xdc, 0xb5, 0x2c,
	0x27, 0xd8, 0xcc, 0x66, 0xac, 0x22, 0xaa, 0x26, 0x94, 0x4b, 0xd0, 0x82, 0xa7, 0x03, 0x6b, 0xff,
	0xe5, 0x97, 0xf3, 0x9e, 0xf5, 0xf5, 0xbc, 0x67, 0x7d, 0x3b, 0xef, 0x59, 0x9f, 0xbe, 0xf7, 0xfe,
	0xfb, 0xb0, 0xdb, 0xf0, 0x0b, 0xf3, 0xc2, 0xfc, 0x1d, 0xb5, 0x8b, 0x4f, 0xc2, 0xb3, 0x5f, 0x01,
	0x00, 0x00, 0xff, 0xff, 0xc2, 0xc1, 0x44, 0x83, 0x98, 0x06, 0x00, 0x00,
}
