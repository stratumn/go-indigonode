// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/alice/grpc/event/event.proto

/*
	Package event is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/alice/grpc/event/event.proto

	It has these top-level messages:
		ListenReq
		Event
*/
package event

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/grpc/ext"

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
	Level   Level  `protobuf:"varint,2,opt,name=level,proto3,enum=stratumn.alice.grpc.event.Level" json:"level,omitempty"`
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
	proto.RegisterType((*ListenReq)(nil), "stratumn.alice.grpc.event.ListenReq")
	proto.RegisterType((*Event)(nil), "stratumn.alice.grpc.event.Event")
	proto.RegisterEnum("stratumn.alice.grpc.event.Level", Level_name, Level_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Emitter service

type EmitterClient interface {
	// Listen starts listening to events.
	Listen(ctx context.Context, in *ListenReq, opts ...grpc.CallOption) (Emitter_ListenClient, error)
}

type emitterClient struct {
	cc *grpc.ClientConn
}

func NewEmitterClient(cc *grpc.ClientConn) EmitterClient {
	return &emitterClient{cc}
}

func (c *emitterClient) Listen(ctx context.Context, in *ListenReq, opts ...grpc.CallOption) (Emitter_ListenClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Emitter_serviceDesc.Streams[0], c.cc, "/stratumn.alice.grpc.event.Emitter/Listen", opts...)
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
	grpc.ClientStream
}

type emitterListenClient struct {
	grpc.ClientStream
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

func RegisterEmitterServer(s *grpc.Server, srv EmitterServer) {
	s.RegisterService(&_Emitter_serviceDesc, srv)
}

func _Emitter_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListenReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EmitterServer).Listen(m, &emitterListenServer{stream})
}

type Emitter_ListenServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type emitterListenServer struct {
	grpc.ServerStream
}

