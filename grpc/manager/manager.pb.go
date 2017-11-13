// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/stratumn/alice/grpc/manager/manager.proto

/*
Package manager is a generated protocol buffer package.

It is generated from these files:
	github.com/stratumn/alice/grpc/manager/manager.proto

It has these top-level messages:
	ListReq
	StartReq
	StopReq
	PruneReq
	Service
*/
package manager

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

type Service_Status int32

const (
	Service_STOPPED     Service_Status = 0
	Service_PRESTARTING Service_Status = 1
	Service_STARTING    Service_Status = 2
	Service_RUNNING     Service_Status = 3
	Service_STOPPING    Service_Status = 4
	Service_ERRORED     Service_Status = 5
)

var Service_Status_name = map[int32]string{
	0: "STOPPED",
	1: "PRESTARTING",
	2: "STARTING",
	3: "RUNNING",
	4: "STOPPING",
	5: "ERRORED",
}
var Service_Status_value = map[string]int32{
	"STOPPED":     0,
	"PRESTARTING": 1,
	"STARTING":    2,
	"RUNNING":     3,
	"STOPPING":    4,
	"ERRORED":     5,
}

func (x Service_Status) String() string {
	return proto.EnumName(Service_Status_name, int32(x))
}
func (Service_Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{4, 0} }

// The list request message.
type ListReq struct {
}

func (m *ListReq) Reset()                    { *m = ListReq{} }
func (m *ListReq) String() string            { return proto.CompactTextString(m) }
func (*ListReq) ProtoMessage()               {}
func (*ListReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// The start request message.
type StartReq struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *StartReq) Reset()                    { *m = StartReq{} }
func (m *StartReq) String() string            { return proto.CompactTextString(m) }
func (*StartReq) ProtoMessage()               {}
func (*StartReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *StartReq) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

// The stop request message.
type StopReq struct {
	Id    string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Prune bool   `protobuf:"varint,2,opt,name=prune" json:"prune,omitempty"`
}

func (m *StopReq) Reset()                    { *m = StopReq{} }
func (m *StopReq) String() string            { return proto.CompactTextString(m) }
func (*StopReq) ProtoMessage()               {}
func (*StopReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *StopReq) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *StopReq) GetPrune() bool {
	if m != nil {
		return m.Prune
	}
	return false
}

// The prune request message.
type PruneReq struct {
}

