// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/core/app/metrics/grpc/metrics.proto

/*
	Package grpc is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/core/app/metrics/grpc/metrics.proto

	It has these top-level messages:
		BandwidthReq
		BandwidthStats
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/go-indigonode/cli/grpc/ext"

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

// The bandwidth request message.
type BandwidthReq struct {
	PeerId     []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	ProtocolId string `protobuf:"bytes,2,opt,name=protocol_id,json=protocolId,proto3" json:"protocol_id,omitempty"`
}

func (m *BandwidthReq) Reset()                    { *m = BandwidthReq{} }
func (m *BandwidthReq) String() string            { return proto.CompactTextString(m) }
func (*BandwidthReq) ProtoMessage()               {}
func (*BandwidthReq) Descriptor() ([]byte, []int) { return fileDescriptorMetrics, []int{0} }

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
	TotalIn  uint64 `protobuf:"varint,1,opt,name=total_in,json=totalIn,proto3" json:"total_in,omitempty"`
	TotalOut uint64 `protobuf:"varint,2,opt,name=total_out,json=totalOut,proto3" json:"total_out,omitempty"`
	RateIn   uint64 `protobuf:"varint,3,opt,name=rate_in,json=rateIn,proto3" json:"rate_in,omitempty"`
	RateOut  uint64 `protobuf:"varint,4,opt,name=rate_out,json=rateOut,proto3" json:"rate_out,omitempty"`
}

func (m *BandwidthStats) Reset()                    { *m = BandwidthStats{} }
func (m *BandwidthStats) String() string            { return proto.CompactTextString(m) }
func (*BandwidthStats) ProtoMessage()               {}
func (*BandwidthStats) Descriptor() ([]byte, []int) { return fileDescriptorMetrics, []int{1} }

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
	proto.RegisterType((*BandwidthReq)(nil), "stratumn.indigonode.core.app.metrics.grpc.BandwidthReq")
	proto.RegisterType((*BandwidthStats)(nil), "stratumn.indigonode.core.app.metrics.grpc.BandwidthStats")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for Metrics service

type MetricsClient interface {
	// Returns bandwidth usage.
	Bandwidth(ctx context.Context, in *BandwidthReq, opts ...grpc1.CallOption) (*BandwidthStats, error)
}

type metricsClient struct {
	cc *grpc1.ClientConn
}

func NewMetricsClient(cc *grpc1.ClientConn) MetricsClient {
	return &metricsClient{cc}
}

func (c *metricsClient) Bandwidth(ctx context.Context, in *BandwidthReq, opts ...grpc1.CallOption) (*BandwidthStats, error) {
	out := new(BandwidthStats)
	err := grpc1.Invoke(ctx, "/stratumn.indigonode.core.app.metrics.grpc.Metrics/Bandwidth", in, out, c.cc, opts...)
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

func RegisterMetricsServer(s *grpc1.Server, srv MetricsServer) {
	s.RegisterService(&_Metrics_serviceDesc, srv)
}

func _Metrics_Bandwidth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(BandwidthReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).Bandwidth(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.indigonode.core.app.metrics.grpc.Metrics/Bandwidth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).Bandwidth(ctx, req.(*BandwidthReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Metrics_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "stratumn.indigonode.core.app.metrics.grpc.Metrics",
	HandlerType: (*MetricsServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "Bandwidth",
			Handler:    _Metrics_Bandwidth_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "github.com/stratumn/go-indigonode/core/app/metrics/grpc/metrics.proto",
}

func (m *BandwidthReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BandwidthReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if len(m.ProtocolId) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(len(m.ProtocolId)))
		i += copy(dAtA[i:], m.ProtocolId)
	}
	return i, nil
}

func (m *BandwidthStats) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BandwidthStats) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.TotalIn != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(m.TotalIn))
	}
	if m.TotalOut != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(m.TotalOut))
	}
	if m.RateIn != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(m.RateIn))
	}
	if m.RateOut != 0 {
		dAtA[i] = 0x20
		i++
		i = encodeVarintMetrics(dAtA, i, uint64(m.RateOut))
	}
	return i, nil
}

func encodeVarintMetrics(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *BandwidthReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovMetrics(uint64(l))
	}
	l = len(m.ProtocolId)
	if l > 0 {
		n += 1 + l + sovMetrics(uint64(l))
	}
	return n
}

func (m *BandwidthStats) Size() (n int) {
	var l int
	_ = l
	if m.TotalIn != 0 {
		n += 1 + sovMetrics(uint64(m.TotalIn))
	}
	if m.TotalOut != 0 {
		n += 1 + sovMetrics(uint64(m.TotalOut))
	}
	if m.RateIn != 0 {
		n += 1 + sovMetrics(uint64(m.RateIn))
	}
	if m.RateOut != 0 {
		n += 1 + sovMetrics(uint64(m.RateOut))
	}
	return n
}

func sovMetrics(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMetrics(x uint64) (n int) {
	return sovMetrics(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *BandwidthReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetrics
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
			return fmt.Errorf("proto: BandwidthReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BandwidthReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
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
				return ErrInvalidLengthMetrics
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
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMetrics
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ProtocolId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMetrics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetrics
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
func (m *BandwidthStats) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMetrics
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
			return fmt.Errorf("proto: BandwidthStats: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BandwidthStats: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalIn", wireType)
			}
			m.TotalIn = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalIn |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalOut", wireType)
			}
			m.TotalOut = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalOut |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RateIn", wireType)
			}
			m.RateIn = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RateIn |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RateOut", wireType)
			}
			m.RateOut = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMetrics
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RateOut |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMetrics(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMetrics
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
func skipMetrics(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMetrics
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
					return 0, ErrIntOverflowMetrics
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
					return 0, ErrIntOverflowMetrics
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
				return 0, ErrInvalidLengthMetrics
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMetrics
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
				next, err := skipMetrics(dAtA[start:])
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
	ErrInvalidLengthMetrics = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMetrics   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-indigonode/core/app/metrics/grpc/metrics.proto", fileDescriptorMetrics)
}

var fileDescriptorMetrics = []byte{
	// 396 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x72, 0x4d, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x2f, 0x2e, 0x29, 0x4a, 0x2c, 0x29, 0xcd, 0xcd, 0xd3,
	0x4f, 0xcf, 0xd7, 0xcd, 0xcc, 0x4b, 0xc9, 0x4c, 0xcf, 0xcf, 0xcb, 0x4f, 0x49, 0xd5, 0x4f, 0xce,
	0x2f, 0x4a, 0xd5, 0x4f, 0x2c, 0x28, 0xd0, 0xcf, 0x4d, 0x2d, 0x29, 0xca, 0x4c, 0x2e, 0xd6, 0x4f,
	0x2f, 0x2a, 0x48, 0x86, 0x71, 0xf4, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x34, 0x61, 0x7a, 0xf5,
	0x10, 0x1a, 0xf5, 0x40, 0x1a, 0xf5, 0x12, 0x0b, 0x0a, 0xf4, 0x60, 0x6a, 0x41, 0x1a, 0xa5, 0x2c,
	0x88, 0xb0, 0x31, 0x27, 0x13, 0x62, 0x49, 0x6a, 0x45, 0x09, 0x08, 0x43, 0x2c, 0x51, 0x6a, 0x62,
	0xe4, 0xe2, 0x71, 0x4a, 0xcc, 0x4b, 0x29, 0xcf, 0x4c, 0x29, 0xc9, 0x08, 0x4a, 0x2d, 0x14, 0x32,
	0xe1, 0x62, 0x2f, 0x48, 0x4d, 0x2d, 0x8a, 0xcf, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x71,
	0x92, 0x5e, 0xb4, 0x5b, 0x42, 0x38, 0x38, 0x23, 0xbf, 0x5c, 0xa1, 0xb4, 0x38, 0x31, 0x3d, 0x55,
	0x21, 0x2d, 0xbf, 0x48, 0x01, 0xa4, 0xe2, 0xc3, 0x6e, 0x09, 0xc6, 0x20, 0x36, 0x10, 0xcb, 0x33,
	0x45, 0xc8, 0x86, 0x8b, 0x1b, 0x6c, 0x5e, 0x72, 0x7e, 0x0e, 0x48, 0x27, 0x93, 0x02, 0xa3, 0x06,
	0x27, 0x58, 0xa7, 0x38, 0xba, 0x4e, 0xa8, 0xaa, 0x20, 0x2e, 0x18, 0xcb, 0x33, 0x45, 0x69, 0x0a,
	0x23, 0x17, 0x1f, 0xdc, 0x11, 0xc1, 0x25, 0x89, 0x25, 0xc5, 0x42, 0xf2, 0x5c, 0x1c, 0x25, 0xf9,
	0x25, 0x89, 0x39, 0xf1, 0x99, 0x79, 0x60, 0x77, 0xb0, 0x38, 0xb1, 0x4c, 0xd8, 0x2b, 0xc1, 0x18,
	0xc4, 0x0e, 0x16, 0xf5, 0xcc, 0x13, 0x52, 0xe4, 0xe2, 0x84, 0x28, 0xc8, 0x2f, 0x2d, 0x01, 0xdb,
	0x07, 0x53, 0x01, 0xd1, 0xe7, 0x5f, 0x5a, 0x22, 0x24, 0xcb, 0xc5, 0x5e, 0x94, 0x58, 0x92, 0x0a,
	0x32, 0x82, 0x19, 0xa2, 0x60, 0x06, 0x48, 0x01, 0x1b, 0x48, 0xd0, 0x33, 0x0f, 0x64, 0x05, 0x58,
	0x1a, 0x64, 0x00, 0x0b, 0x92, 0x3c, 0x58, 0x93, 0x7f, 0x69, 0x89, 0xd1, 0x0a, 0x46, 0x2e, 0x76,
	0x5f, 0x48, 0x30, 0x0b, 0xcd, 0x61, 0xe4, 0xe2, 0x84, 0x3b, 0x51, 0xc8, 0x5c, 0x8f, 0xe8, 0xb8,
	0xd1, 0x43, 0x0e, 0x5d, 0x29, 0x4b, 0x72, 0x34, 0x82, 0x43, 0x44, 0x49, 0xba, 0x69, 0xab, 0x84,
	0xb8, 0x4b, 0x66, 0x71, 0x41, 0x4e, 0x62, 0xa5, 0x42, 0x12, 0x4c, 0x0e, 0x12, 0xb4, 0x4e, 0x7e,
	0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3, 0x8c, 0xc7, 0x72,
	0x0c, 0x51, 0x36, 0x64, 0x26, 0x42, 0x6b, 0x10, 0x91, 0xc4, 0x06, 0x8e, 0x1d, 0x63, 0x40, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x0d, 0x66, 0x15, 0x08, 0xcb, 0x02, 0x00, 0x00,
}
