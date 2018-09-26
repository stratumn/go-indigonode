// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-node/core/app/bootstrap/grpc/bootstrap.proto

package grpc // import "github.com/stratumn/go-node/core/app/bootstrap/grpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/go-node/cli/grpc/ext"
import pb "github.com/stratumn/go-node/core/app/bootstrap/pb"

import context "context"
import grpc "google.golang.org/grpc"

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
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Bootstrap service

type BootstrapClient interface {
	// Propose adding a node to the network.
	AddNode(ctx context.Context, in *pb.NodeIdentity, opts ...grpc.CallOption) (*pb.Ack, error)
	// Propose removing a node from the network.
	RemoveNode(ctx context.Context, in *pb.NodeIdentity, opts ...grpc.CallOption) (*pb.Ack, error)
	// Accept a proposal to add or remove a network node.
	Accept(ctx context.Context, in *pb.PeerID, opts ...grpc.CallOption) (*pb.Ack, error)
	// Reject a proposal to add or remove a network node.
	Reject(ctx context.Context, in *pb.PeerID, opts ...grpc.CallOption) (*pb.Ack, error)
	// List pending proposals to add or remove a network node.
	List(ctx context.Context, in *pb.Filter, opts ...grpc.CallOption) (Bootstrap_ListClient, error)
	// Complete the network bootstrap phase.
	Complete(ctx context.Context, in *pb.CompleteReq, opts ...grpc.CallOption) (*pb.Ack, error)
}

type bootstrapClient struct {
	cc *grpc.ClientConn
}

func NewBootstrapClient(cc *grpc.ClientConn) BootstrapClient {
	return &bootstrapClient{cc}
}

