// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/alice/core/app/bootstrap/grpc/bootstrap.proto

/*
	Package grpc is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/alice/core/app/bootstrap/grpc/bootstrap.proto

	It has these top-level messages:
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/cli/grpc/ext"
import stratumn_alice_core_app_bootstrap "github.com/stratumn/alice/core/app/bootstrap/pb"

import context "context"
import grpc1 "google.golang.org/grpc"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for Bootstrap service

type BootstrapClient interface {
	// Propose adding a node to the network.
	AddNode(ctx context.Context, in *stratumn_alice_core_app_bootstrap.NodeIdentity, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// Propose removing a node from the network.
	RemoveNode(ctx context.Context, in *stratumn_alice_core_app_bootstrap.NodeIdentity, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// Accept a proposal to add or remove a network node.
	Accept(ctx context.Context, in *stratumn_alice_core_app_bootstrap.PeerID, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// Reject a proposal to add or remove a network node.
	Reject(ctx context.Context, in *stratumn_alice_core_app_bootstrap.PeerID, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// List pending proposals to add or remove a network node.
	List(ctx context.Context, in *stratumn_alice_core_app_bootstrap.Filter, opts ...grpc1.CallOption) (Bootstrap_ListClient, error)
	// Complete the network bootstrap phase.
	Complete(ctx context.Context, in *stratumn_alice_core_app_bootstrap.CompleteReq, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error)
}

type bootstrapClient struct {
	cc *grpc1.ClientConn
}

func NewBootstrapClient(cc *grpc1.ClientConn) BootstrapClient {
	return &bootstrapClient{cc}
}

func (c *bootstrapClient) AddNode(ctx context.Context, in *stratumn_alice_core_app_bootstrap.NodeIdentity, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error) {
	out := new(stratumn_alice_core_app_bootstrap.Ack)
	err := grpc1.Invoke(ctx, "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/AddNode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) RemoveNode(ctx context.Context, in *stratumn_alice_core_app_bootstrap.NodeIdentity, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error) {
	out := new(stratumn_alice_core_app_bootstrap.Ack)
	err := grpc1.Invoke(ctx, "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/RemoveNode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) Accept(ctx context.Context, in *stratumn_alice_core_app_bootstrap.PeerID, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error) {
	out := new(stratumn_alice_core_app_bootstrap.Ack)
	err := grpc1.Invoke(ctx, "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/Accept", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) Reject(ctx context.Context, in *stratumn_alice_core_app_bootstrap.PeerID, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error) {
	out := new(stratumn_alice_core_app_bootstrap.Ack)
	err := grpc1.Invoke(ctx, "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/Reject", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) List(ctx context.Context, in *stratumn_alice_core_app_bootstrap.Filter, opts ...grpc1.CallOption) (Bootstrap_ListClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Bootstrap_serviceDesc.Streams[0], c.cc, "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/List", opts...)
	if err != nil {
		return nil, err
	}
	x := &bootstrapListClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Bootstrap_ListClient interface {
	Recv() (*stratumn_alice_core_app_bootstrap.UpdateProposal, error)
	grpc1.ClientStream
}

type bootstrapListClient struct {
	grpc1.ClientStream
}

func (x *bootstrapListClient) Recv() (*stratumn_alice_core_app_bootstrap.UpdateProposal, error) {
	m := new(stratumn_alice_core_app_bootstrap.UpdateProposal)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *bootstrapClient) Complete(ctx context.Context, in *stratumn_alice_core_app_bootstrap.CompleteReq, opts ...grpc1.CallOption) (*stratumn_alice_core_app_bootstrap.Ack, error) {
	out := new(stratumn_alice_core_app_bootstrap.Ack)
	err := grpc1.Invoke(ctx, "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/Complete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Bootstrap service

type BootstrapServer interface {
	// Propose adding a node to the network.
	AddNode(context.Context, *stratumn_alice_core_app_bootstrap.NodeIdentity) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// Propose removing a node from the network.
	RemoveNode(context.Context, *stratumn_alice_core_app_bootstrap.NodeIdentity) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// Accept a proposal to add or remove a network node.
	Accept(context.Context, *stratumn_alice_core_app_bootstrap.PeerID) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// Reject a proposal to add or remove a network node.
	Reject(context.Context, *stratumn_alice_core_app_bootstrap.PeerID) (*stratumn_alice_core_app_bootstrap.Ack, error)
	// List pending proposals to add or remove a network node.
	List(*stratumn_alice_core_app_bootstrap.Filter, Bootstrap_ListServer) error
	// Complete the network bootstrap phase.
	Complete(context.Context, *stratumn_alice_core_app_bootstrap.CompleteReq) (*stratumn_alice_core_app_bootstrap.Ack, error)
}

func RegisterBootstrapServer(s *grpc1.Server, srv BootstrapServer) {
	s.RegisterService(&_Bootstrap_serviceDesc, srv)
}

func _Bootstrap_AddNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_core_app_bootstrap.NodeIdentity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).AddNode(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/AddNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).AddNode(ctx, req.(*stratumn_alice_core_app_bootstrap.NodeIdentity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_RemoveNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_core_app_bootstrap.NodeIdentity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).RemoveNode(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/RemoveNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).RemoveNode(ctx, req.(*stratumn_alice_core_app_bootstrap.NodeIdentity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_Accept_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_core_app_bootstrap.PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).Accept(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/Accept",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).Accept(ctx, req.(*stratumn_alice_core_app_bootstrap.PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_Reject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_core_app_bootstrap.PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).Reject(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/Reject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).Reject(ctx, req.(*stratumn_alice_core_app_bootstrap.PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_List_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(stratumn_alice_core_app_bootstrap.Filter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BootstrapServer).List(m, &bootstrapListServer{stream})
}

type Bootstrap_ListServer interface {
	Send(*stratumn_alice_core_app_bootstrap.UpdateProposal) error
	grpc1.ServerStream
}

type bootstrapListServer struct {
	grpc1.ServerStream
}

func (x *bootstrapListServer) Send(m *stratumn_alice_core_app_bootstrap.UpdateProposal) error {
	return x.ServerStream.SendMsg(m)
}

func _Bootstrap_Complete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_core_app_bootstrap.CompleteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).Complete(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.core.app.bootstrap.grpc.Bootstrap/Complete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).Complete(ctx, req.(*stratumn_alice_core_app_bootstrap.CompleteReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Bootstrap_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "stratumn.alice.core.app.bootstrap.grpc.Bootstrap",
	HandlerType: (*BootstrapServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "AddNode",
			Handler:    _Bootstrap_AddNode_Handler,
		},
		{
			MethodName: "RemoveNode",
			Handler:    _Bootstrap_RemoveNode_Handler,
		},
		{
			MethodName: "Accept",
			Handler:    _Bootstrap_Accept_Handler,
		},
		{
			MethodName: "Reject",
			Handler:    _Bootstrap_Reject_Handler,
		},
		{
			MethodName: "Complete",
			Handler:    _Bootstrap_Complete_Handler,
		},
	},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "List",
			Handler:       _Bootstrap_List_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/alice/core/app/bootstrap/grpc/bootstrap.proto",
}

func init() {
	proto.RegisterFile("github.com/stratumn/alice/core/app/bootstrap/grpc/bootstrap.proto", fileDescriptorBootstrap)
}

var fileDescriptorBootstrap = []byte{
	// 415 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0xd4, 0xbf, 0x0b, 0xd3, 0x40,
	0x14, 0x07, 0x70, 0x0f, 0xa4, 0xd5, 0x8c, 0x99, 0x42, 0x86, 0x70, 0x14, 0x29, 0x15, 0xe4, 0xd2,
	0x2a, 0x76, 0xd0, 0x41, 0x52, 0x45, 0x2c, 0x88, 0x94, 0x82, 0x8b, 0xdb, 0xe5, 0xee, 0xd9, 0xc6,
	0x26, 0x79, 0xe7, 0xe5, 0xea, 0x8f, 0x35, 0xb3, 0x38, 0xe8, 0xe2, 0xec, 0xff, 0x90, 0xff, 0xc1,
	0xd1, 0x3f, 0x41, 0xea, 0x3f, 0x22, 0x97, 0x18, 0x5b, 0x1d, 0x6c, 0x32, 0x74, 0x48, 0x20, 0xc9,
	0xfb, 0xbe, 0xf7, 0x21, 0xbc, 0xc4, 0x89, 0x36, 0x89, 0xd9, 0xee, 0x63, 0x26, 0x30, 0x0b, 0x0b,
	0xa3, 0xb9, 0xd9, 0x67, 0x79, 0xc8, 0xd3, 0x44, 0x40, 0x28, 0x50, 0x43, 0xc8, 0x95, 0x0a, 0x63,
	0x44, 0x63, 0x9f, 0xa9, 0x70, 0xa3, 0x95, 0x38, 0x5e, 0x32, 0xa5, 0xd1, 0xa0, 0x3b, 0x6e, 0x73,
	0xac, 0xce, 0x31, 0x9b, 0x63, 0x5c, 0x29, 0x76, 0x2c, 0xb4, 0x39, 0x7f, 0xfa, 0x9f, 0x51, 0x69,
	0xd2, 0xf4, 0x86, 0x77, 0xc6, 0x1e, 0x4d, 0x67, 0xff, 0x41, 0x2f, 0x9c, 0x8a, 0xff, 0xa5, 0xdd,
	0xfe, 0x38, 0x74, 0xae, 0x2f, 0xda, 0x7b, 0xee, 0x07, 0xe2, 0x0c, 0x23, 0x29, 0x9f, 0xa1, 0x04,
	0x37, 0x64, 0xe7, 0xd5, 0xb6, 0x70, 0x29, 0x21, 0x37, 0x89, 0x79, 0xef, 0x8f, 0x3b, 0x04, 0x22,
	0xb1, 0x1b, 0x4d, 0xca, 0xca, 0xbb, 0xb1, 0xd2, 0xa8, 0xb0, 0x00, 0xca, 0xa5, 0x4c, 0xf2, 0x0d,
	0xe5, 0x34, 0x47, 0x09, 0xd4, 0x20, 0x35, 0x5b, 0xa0, 0x39, 0x98, 0xb7, 0xa8, 0x77, 0xee, 0x67,
	0xe2, 0x38, 0x6b, 0xc8, 0xf0, 0x0d, 0x5c, 0x56, 0x74, 0xab, 0xac, 0xbc, 0x49, 0x2b, 0xd2, 0x76,
	0xe0, 0x89, 0xe9, 0xa5, 0xc6, 0xec, 0x2f, 0xd5, 0x27, 0xe2, 0x0c, 0x22, 0x21, 0x40, 0x19, 0xf7,
	0x66, 0x87, 0x01, 0x2b, 0x00, 0xbd, 0x7c, 0xd4, 0xd9, 0x72, 0xb7, 0xac, 0xbc, 0x59, 0xd3, 0x9e,
	0x72, 0xaa, 0x6a, 0x14, 0x4f, 0xed, 0x9b, 0xe1, 0x52, 0x52, 0xd4, 0x8d, 0x0f, 0xac, 0xae, 0xf1,
	0xd4, 0xca, 0x1a, 0xb5, 0x86, 0x57, 0x20, 0x2e, 0x87, 0x6a, 0xda, 0xf7, 0x41, 0x7d, 0x25, 0xce,
	0xd5, 0xa7, 0x49, 0xd1, 0x8d, 0xf4, 0x38, 0x49, 0x0d, 0x68, 0x7f, 0xd6, 0xa1, 0xf4, 0xb9, 0x92,
	0xdc, 0xc0, 0xea, 0x37, 0x63, 0x74, 0xaf, 0xac, 0xbc, 0xb9, 0x9d, 0x43, 0x15, 0xe4, 0xf5, 0x3a,
	0xb5, 0xc2, 0xe2, 0x1c, 0x71, 0x4a, 0xec, 0xce, 0x5f, 0x7b, 0x88, 0x99, 0x4a, 0xc1, 0x80, 0xcb,
	0x3a, 0x4c, 0x6f, 0x8b, 0xd7, 0xf0, 0xba, 0xdf, 0xce, 0xb7, 0xc1, 0xd3, 0x6d, 0xa2, 0x7f, 0xea,
	0xa8, 0xda, 0xf2, 0x02, 0x16, 0x4f, 0xbe, 0x1d, 0x02, 0xf2, 0xfd, 0x10, 0x90, 0x1f, 0x87, 0x80,
	0x7c, 0xf9, 0x19, 0x5c, 0x79, 0x31, 0xef, 0xfd, 0x03, 0xba, 0x6f, 0x4f, 0xf1, 0xa0, 0xfe, 0xc2,
	0xef, 0xfc, 0x0a, 0x00, 0x00, 0xff, 0xff, 0xdf, 0xb6, 0xec, 0xa6, 0xc1, 0x04, 0x00, 0x00,
}
