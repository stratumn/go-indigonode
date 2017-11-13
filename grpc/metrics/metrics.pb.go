// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/stratumn/alice/grpc/metrics/metrics.proto

/*
Package metrics is a generated protocol buffer package.

It is generated from these files:
	github.com/stratumn/alice/grpc/metrics/metrics.proto

It has these top-level messages:
	BandwidthReq
	BandwidthStats
*/
package metrics

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

// The bandwidth request message.
type BandwidthReq struct {
	PeerId     []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	ProtocolId string `protobuf:"bytes,2,opt,name=protocol_id,json=protocolId" json:"protocol_id,omitempty"`
}

func (m *BandwidthReq) Reset()                    { *m = BandwidthReq{} }
func (m *BandwidthReq) String() string            { return proto.CompactTextString(m) }
func (*BandwidthReq) ProtoMessage()               {}
func (*BandwidthReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BandwidthReq) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

func (m *BandwidthReq) GetProtocolId() string {
	if m != nil {
		return m.ProtocolId
	}
	return ""
}

// The bandwidth stats message containing totals and rates.
type BandwidthStats struct {
	TotalIn  uint64 `protobuf:"varint,1,opt,name=total_in,json=totalIn" json:"total_in,omitempty"`
	TotalOut uint64 `protobuf:"varint,2,opt,name=total_out,json=totalOut" json:"total_out,omitempty"`
	RateIn   uint64 `protobuf:"varint,3,opt,name=rate_in,json=rateIn" json:"rate_in,omitempty"`
	RateOut  uint64 `protobuf:"varint,4,opt,name=rate_out,json=rateOut" json:"rate_out,omitempty"`
}

func (m *BandwidthStats) Reset()                    { *m = BandwidthStats{} }
func (m *BandwidthStats) String() string            { return proto.CompactTextString(m) }
func (*BandwidthStats) ProtoMessage()               {}
func (*BandwidthStats) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *BandwidthStats) GetTotalIn() uint64 {
	if m != nil {
		return m.TotalIn
	}
	return 0
}

func (m *BandwidthStats) GetTotalOut() uint64 {
	if m != nil {
		return m.TotalOut
	}
	return 0
}

func (m *BandwidthStats) GetRateIn() uint64 {
	if m != nil {
		return m.RateIn
	}
	return 0
}

func (m *BandwidthStats) GetRateOut() uint64 {
	if m != nil {
		return m.RateOut
	}
	return 0
}