func (c *bootstrapClient) AddNode(ctx context.Context, in *pb.NodeIdentity, opts ...grpc.CallOption) (*pb.Ack, error) {
	out := new(pb.Ack)
	err := c.cc.Invoke(ctx, "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/AddNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) RemoveNode(ctx context.Context, in *pb.NodeIdentity, opts ...grpc.CallOption) (*pb.Ack, error) {
	out := new(pb.Ack)
	err := c.cc.Invoke(ctx, "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/RemoveNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) Accept(ctx context.Context, in *pb.PeerID, opts ...grpc.CallOption) (*pb.Ack, error) {
	out := new(pb.Ack)
	err := c.cc.Invoke(ctx, "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/Accept", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) Reject(ctx context.Context, in *pb.PeerID, opts ...grpc.CallOption) (*pb.Ack, error) {
	out := new(pb.Ack)
	err := c.cc.Invoke(ctx, "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/Reject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bootstrapClient) List(ctx context.Context, in *pb.Filter, opts ...grpc.CallOption) (Bootstrap_ListClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Bootstrap_serviceDesc.Streams[0], "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/List", opts...)
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
	Recv() (*pb.UpdateProposal, error)
	grpc.ClientStream
}

type bootstrapListClient struct {
	grpc.ClientStream
}

func (x *bootstrapListClient) Recv() (*pb.UpdateProposal, error) {
	m := new(pb.UpdateProposal)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *bootstrapClient) Complete(ctx context.Context, in *pb.CompleteReq, opts ...grpc.CallOption) (*pb.Ack, error) {
	out := new(pb.Ack)
	err := c.cc.Invoke(ctx, "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/Complete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Bootstrap service

type BootstrapServer interface {
	// Propose adding a node to the network.
	AddNode(context.Context, *pb.NodeIdentity) (*pb.Ack, error)
	// Propose removing a node from the network.
	RemoveNode(context.Context, *pb.NodeIdentity) (*pb.Ack, error)
	// Accept a proposal to add or remove a network node.
	Accept(context.Context, *pb.PeerID) (*pb.Ack, error)
	// Reject a proposal to add or remove a network node.
	Reject(context.Context, *pb.PeerID) (*pb.Ack, error)
	// List pending proposals to add or remove a network node.
	List(*pb.Filter, Bootstrap_ListServer) error
	// Complete the network bootstrap phase.
	Complete(context.Context, *pb.CompleteReq) (*pb.Ack, error)
}

func RegisterBootstrapServer(s *grpc.Server, srv BootstrapServer) {
	s.RegisterService(&_Bootstrap_serviceDesc, srv)
}

func _Bootstrap_AddNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.NodeIdentity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).AddNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/AddNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).AddNode(ctx, req.(*pb.NodeIdentity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_RemoveNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.NodeIdentity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).RemoveNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/RemoveNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).RemoveNode(ctx, req.(*pb.NodeIdentity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_Accept_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).Accept(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/Accept",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).Accept(ctx, req.(*pb.PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_Reject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.PeerID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).Reject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/Reject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).Reject(ctx, req.(*pb.PeerID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Bootstrap_List_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(pb.Filter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BootstrapServer).List(m, &bootstrapListServer{stream})
}

type Bootstrap_ListServer interface {
	Send(*pb.UpdateProposal) error
	grpc.ServerStream
}

type bootstrapListServer struct {
	grpc.ServerStream
}

func (x *bootstrapListServer) Send(m *pb.UpdateProposal) error {
	return x.ServerStream.SendMsg(m)
}

func _Bootstrap_Complete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(pb.CompleteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BootstrapServer).Complete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.core.app.bootstrap.grpc.Bootstrap/Complete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BootstrapServer).Complete(ctx, req.(*pb.CompleteReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Bootstrap_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.node.core.app.bootstrap.grpc.Bootstrap",
	HandlerType: (*BootstrapServer)(nil),
	Methods: []grpc.MethodDesc{
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
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "List",
			Handler:       _Bootstrap_List_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/go-node/core/app/bootstrap/grpc/bootstrap.proto",
}

func init() {
	proto.RegisterFile("github.com/stratumn/go-node/core/app/bootstrap/grpc/bootstrap.proto", fileDescriptor_bootstrap_e6f1125b2fa86251)
}

var fileDescriptor_bootstrap_e6f1125b2fa86251 = []byte{
	// 417 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0xd4, 0xcd, 0xae, 0x93, 0x40,
	0x14, 0x07, 0x70, 0x27, 0x31, 0xad, 0xb2, 0x64, 0x45, 0x58, 0x90, 0x49, 0x63, 0x93, 0x2e, 0xec,
	0x50, 0x6b, 0x34, 0x46, 0x57, 0xb4, 0xc6, 0xa4, 0xc6, 0x98, 0xa6, 0x89, 0x1b, 0x77, 0x03, 0x73,
	0xa4, 0x58, 0xe0, 0x8c, 0xc3, 0xd4, 0x8f, 0x2d, 0x4b, 0x57, 0xba, 0x73, 0xe9, 0x4b, 0xf0, 0x0e,
	0x2e, 0x7d, 0x04, 0x53, 0x5f, 0xc4, 0x0c, 0xc8, 0x6d, 0xef, 0x5d, 0x5c, 0xda, 0x9b, 0x74, 0x01,
	0x09, 0x70, 0xfe, 0xe7, 0xfc, 0x42, 0x0e, 0x58, 0xf3, 0x38, 0xd1, 0xeb, 0x6d, 0xc8, 0x22, 0xcc,
	0xfc, 0x42, 0x2b, 0xae, 0xb7, 0x59, 0xee, 0xc7, 0x38, 0xce, 0x51, 0x80, 0x1f, 0xa1, 0x02, 0x9f,
	0x4b, 0xe9, 0x87, 0x88, 0xda, 0x3c, 0x95, 0x7e, 0xac, 0x64, 0xb4, 0xbf, 0x64, 0x52, 0xa1, 0x46,
	0x7b, 0xd8, 0x26, 0x99, 0x89, 0x31, 0x13, 0x63, 0x5c, 0x4a, 0xb6, 0xaf, 0x33, 0x31, 0x77, 0x7a,
	0xed, 0xac, 0x34, 0x69, 0x9a, 0xc3, 0x67, 0x6d, 0x8e, 0xa6, 0xb5, 0x1b, 0x9c, 0xe8, 0x93, 0xe1,
	0x55, 0xdd, 0xb4, 0xec, 0x5b, 0x77, 0x67, 0xed, 0x3d, 0xfb, 0x2b, 0xb1, 0xfa, 0x81, 0x10, 0xaf,
	0x51, 0x80, 0xcd, 0x58, 0x27, 0xdc, 0xd4, 0x2d, 0x04, 0xe4, 0x3a, 0xd1, 0x5f, 0xdc, 0x61, 0x77,
	0x7d, 0x10, 0x6d, 0x06, 0xa3, 0xb2, 0x72, 0xee, 0x2d, 0x15, 0x4a, 0x2c, 0x80, 0x72, 0x21, 0x92,
	0x3c, 0xa6, 0x9c, 0x9a, 0x04, 0xd5, 0x48, 0xf5, 0x1a, 0x68, 0x0e, 0xfa, 0x13, 0xaa, 0x8d, 0xfd,
	0x9d, 0x58, 0xd6, 0x0a, 0x32, 0xfc, 0x08, 0xe7, 0xf4, 0xdc, 0x2f, 0x2b, 0x67, 0xd4, 0x7a, 0x94,
	0x19, 0x77, 0x20, 0x7a, 0xa7, 0x30, 0xbb, 0x64, 0xfa, 0x46, 0xac, 0x5e, 0x10, 0x45, 0x20, 0xb5,
	0x3d, 0xea, 0xee, 0xbf, 0x04, 0x50, 0x8b, 0xe7, 0xc7, 0x4a, 0x1e, 0x95, 0x95, 0xf3, 0xa0, 0x69,
	0x4e, 0x39, 0x95, 0x35, 0x89, 0xa7, 0xe6, 0xad, 0x70, 0x21, 0x28, 0xaa, 0x46, 0x07, 0xc6, 0xd6,
	0x68, 0x6a, 0x63, 0x4d, 0x5a, 0xc1, 0x7b, 0x88, 0xce, 0x45, 0x6a, 0x9a, 0x9f, 0x42, 0xfa, 0x49,
	0xac, 0xdb, 0xaf, 0x92, 0xe2, 0x28, 0xd0, 0x8b, 0x24, 0xd5, 0xa0, 0xdc, 0x49, 0x77, 0xe5, 0x1b,
	0x29, 0xb8, 0x86, 0xe5, 0x7f, 0xc4, 0xe0, 0x69, 0x59, 0x39, 0x8f, 0xcd, 0x14, 0x2a, 0x21, 0xaf,
	0xd7, 0xa8, 0xf5, 0x15, 0x5d, 0xc0, 0x09, 0x31, 0x9b, 0x7e, 0x67, 0x8e, 0x99, 0x4c, 0x41, 0x83,
	0x3d, 0xee, 0x1e, 0xde, 0xd6, 0xae, 0xe0, 0xc3, 0x49, 0x9b, 0xde, 0xe6, 0x0e, 0xb7, 0x88, 0x5e,
	0xd4, 0x51, 0xb9, 0xe6, 0x05, 0xcc, 0x5e, 0xfe, 0xda, 0x79, 0xe4, 0xf7, 0xce, 0x23, 0x7f, 0x76,
	0x1e, 0xf9, 0xf1, 0xd7, 0xbb, 0xf5, 0xf6, 0xc9, 0x0d, 0xfe, 0x3c, 0xcf, 0xcc, 0x29, 0xec, 0xd5,
	0xdf, 0xf5, 0xc3, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x1e, 0x47, 0xc2, 0xf5, 0xbc, 0x04, 0x00,
	0x00,
}
