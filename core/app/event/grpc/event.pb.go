// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/core/app/event/grpc/event.proto

/*
	Package grpc is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/core/app/event/grpc/event.proto

	It has these top-level messages:
		ListenReq
		Event
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

type Level int32

const (
	Level_INFO    Level = 0
	Level_DEBUG   Level = 1
	Level_WARNING Level = 2
	Level_ERROR   Level = 3
)

var Level_name = map[int32]string{
	0: "INFO",
	1: "DEBUG",
	2: "WARNING",
	3: "ERROR",
}
var Level_value = map[string]int32{
	"INFO":    0,
	"DEBUG":   1,
	"WARNING": 2,
	"ERROR":   3,
}

func (x Level) String() string {
	return proto.EnumName(Level_name, int32(x))
}
func (Level) EnumDescriptor() ([]byte, []int) { return fileDescriptorEvent, []int{0} }

// The Listen request message.
type ListenReq struct {
	Topic string `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
}

func (m *ListenReq) Reset()                    { *m = ListenReq{} }
func (m *ListenReq) String() string            { return proto.CompactTextString(m) }
func (*ListenReq) ProtoMessage()               {}
func (*ListenReq) Descriptor() ([]byte, []int) { return fileDescriptorEvent, []int{0} }

func (m *ListenReq) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

// The Event message containing what should be displayed, with optional
// display customization (if supported by the listener).
type Event struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Level   Level  `protobuf:"varint,2,opt,name=level,proto3,enum=stratumn.indigonode.core.app.event.grpc.Level" json:"level,omitempty"`
	Topic   string `protobuf:"bytes,3,opt,name=topic,proto3" json:"topic,omitempty"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (m *Event) String() string            { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptorEvent, []int{1} }

func (m *Event) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Event) GetLevel() Level {
	if m != nil {
		return m.Level
	}
	return Level_INFO
}

func (m *Event) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func init() {
	proto.RegisterType((*ListenReq)(nil), "stratumn.indigonode.core.app.event.grpc.ListenReq")
	proto.RegisterType((*Event)(nil), "stratumn.indigonode.core.app.event.grpc.Event")
	proto.RegisterEnum("stratumn.indigonode.core.app.event.grpc.Level", Level_name, Level_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for Emitter service

type EmitterClient interface {
	// Listen starts listening to events.
	Listen(ctx context.Context, in *ListenReq, opts ...grpc1.CallOption) (Emitter_ListenClient, error)
}

type emitterClient struct {
	cc *grpc1.ClientConn
}

func NewEmitterClient(cc *grpc1.ClientConn) EmitterClient {
	return &emitterClient{cc}
}

func (c *emitterClient) Listen(ctx context.Context, in *ListenReq, opts ...grpc1.CallOption) (Emitter_ListenClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Emitter_serviceDesc.Streams[0], c.cc, "/stratumn.indigonode.core.app.event.grpc.Emitter/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &emitterListenClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Emitter_ListenClient interface {
	Recv() (*Event, error)
	grpc1.ClientStream
}

type emitterListenClient struct {
	grpc1.ClientStream
}

func (x *emitterListenClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Emitter service

type EmitterServer interface {
	// Listen starts listening to events.
	Listen(*ListenReq, Emitter_ListenServer) error
}

func RegisterEmitterServer(s *grpc1.Server, srv EmitterServer) {
	s.RegisterService(&_Emitter_serviceDesc, srv)
}

func _Emitter_Listen_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(ListenReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EmitterServer).Listen(m, &emitterListenServer{stream})
}

type Emitter_ListenServer interface {
	Send(*Event) error
	grpc1.ServerStream
}

type emitterListenServer struct {
	grpc1.ServerStream
}

func (x *emitterListenServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

var _Emitter_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "stratumn.indigonode.core.app.event.grpc.Emitter",
	HandlerType: (*EmitterServer)(nil),
	Methods:     []grpc1.MethodDesc{},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _Emitter_Listen_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/go-indigonode/core/app/event/grpc/event.proto",
}

func (m *ListenReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ListenReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Topic) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Topic)))
		i += copy(dAtA[i:], m.Topic)
	}
	return i, nil
}

func (m *Event) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Event) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Message) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Message)))
		i += copy(dAtA[i:], m.Message)
	}
	if m.Level != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintEvent(dAtA, i, uint64(m.Level))
	}
	if len(m.Topic) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintEvent(dAtA, i, uint64(len(m.Topic)))
		i += copy(dAtA[i:], m.Topic)
	}
	return i, nil
}

func encodeVarintEvent(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ListenReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.Topic)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	return n
}

func (m *Event) Size() (n int) {
	var l int
	_ = l
	l = len(m.Message)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	if m.Level != 0 {
		n += 1 + sovEvent(uint64(m.Level))
	}
	l = len(m.Topic)
	if l > 0 {
		n += 1 + l + sovEvent(uint64(l))
	}
	return n
}

func sovEvent(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozEvent(x uint64) (n int) {
	return sovEvent(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ListenReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvent
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
			return fmt.Errorf("proto: ListenReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ListenReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Topic", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Topic = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvent(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthEvent
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
func (m *Event) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvent
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
			return fmt.Errorf("proto: Event: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Event: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Message = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Level", wireType)
			}
			m.Level = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Level |= (Level(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Topic", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvent
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
				return ErrInvalidLengthEvent
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Topic = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvent(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthEvent
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
func skipEvent(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvent
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
					return 0, ErrIntOverflowEvent
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
					return 0, ErrIntOverflowEvent
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
				return 0, ErrInvalidLengthEvent
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowEvent
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
				next, err := skipEvent(dAtA[start:])
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
	ErrInvalidLengthEvent = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvent   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-indigonode/core/app/event/grpc/event.proto", fileDescriptorEvent)
}

var fileDescriptorEvent = []byte{
	// 441 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0x4f, 0x8b, 0xd3, 0x40,
	0x14, 0xdf, 0xe9, 0x9a, 0xad, 0x9d, 0x05, 0xa9, 0x23, 0xc2, 0xb0, 0x87, 0xf2, 0xe8, 0x65, 0x57,
	0xc1, 0x89, 0x54, 0x11, 0x59, 0x41, 0xd8, 0x6a, 0x5d, 0x0a, 0xa5, 0x0b, 0x01, 0x11, 0xbc, 0xa5,
	0xc9, 0x6c, 0x1a, 0x48, 0xf3, 0xc6, 0xc9, 0x74, 0xd1, 0xab, 0xac, 0x1f, 0xc0, 0x9b, 0xe7, 0x9e,
	0xfc, 0x02, 0x9e, 0xf2, 0x05, 0x3c, 0xfa, 0x11, 0xa4, 0x7e, 0x11, 0x99, 0x99, 0x56, 0x7b, 0x0c,
	0x7b, 0x48, 0x48, 0xde, 0xfc, 0xfe, 0xcc, 0xfb, 0xbd, 0x47, 0xcf, 0xb2, 0xdc, 0xcc, 0x97, 0x33,
	0x91, 0xe0, 0x22, 0xac, 0x8c, 0x8e, 0xcd, 0x72, 0x51, 0x86, 0x19, 0x3e, 0xca, 0xcb, 0x34, 0xcf,
	0xb0, 0xc4, 0x54, 0x86, 0x09, 0x6a, 0x19, 0xc6, 0x4a, 0x85, 0xf2, 0x4a, 0x96, 0x26, 0xcc, 0xb4,
	0x4a, 0xfc, 0xa7, 0x50, 0x1a, 0x0d, 0xb2, 0xe3, 0x2d, 0x4f, 0xfc, 0x27, 0x09, 0x4b, 0x12, 0xb1,
	0x52, 0xc2, 0x23, 0x2d, 0xe9, 0xe8, 0x79, 0x03, 0xaf, 0x22, 0xdf, 0x18, 0x7c, 0x34, 0xf6, 0xf1,
	0x16, 0xfd, 0x57, 0xb4, 0x33, 0xc9, 0x2b, 0x23, 0xcb, 0x48, 0x7e, 0x60, 0xcf, 0x68, 0x60, 0x50,
	0xe5, 0x09, 0x27, 0x40, 0x4e, 0x3a, 0x43, 0x58, 0xd5, 0x9c, 0x8f, 0xac, 0x0b, 0xb8, 0x32, 0x18,
	0x84, 0xc2, 0x81, 0xc1, 0xe0, 0xf7, 0x9a, 0x93, 0xc8, 0xc3, 0xfb, 0xd7, 0x2d, 0x1a, 0x38, 0x1c,
	0x1b, 0xd3, 0xf6, 0x42, 0x56, 0x55, 0x9c, 0xc9, 0x8d, 0x46, 0xb8, 0xaa, 0xf9, 0xb1, 0xd7, 0xd8,
	0x1c, 0x58, 0x95, 0x99, 0x84, 0x34, 0xaf, 0x54, 0x11, 0x7f, 0x92, 0x29, 0xa0, 0x86, 0x02, 0xb3,
	0x4c, 0xa6, 0x4e, 0x72, 0xcb, 0x67, 0xd7, 0x84, 0x06, 0x85, 0xbc, 0x92, 0x05, 0x6f, 0x01, 0x39,
	0xb9, 0x33, 0x10, 0xa2, 0x61, 0x1a, 0x62, 0x62, 0x59, 0xc3, 0x97, 0xab, 0x9a, 0x9f, 0x7a, 0x67,
	0x27, 0x03, 0x66, 0x1e, 0x1b, 0xa8, 0xe6, 0xb8, 0x2c, 0x52, 0x48, 0xb0, 0x34, 0x1a, 0x6d, 0xed,
	0xdf, 0x4d, 0x00, 0x2f, 0xdd, 0xef, 0xc6, 0x3f, 0xf2, 0xe6, 0xec, 0xc1, 0x36, 0x93, 0x7d, 0xd7,
	0xcf, 0xbd, 0x55, 0xcd, 0x0f, 0x77, 0x32, 0xd9, 0x89, 0xe1, 0xe1, 0x53, 0x1a, 0x38, 0x6b, 0x76,
	0x9b, 0xde, 0x1a, 0x4f, 0xdf, 0x5c, 0x74, 0xf7, 0x58, 0x87, 0x06, 0xaf, 0x47, 0xc3, 0xb7, 0xe7,
	0x5d, 0xc2, 0x0e, 0x69, 0xfb, 0xdd, 0x59, 0x34, 0x1d, 0x4f, 0xcf, 0xbb, 0x2d, 0x5b, 0x1f, 0x45,
	0xd1, 0x45, 0xd4, 0xdd, 0x1f, 0x7c, 0x25, 0xb4, 0x3d, 0x5a, 0xe4, 0xc6, 0x48, 0xcd, 0xbe, 0x10,
	0x7a, 0xe0, 0xc7, 0xc1, 0x06, 0xcd, 0xdb, 0xdd, 0xce, 0xef, 0xa8, 0x79, 0x44, 0xae, 0x83, 0xfe,
	0xfd, 0xcf, 0x3f, 0xf8, 0x5d, 0x4f, 0x87, 0x4b, 0xd4, 0xe0, 0x00, 0xd5, 0x63, 0x32, 0x9c, 0xfc,
	0x5c, 0xf7, 0xc8, 0xaf, 0x75, 0x8f, 0xfc, 0x5e, 0xf7, 0xc8, 0xb7, 0x3f, 0xbd, 0xbd, 0xf7, 0xa7,
	0x37, 0xda, 0xe6, 0x17, 0xf6, 0x35, 0x3b, 0x70, 0xab, 0xf6, 0xe4, 0x6f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x20, 0xb8, 0x6a, 0x97, 0x12, 0x03, 0x00, 0x00,
}