func (m *PruneReq) Reset()                    { *m = PruneReq{} }
func (m *PruneReq) String() string            { return proto.CompactTextString(m) }
func (*PruneReq) ProtoMessage()               {}
func (*PruneReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// The service message containing information about a service.
type Service struct {
	Id        string         `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Status    Service_Status `protobuf:"varint,2,opt,name=status,enum=stratumn.alice.grpc.manager.Service_Status" json:"status,omitempty"`
	Needs     []string       `protobuf:"bytes,3,rep,name=needs" json:"needs,omitempty"`
	Stoppable bool           `protobuf:"varint,4,opt,name=stoppable" json:"stoppable,omitempty"`
	Prunable  bool           `protobuf:"varint,5,opt,name=prunable" json:"prunable,omitempty"`
	Name      string         `protobuf:"bytes,6,opt,name=name" json:"name,omitempty"`
	Desc      string         `protobuf:"bytes,7,opt,name=desc" json:"desc,omitempty"`
}

func (m *Service) Reset()                    { *m = Service{} }
func (m *Service) String() string            { return proto.CompactTextString(m) }
func (*Service) ProtoMessage()               {}
func (*Service) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Service) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Service) GetStatus() Service_Status {
	if m != nil {
		return m.Status
	}
	return Service_STOPPED
}

func (m *Service) GetNeeds() []string {
	if m != nil {
		return m.Needs
	}
	return nil
}

func (m *Service) GetStoppable() bool {
	if m != nil {
		return m.Stoppable
	}
	return false
}

func (m *Service) GetPrunable() bool {
	if m != nil {
		return m.Prunable
	}
	return false
}

func (m *Service) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Service) GetDesc() string {
	if m != nil {
		return m.Desc
	}
	return ""
}

func init() {
	proto.RegisterType((*ListReq)(nil), "stratumn.alice.grpc.manager.ListReq")
	proto.RegisterType((*StartReq)(nil), "stratumn.alice.grpc.manager.StartReq")
	proto.RegisterType((*StopReq)(nil), "stratumn.alice.grpc.manager.StopReq")
	proto.RegisterType((*PruneReq)(nil), "stratumn.alice.grpc.manager.PruneReq")
	proto.RegisterType((*Service)(nil), "stratumn.alice.grpc.manager.Service")
	proto.RegisterEnum("stratumn.alice.grpc.manager.Service_Status", Service_Status_name, Service_Status_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Manager service

type ManagerClient interface {
	// Streams the registered services.
	List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (Manager_ListClient, error)
	// Starts a service.
	Start(ctx context.Context, in *StartReq, opts ...grpc.CallOption) (*Service, error)
	// Stops a service.
	Stop(ctx context.Context, in *StopReq, opts ...grpc.CallOption) (*Service, error)
	// Prunes services.
	Prune(ctx context.Context, in *PruneReq, opts ...grpc.CallOption) (Manager_PruneClient, error)
}

type managerClient struct {
	cc *grpc.ClientConn
}

func NewManagerClient(cc *grpc.ClientConn) ManagerClient {
	return &managerClient{cc}
}

func (c *managerClient) List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (Manager_ListClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Manager_serviceDesc.Streams[0], c.cc, "/stratumn.alice.grpc.manager.Manager/List", opts...)
	if err != nil {
		return nil, err
	}
	x := &managerListClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Manager_ListClient interface {
	Recv() (*Service, error)
	grpc.ClientStream
}

type managerListClient struct {
	grpc.ClientStream
}

func (x *managerListClient) Recv() (*Service, error) {
	m := new(Service)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *managerClient) Start(ctx context.Context, in *StartReq, opts ...grpc.CallOption) (*Service, error) {
	out := new(Service)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.manager.Manager/Start", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *managerClient) Stop(ctx context.Context, in *StopReq, opts ...grpc.CallOption) (*Service, error) {
	out := new(Service)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.manager.Manager/Stop", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *managerClient) Prune(ctx context.Context, in *PruneReq, opts ...grpc.CallOption) (Manager_PruneClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Manager_serviceDesc.Streams[1], c.cc, "/stratumn.alice.grpc.manager.Manager/Prune", opts...)
	if err != nil {
		return nil, err
	}
	x := &managerPruneClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Manager_PruneClient interface {
	Recv() (*Service, error)
	grpc.ClientStream
}

type managerPruneClient struct {
	grpc.ClientStream
}

func (x *managerPruneClient) Recv() (*Service, error) {
	m := new(Service)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Manager service

type ManagerServer interface {
	// Streams the registered services.
	List(*ListReq, Manager_ListServer) error
	// Starts a service.
	Start(context.Context, *StartReq) (*Service, error)
	// Stops a service.
	Stop(context.Context, *StopReq) (*Service, error)
	// Prunes services.
	Prune(*PruneReq, Manager_PruneServer) error
}

func RegisterManagerServer(s *grpc.Server, srv ManagerServer) {
	s.RegisterService(&_Manager_serviceDesc, srv)
}

func _Manager_List_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ManagerServer).List(m, &managerListServer{stream})
}

type Manager_ListServer interface {
	Send(*Service) error
	grpc.ServerStream
}

type managerListServer struct {
	grpc.ServerStream
}

func (x *managerListServer) Send(m *Service) error {
	return x.ServerStream.SendMsg(m)
}

func _Manager_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagerServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.manager.Manager/Start",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagerServer).Start(ctx, req.(*StartReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Manager_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagerServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.manager.Manager/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagerServer).Stop(ctx, req.(*StopReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Manager_Prune_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PruneReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ManagerServer).Prune(m, &managerPruneServer{stream})
}

type Manager_PruneServer interface {
	Send(*Service) error
	grpc.ServerStream
}

type managerPruneServer struct {
	grpc.ServerStream
}

func (x *managerPruneServer) Send(m *Service) error {
	return x.ServerStream.SendMsg(m)
}

var _Manager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.alice.grpc.manager.Manager",
	HandlerType: (*ManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Start",
			Handler:    _Manager_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _Manager_Stop_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "List",
			Handler:       _Manager_List_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Prune",
			Handler:       _Manager_Prune_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/alice/grpc/manager/manager.proto",
}

func init() {
	proto.RegisterFile("github.com/stratumn/alice/grpc/manager/manager.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 522 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0x4f, 0x8f, 0xd2, 0x40,
	0x14, 0xb7, 0x40, 0x29, 0xbc, 0x35, 0xbb, 0x64, 0xd4, 0x64, 0x44, 0x13, 0x27, 0x75, 0x4d, 0x30,
	0xea, 0xb0, 0x59, 0xbd, 0x79, 0x12, 0x21, 0x66, 0x13, 0x85, 0x66, 0x8a, 0x17, 0xe3, 0x65, 0xa0,
	0xb3, 0x6c, 0x93, 0xd2, 0x76, 0x3b, 0xd3, 0x8d, 0x5e, 0xf9, 0x28, 0x7b, 0xf2, 0x03, 0xc8, 0x89,
	0x6f, 0xe0, 0xa7, 0x32, 0x33, 0x2d, 0xec, 0x9e, 0xa0, 0x07, 0xd2, 0x79, 0x7f, 0x7e, 0xef, 0xf7,
	0xde, 0x9b, 0xdf, 0x00, 0x1f, 0x16, 0xa1, 0xba, 0xca, 0x67, 0x74, 0x9e, 0x2c, 0xfb, 0x52, 0x65,
	0x5c, 0xe5, 0xcb, 0xb8, 0xcf, 0xa3, 0x70, 0x2e, 0xfa, 0x8b, 0x2c, 0x9d, 0xf7, 0x97, 0x3c, 0xe6,
	0x0b, 0x91, 0x6d, 0xbf, 0x34, 0xcd, 0x12, 0x95, 0xa0, 0x67, 0xdb, 0x54, 0x6a, 0x52, 0xa9, 0x4e,
	0xa5, 0x65, 0x4a, 0xf7, 0xed, 0x81, 0x92, 0xe2, 0x97, 0xd2, 0xbf, 0xa2, 0x94, 0xdb, 0x06, 0xe7,
	0x6b, 0x28, 0x15, 0x13, 0xd7, 0x2e, 0x85, 0x96, 0xaf, 0x78, 0xa6, 0xcf, 0xc8, 0x85, 0x5a, 0x18,
	0x60, 0x8b, 0x58, 0xbd, 0xf6, 0x00, 0xdd, 0x6e, 0x30, 0xf8, 0x22, 0xbb, 0x09, 0xe7, 0x82, 0x5c,
	0x0c, 0xff, 0x6c, 0xb0, 0xc5, 0x6a, 0x61, 0xe0, 0x66, 0xe0, 0xf8, 0x2a, 0x49, 0x2b, 0xa6, 0xa3,
	0xcf, 0x60, 0xa7, 0x59, 0x1e, 0x0b, 0x5c, 0x23, 0x56, 0xaf, 0x35, 0x78, 0x77, 0xbb, 0xc1, 0xaf,
	0x3d, 0xed, 0x20, 0xb2, 0x48, 0x96, 0x84, 0x5f, 0x2a, 0x91, 0x11, 0xa9, 0x92, 0x34, 0x0d, 0xe3,
	0x05, 0x51, 0x57, 0xbb, 0x18, 0x2b, 0xb0, 0x2e, 0x40, 0xcb, 0x60, 0x74, 0xbf, 0x7f, 0x6b, 0xe0,
	0x94, 0x3c, 0xe8, 0xf8, 0xae, 0x81, 0x92, 0xac, 0x29, 0x15, 0x57, 0xb9, 0x34, 0x6c, 0xc7, 0xe7,
	0x6f, 0xe8, 0x9e, 0x95, 0xd1, 0xb2, 0x0a, 0xf5, 0x0d, 0x84, 0x95, 0x50, 0xf4, 0x18, 0xec, 0x58,
	0x88, 0x40, 0xe2, 0x3a, 0xa9, 0xf7, 0xda, 0xac, 0x30, 0xd0, 0x73, 0x68, 0x9b, 0x0e, 0xf9, 0x2c,
	0x12, 0xb8, 0xa1, 0x67, 0x61, 0x77, 0x0e, 0xd4, 0x85, 0x96, 0xee, 0xd4, 0x04, 0x6d, 0x13, 0xdc,
	0xd9, 0x08, 0x41, 0x23, 0xe6, 0x4b, 0x81, 0x9b, 0xa6, 0x4d, 0x73, 0xd6, 0xbe, 0x40, 0xc8, 0x39,
	0x76, 0x0a, 0x9f, 0x3e, 0xbb, 0x3f, 0xa1, 0x59, 0x74, 0x82, 0x8e, 0xc0, 0xf1, 0xa7, 0x13, 0xcf,
	0x1b, 0x0d, 0x3b, 0x0f, 0xd0, 0x09, 0x1c, 0x79, 0x6c, 0xe4, 0x4f, 0x3f, 0xb1, 0xe9, 0xc5, 0xf8,
	0x4b, 0xc7, 0x42, 0x0f, 0xa1, 0xb5, 0xb3, 0x6a, 0x3a, 0x97, 0x7d, 0x1f, 0x8f, 0xb5, 0x51, 0x2f,
	0x42, 0x13, 0xcf, 0xd3, 0x56, 0x43, 0x87, 0x46, 0x8c, 0x4d, 0xd8, 0x68, 0xd8, 0xb1, 0xcf, 0xff,
	0xd5, 0xc1, 0xf9, 0x56, 0x0c, 0x8e, 0x7e, 0x43, 0x43, 0xdf, 0x3e, 0x3a, 0xdd, 0xbb, 0x9e, 0x52,
	0x20, 0xdd, 0xd3, 0x2a, 0x4b, 0x74, 0x5f, 0xae, 0xd6, 0xf8, 0x85, 0x86, 0x10, 0x1e, 0x45, 0xe6,
	0x0e, 0xf9, 0x0d, 0x0f, 0x23, 0xbd, 0x81, 0xdd, 0x4d, 0x9f, 0x59, 0x28, 0x04, 0xdb, 0xa8, 0x0d,
	0xbd, 0xda, 0x5f, 0xb5, 0x54, 0x64, 0x45, 0xf2, 0x47, 0xab, 0x35, 0x3e, 0x31, 0x18, 0xc2, 0xb7,
	0x6c, 0xe8, 0x12, 0x1a, 0x5a, 0xa8, 0x07, 0xa6, 0x2c, 0xb5, 0x5c, 0x91, 0x08, 0xad, 0xd6, 0xf8,
	0x58, 0x43, 0xee, 0xf1, 0x5c, 0x83, 0x6d, 0xc4, 0x79, 0x60, 0xa4, 0xad, 0x80, 0x2b, 0x32, 0x3d,
	0x5d, 0xad, 0xf1, 0x93, 0xe2, 0xa1, 0xe4, 0x71, 0x2e, 0x45, 0x70, 0x6f, 0x8b, 0x83, 0xb3, 0x1f,
	0xb4, 0xda, 0x3f, 0xc8, 0xc7, 0xf2, 0x3b, 0x6b, 0x9a, 0x77, 0xff, 0xfe, 0x7f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xe3, 0x8f, 0x65, 0x30, 0x7a, 0x04, 0x00, 0x00,
}