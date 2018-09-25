// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-node/app/clock/grpc/clock.proto

package grpc // import "github.com/stratumn/go-node/app/clock/grpc"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/go-node/cli/grpc/ext"

import context "context"
import grpc "google.golang.org/grpc"

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

// The Local request message.
type LocalReq struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalReq) Reset()         { *m = LocalReq{} }
func (m *LocalReq) String() string { return proto.CompactTextString(m) }
func (*LocalReq) ProtoMessage()    {}
func (*LocalReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_clock_7e1910715cd89f60, []int{0}
}
func (m *LocalReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LocalReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LocalReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *LocalReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalReq.Merge(dst, src)
}
func (m *LocalReq) XXX_Size() int {
	return m.Size()
}
func (m *LocalReq) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalReq.DiscardUnknown(m)
}

var xxx_messageInfo_LocalReq proto.InternalMessageInfo

// The Remote request message.
type RemoteReq struct {
	PeerId               []byte   `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoteReq) Reset()         { *m = RemoteReq{} }
func (m *RemoteReq) String() string { return proto.CompactTextString(m) }
func (*RemoteReq) ProtoMessage()    {}
func (*RemoteReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_clock_7e1910715cd89f60, []int{1}
}
func (m *RemoteReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RemoteReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RemoteReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *RemoteReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoteReq.Merge(dst, src)
}
func (m *RemoteReq) XXX_Size() int {
	return m.Size()
}
func (m *RemoteReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoteReq.DiscardUnknown(m)
}

var xxx_messageInfo_RemoteReq proto.InternalMessageInfo

func (m *RemoteReq) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

// The time message containing a Unix nano timestamp.
type Time struct {
	Timestamp            int64    `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Time) Reset()         { *m = Time{} }
func (m *Time) String() string { return proto.CompactTextString(m) }
func (*Time) ProtoMessage()    {}
func (*Time) Descriptor() ([]byte, []int) {
	return fileDescriptor_clock_7e1910715cd89f60, []int{2}
}
func (m *Time) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Time) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Time.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *Time) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Time.Merge(dst, src)
}
func (m *Time) XXX_Size() int {
	return m.Size()
}
func (m *Time) XXX_DiscardUnknown() {
	xxx_messageInfo_Time.DiscardUnknown(m)
}

var xxx_messageInfo_Time proto.InternalMessageInfo