func init() {
	proto.RegisterType((*BandwidthReq)(nil), "stratumn.alice.grpc.metrics.BandwidthReq")
	proto.RegisterType((*BandwidthStats)(nil), "stratumn.alice.grpc.metrics.BandwidthStats")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Metrics service

type MetricsClient interface {
	// Returns bandwidth usage.
	Bandwidth(ctx context.Context, in *BandwidthReq, opts ...grpc.CallOption) (*BandwidthStats, error)
}

type metricsClient struct {
	cc *grpc.ClientConn
}

func NewMetricsClient(cc *grpc.ClientConn) MetricsClient {
	return &metricsClient{cc}
}

func (c *metricsClient) Bandwidth(ctx context.Context, in *BandwidthReq, opts ...grpc.CallOption) (*BandwidthStats, error) {
	out := new(BandwidthStats)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.metrics.Metrics/Bandwidth", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Metrics service

type MetricsServer interface {
	// Returns bandwidth usage.
	Bandwidth(context.Context, *BandwidthReq) (*BandwidthStats, error)
}

func RegisterMetricsServer(s *grpc.Server, srv MetricsServer) {
	s.RegisterService(&_Metrics_serviceDesc, srv)
}

func _Metrics_Bandwidth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BandwidthReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).Bandwidth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.metrics.Metrics/Bandwidth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).Bandwidth(ctx, req.(*BandwidthReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Metrics_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.alice.grpc.metrics.Metrics",
	HandlerType: (*MetricsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Bandwidth",
			Handler:    _Metrics_Bandwidth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/stratumn/alice/grpc/metrics/metrics.proto",
}

func init() {
	proto.RegisterFile("github.com/stratumn/alice/grpc/metrics/metrics.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 349 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xb1, 0x4e, 0xf3, 0x30,
	0x14, 0x85, 0xe5, 0xff, 0xaf, 0x1a, 0x6a, 0x2a, 0x06, 0x33, 0x34, 0x6a, 0x85, 0x08, 0x9d, 0x8a,
	0x40, 0x0e, 0x82, 0x6e, 0x30, 0x45, 0x2c, 0x19, 0x50, 0xa5, 0x74, 0x63, 0xa9, 0xdc, 0xc4, 0xb4,
	0x96, 0xd2, 0x38, 0x38, 0x37, 0x2a, 0x6c, 0xa8, 0x33, 0x03, 0x03, 0x03, 0x33, 0xef, 0xc0, 0x54,
	0xf1, 0x4e, 0x3c, 0x02, 0xb2, 0xdd, 0x54, 0x15, 0x03, 0x74, 0x88, 0x6c, 0xdd, 0xfb, 0x9d, 0x73,
	0x6f, 0x4e, 0x82, 0xfb, 0x13, 0x01, 0xd3, 0x72, 0x4c, 0x63, 0x39, 0xf3, 0x0b, 0x50, 0x0c, 0xca,
	0x59, 0xe6, 0xb3, 0x54, 0xc4, 0xdc, 0x9f, 0xa8, 0x3c, 0xf6, 0x67, 0x1c, 0x94, 0x88, 0x8b, 0xea,
	0xa4, 0xb9, 0x92, 0x20, 0x49, 0xa7, 0x42, 0xa9, 0x41, 0xa9, 0x46, 0xe9, 0x0a, 0x69, 0x9f, 0xfe,
	0x61, 0xc9, 0x1f, 0x40, 0x3f, 0xd6, 0xaa, 0xbb, 0x40, 0xb8, 0x19, 0xb0, 0x2c, 0x99, 0x8b, 0x04,
	0xa6, 0x11, 0xbf, 0x27, 0x7d, 0xec, 0xe4, 0x9c, 0xab, 0x91, 0x48, 0x5c, 0xe4, 0xa1, 0x5e, 0x33,
	0xe8, 0xbc, 0x2f, 0xdd, 0xfd, 0xe1, 0x54, 0xce, 0xbd, 0xb2, 0x60, 0x13, 0xee, 0xdd, 0x49, 0xe5,
	0x69, 0xe2, 0x6b, 0xe9, 0xa2, 0xa8, 0xae, 0x6f, 0x61, 0x42, 0xae, 0xf0, 0xae, 0xf1, 0x8b, 0x65,
	0xaa, 0x95, 0xff, 0x3c, 0xd4, 0x6b, 0x18, 0x65, 0xeb, 0xa7, 0x72, 0x45, 0x45, 0xb8, 0xba, 0x85,
	0x49, 0xf7, 0x15, 0xe1, 0xbd, 0xf5, 0x12, 0x43, 0x60, 0x50, 0x90, 0x43, 0xbc, 0x03, 0x12, 0x58,
	0x3a, 0x12, 0x99, 0xd9, 0xa3, 0x16, 0xd4, 0x5e, 0x3e, 0x5d, 0x14, 0x39, 0xa6, 0x1a, 0x66, 0xe4,
	0x08, 0x37, 0x2c, 0x20, 0x4b, 0x30, 0xf3, 0x2a, 0xc2, 0xea, 0x06, 0x25, 0x90, 0x03, 0xec, 0x28,
	0x06, 0x5c, 0x5b, 0xfc, 0xb7, 0xc0, 0x9b, 0x06, 0xea, 0xba, 0x18, 0x66, 0x7a, 0x84, 0x69, 0x6b,
	0x83, 0xda, 0x46, 0xdf, 0x88, 0x06, 0x25, 0x9c, 0x3f, 0x23, 0xec, 0xdc, 0xd8, 0x54, 0xc9, 0x13,
	0xc2, 0x8d, 0xf5, 0x8a, 0xe4, 0x98, 0xfe, 0xf2, 0x05, 0xe8, 0x66, 0x9e, 0xed, 0x93, 0xed, 0x50,
	0xf3, 0xd6, 0xdd, 0xce, 0xe2, 0xc3, 0x6d, 0x5d, 0x8b, 0x22, 0x4f, 0xd9, 0xa3, 0x37, 0xae, 0x7a,
	0x36, 0xbe, 0xe0, 0xec, 0x96, 0x6e, 0xf7, 0xb7, 0x5c, 0xae, 0xce, 0x71, 0xdd, 0x64, 0x7c, 0xf1,
	0x1d, 0x00, 0x00, 0xff, 0xff, 0x3d, 0xda, 0x77, 0x1d, 0x66, 0x02, 0x00, 0x00,
}