func (x *emitterListenServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

var _Emitter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.alice.grpc.event.Emitter",
	HandlerType: (*EmitterServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _Emitter_Listen_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/alice/grpc/event/event.proto",
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
	proto.RegisterFile("github.com/stratumn/alice/grpc/event/event.proto", fileDescriptorEvent)
}

var fileDescriptorEvent = []byte{
	// 415 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xc1, 0xaa, 0xd3, 0x40,
	0x14, 0x86, 0xef, 0xdc, 0x6b, 0x6e, 0xed, 0x5c, 0x90, 0x38, 0x22, 0x8c, 0x77, 0x11, 0x0e, 0x45,
	0xf0, 0x5a, 0x24, 0x29, 0x55, 0x5c, 0xa8, 0x08, 0x46, 0x63, 0x29, 0x94, 0x16, 0x02, 0x22, 0xb8,
	0x4b, 0x93, 0x69, 0x1a, 0x48, 0x32, 0x31, 0x39, 0x2d, 0xba, 0xf5, 0x29, 0x5c, 0x67, 0xe5, 0x0b,
	0xb8, 0xca, 0x0b, 0xb8, 0xf4, 0x11, 0xa4, 0x3e, 0x83, 0x7b, 0xc9, 0x4c, 0x23, 0xdd, 0x58, 0x5d,
	0x64, 0x20, 0x67, 0xbe, 0xf3, 0xff, 0x39, 0x7f, 0x0e, 0x1d, 0xc5, 0x09, 0xae, 0x37, 0x4b, 0x3b,
	0x94, 0x99, 0x53, 0x61, 0x19, 0xe0, 0x26, 0xcb, 0x9d, 0x20, 0x4d, 0x42, 0xe1, 0xc4, 0x65, 0x11,
	0x3a, 0x62, 0x2b, 0x72, 0xd4, 0xa7, 0x5d, 0x94, 0x12, 0x25, 0xbb, 0xd3, 0x61, 0xb6, 0xc2, 0xec,
	0x16, 0xb3, 0x15, 0x70, 0xf9, 0xe0, 0x5f, 0x62, 0x1f, 0xb0, 0x7d, 0xb4, 0xd0, 0xe0, 0x25, 0xed,
	0xcf, 0x92, 0x0a, 0x45, 0xee, 0x8b, 0xf7, 0xec, 0x31, 0x35, 0x50, 0x16, 0x49, 0xc8, 0x09, 0x90,
	0xab, 0xbe, 0x0b, 0x75, 0xc3, 0xb9, 0xd7, 0x8a, 0x82, 0x2a, 0x03, 0x4a, 0x48, 0x15, 0x0c, 0x28,
	0xbf, 0x34, 0x9c, 0xf8, 0x1a, 0x1f, 0xfc, 0x22, 0xd4, 0x50, 0x1c, 0x9b, 0xd2, 0x5e, 0x26, 0xaa,
	0x2a, 0x88, 0xc5, 0x5e, 0xc3, 0xa9, 0x1b, 0x7e, 0x4f, 0x6b, 0xec, 0x2f, 0x5a, 0x95, 0xa5, 0x80,
	0x28, 0xa9, 0x8a, 0x34, 0xf8, 0x28, 0x22, 0x90, 0x25, 0xa4, 0x32, 0x8e, 0x45, 0xa4, 0x24, 0xbb,
	0x7e, 0xb6, 0xa5, 0x46, 0x2a, 0xb6, 0x22, 0xe5, 0xa7, 0x40, 0xae, 0x6e, 0x8c, 0xc1, 0xfe, 0xeb,
	0xc8, 0xf6, 0xac, 0xe5, 0xdc, 0xe7, 0x75, 0xc3, 0x9f, 0x68, 0x2b, 0xd5, 0x08, 0xb8, 0x0e, 0x10,
	0xaa, 0xb5, 0xdc, 0xa4, 0x11, 0x84, 0x32, 0xc7, 0x52, 0xb6, 0xb5, 0x3f, 0xd6, 0x20, 0x57, 0xea,
	0x75, 0x6f, 0xe8, 0x6b, 0x3b, 0x76, 0xbf, 0x0b, 0xe1, 0x4c, 0x0d, 0x70, 0xab, 0x6e, 0xf8, 0xc5,
	0x41, 0x08, 0x07, 0x73, 0x0f, 0x1f, 0x51, 0x43, 0x59, 0xb3, 0xeb, 0xf4, 0xda, 0x74, 0xfe, 0x7a,
	0x61, 0x9e, 0xb0, 0x3e, 0x35, 0x5e, 0x79, 0xee, 0x9b, 0x89, 0x49, 0xd8, 0x05, 0xed, 0xbd, 0x7d,
	0xe1, 0xcf, 0xa7, 0xf3, 0x89, 0x79, 0xda, 0xd6, 0x3d, 0xdf, 0x5f, 0xf8, 0xe6, 0xd9, 0x18, 0x69,
	0xcf, 0xcb, 0x12, 0x44, 0x51, 0xb2, 0x84, 0x9e, 0xeb, 0xf4, 0xd9, 0xdd, 0x63, 0xe3, 0x75, 0x3f,
	0xe8, 0xf2, 0x58, 0x08, 0xea, 0x1b, 0x07, 0xb7, 0x3f, 0x7d, 0xe5, 0x37, 0x75, 0x03, 0xac, 0x64,
	0x09, 0xea, 0xae, 0x1a, 0x11, 0xf7, 0xd9, 0xb7, 0x9d, 0x45, 0xbe, 0xef, 0x2c, 0xf2, 0x63, 0x67,
	0x91, 0xcf, 0x3f, 0xad, 0x93, 0x77, 0xc3, 0xff, 0xd9, 0xba, 0xa7, 0xea, 0x5c, 0x9e, 0xab, 0x6d,
	0x79, 0xf8, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xbf, 0x85, 0x03, 0xa4, 0xaa, 0x02, 0x00, 0x00,
}