func (m *Time) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func init() {
	proto.RegisterType((*LocalReq)(nil), "stratumn.node.app.clock.grpc.LocalReq")
	proto.RegisterType((*RemoteReq)(nil), "stratumn.node.app.clock.grpc.RemoteReq")
	proto.RegisterType((*Time)(nil), "stratumn.node.app.clock.grpc.Time")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Clock service

type ClockClient interface {
	// Returns the local time.
	Local(ctx context.Context, in *LocalReq, opts ...grpc.CallOption) (*Time, error)
	// Returns a peer's remote time.
	Remote(ctx context.Context, in *RemoteReq, opts ...grpc.CallOption) (*Time, error)
}

type clockClient struct {
	cc *grpc.ClientConn
}

func NewClockClient(cc *grpc.ClientConn) ClockClient {
	return &clockClient{cc}
}

func (c *clockClient) Local(ctx context.Context, in *LocalReq, opts ...grpc.CallOption) (*Time, error) {
	out := new(Time)
	err := c.cc.Invoke(ctx, "/stratumn.node.app.clock.grpc.Clock/Local", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clockClient) Remote(ctx context.Context, in *RemoteReq, opts ...grpc.CallOption) (*Time, error) {
	out := new(Time)
	err := c.cc.Invoke(ctx, "/stratumn.node.app.clock.grpc.Clock/Remote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Clock service

type ClockServer interface {
	// Returns the local time.
	Local(context.Context, *LocalReq) (*Time, error)
	// Returns a peer's remote time.
	Remote(context.Context, *RemoteReq) (*Time, error)
}

func RegisterClockServer(s *grpc.Server, srv ClockServer) {
	s.RegisterService(&_Clock_serviceDesc, srv)
}

func _Clock_Local_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocalReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClockServer).Local(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.app.clock.grpc.Clock/Local",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClockServer).Local(ctx, req.(*LocalReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Clock_Remote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClockServer).Remote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.app.clock.grpc.Clock/Remote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClockServer).Remote(ctx, req.(*RemoteReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Clock_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.node.app.clock.grpc.Clock",
	HandlerType: (*ClockServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Local",
			Handler:    _Clock_Local_Handler,
		},
		{
			MethodName: "Remote",
			Handler:    _Clock_Remote_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/stratumn/go-node/app/clock/grpc/clock.proto",
}

func (m *LocalReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LocalReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *RemoteReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RemoteReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintClock(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *Time) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Time) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Timestamp != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintClock(dAtA, i, uint64(m.Timestamp))
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintClock(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *LocalReq) Size() (n int) {
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *RemoteReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovClock(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *Time) Size() (n int) {
	var l int
	_ = l
	if m.Timestamp != 0 {
		n += 1 + sovClock(uint64(m.Timestamp))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovClock(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozClock(x uint64) (n int) {
	return sovClock(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *LocalReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowClock
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
			return fmt.Errorf("proto: LocalReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LocalReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipClock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthClock
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RemoteReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowClock
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
			return fmt.Errorf("proto: RemoteReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RemoteReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClock
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
				return ErrInvalidLengthClock
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
			skippy, err := skipClock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthClock
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Time) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowClock
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
			return fmt.Errorf("proto: Time: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Time: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowClock
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipClock(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthClock
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipClock(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowClock
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
					return 0, ErrIntOverflowClock
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
					return 0, ErrIntOverflowClock
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
				return 0, ErrInvalidLengthClock
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowClock
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
				next, err := skipClock(dAtA[start:])
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
	ErrInvalidLengthClock = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowClock   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-node/app/clock/grpc/clock.proto", fileDescriptor_clock_7e1910715cd89f60)
}

var fileDescriptor_clock_7e1910715cd89f60 = []byte{
	// 335 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x32, 0x4b, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x2f, 0x2e, 0x29, 0x4a, 0x2c, 0x29, 0xcd, 0xcd, 0xd3,
	0x4f, 0xcf, 0xd7, 0xcd, 0xcb, 0x4f, 0x49, 0xd5, 0x4f, 0x2c, 0x28, 0xd0, 0x4f, 0xce, 0xc9, 0x4f,
	0xce, 0xd6, 0x4f, 0x2f, 0x2a, 0x48, 0x86, 0x30, 0xf5, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x64,
	0x60, 0x8a, 0xf5, 0x40, 0x2a, 0xf5, 0x12, 0x0b, 0x0a, 0xf4, 0x20, 0xd2, 0x20, 0x95, 0x52, 0x46,
	0xf8, 0x4c, 0x4d, 0xce, 0xc9, 0x84, 0x98, 0x97, 0x5a, 0x51, 0x02, 0xc2, 0x10, 0x13, 0x95, 0xb8,
	0xb8, 0x38, 0x7c, 0xf2, 0x93, 0x13, 0x73, 0x82, 0x52, 0x0b, 0x95, 0x2c, 0xb9, 0x38, 0x83, 0x52,
	0x73, 0xf3, 0x4b, 0x52, 0x83, 0x52, 0x0b, 0x85, 0x74, 0xb8, 0xd8, 0x0b, 0x52, 0x53, 0x8b, 0xe2,
	0x33, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0x78, 0x9c, 0x84, 0x17, 0xed, 0x96, 0x60, 0x0f, 0x48,
	0x4d, 0x2d, 0x52, 0xf0, 0x74, 0x59, 0xb1, 0x5b, 0x82, 0xf1, 0xc3, 0x6e, 0x09, 0xc6, 0x20, 0x36,
	0x90, 0x1a, 0xcf, 0x14, 0x25, 0x2d, 0x2e, 0x96, 0x90, 0xcc, 0xdc, 0x54, 0x21, 0x25, 0x2e, 0xce,
	0x92, 0xcc, 0xdc, 0xd4, 0xe2, 0x92, 0xc4, 0xdc, 0x02, 0xb0, 0x3e, 0x66, 0x27, 0x96, 0x03, 0x7b,
	0x24, 0x18, 0x83, 0x10, 0xc2, 0x46, 0xbf, 0x18, 0xb9, 0x58, 0x9d, 0x41, 0xae, 0x16, 0x2a, 0xe3,
	0x62, 0x05, 0x5b, 0x2e, 0xa4, 0xa6, 0x87, 0xcf, 0x63, 0x7a, 0x30, 0x17, 0x4a, 0x29, 0xe1, 0x57,
	0x07, 0x72, 0x82, 0x92, 0x62, 0xd3, 0x56, 0x09, 0x59, 0x97, 0xcc, 0xe2, 0x82, 0x9c, 0xc4, 0x4a,
	0x85, 0x92, 0x8c, 0x54, 0x05, 0x90, 0x4a, 0xf5, 0x62, 0x85, 0x1c, 0x90, 0x21, 0x0a, 0x20, 0x67,
	0x08, 0x55, 0x72, 0xb1, 0x41, 0x3c, 0x2a, 0xa4, 0x8e, 0xdf, 0x40, 0x78, 0x70, 0x10, 0x65, 0xb3,
	0x52, 0xd3, 0x56, 0x09, 0x39, 0x64, 0x9b, 0x41, 0x81, 0xa3, 0x5e, 0xac, 0x50, 0x04, 0x36, 0x05,
	0x6c, 0xb5, 0x93, 0xe3, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7,
	0x38, 0xe3, 0xb1, 0x1c, 0x43, 0x94, 0x3e, 0xf1, 0x69, 0xc1, 0x1a, 0x44, 0x24, 0xb1, 0x81, 0x63,
	0xce, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xcc, 0xd5, 0x05, 0x12, 0x45, 0x02, 0x00, 0x00,
}
