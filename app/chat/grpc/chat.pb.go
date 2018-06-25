// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/app/chat/grpc/chat.proto

/*
	Package grpc is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/app/chat/grpc/chat.proto

	It has these top-level messages:
		ChatMessage
		Ack
		HistoryReq
		DatedMessage
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/gogo/protobuf/types"
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

// A chat message.
type ChatMessage struct {
	PeerId  []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *ChatMessage) Reset()                    { *m = ChatMessage{} }
func (m *ChatMessage) String() string            { return proto.CompactTextString(m) }
func (*ChatMessage) ProtoMessage()               {}
func (*ChatMessage) Descriptor() ([]byte, []int) { return fileDescriptorChat, []int{0} }

func (m *ChatMessage) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

func (m *ChatMessage) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

// An empty ack.
type Ack struct {
}

func (m *Ack) Reset()                    { *m = Ack{} }
func (m *Ack) String() string            { return proto.CompactTextString(m) }
func (*Ack) ProtoMessage()               {}
func (*Ack) Descriptor() ([]byte, []int) { return fileDescriptorChat, []int{1} }

// The request message to get the chat history with a peer.
type HistoryReq struct {
	PeerId []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
}

func (m *HistoryReq) Reset()                    { *m = HistoryReq{} }
func (m *HistoryReq) String() string            { return proto.CompactTextString(m) }
func (*HistoryReq) ProtoMessage()               {}
func (*HistoryReq) Descriptor() ([]byte, []int) { return fileDescriptorChat, []int{2} }

func (m *HistoryReq) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

type DatedMessage struct {
	From    []byte                     `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To      []byte                     `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Content string                     `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
	Time    *google_protobuf.Timestamp `protobuf:"bytes,4,opt,name=time" json:"time,omitempty"`
}

func (m *DatedMessage) Reset()                    { *m = DatedMessage{} }
func (m *DatedMessage) String() string            { return proto.CompactTextString(m) }
func (*DatedMessage) ProtoMessage()               {}
func (*DatedMessage) Descriptor() ([]byte, []int) { return fileDescriptorChat, []int{3} }

func (m *DatedMessage) GetFrom() []byte {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *DatedMessage) GetTo() []byte {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *DatedMessage) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func (m *DatedMessage) GetTime() *google_protobuf.Timestamp {
	if m != nil {
		return m.Time
	}
	return nil
}

func init() {
	proto.RegisterType((*ChatMessage)(nil), "stratumn.indigonode.app.chat.grpc.ChatMessage")
	proto.RegisterType((*Ack)(nil), "stratumn.indigonode.app.chat.grpc.Ack")
	proto.RegisterType((*HistoryReq)(nil), "stratumn.indigonode.app.chat.grpc.HistoryReq")
	proto.RegisterType((*DatedMessage)(nil), "stratumn.indigonode.app.chat.grpc.DatedMessage")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for Chat service

type ChatClient interface {
	// Sends a message to a peer.
	Message(ctx context.Context, in *ChatMessage, opts ...grpc1.CallOption) (*Ack, error)
	// Gets the chat history with a peer.
	GetHistory(ctx context.Context, in *HistoryReq, opts ...grpc1.CallOption) (Chat_GetHistoryClient, error)
}

type chatClient struct {
	cc *grpc1.ClientConn
}

func NewChatClient(cc *grpc1.ClientConn) ChatClient {
	return &chatClient{cc}
}

func (c *chatClient) Message(ctx context.Context, in *ChatMessage, opts ...grpc1.CallOption) (*Ack, error) {
	out := new(Ack)
	err := grpc1.Invoke(ctx, "/stratumn.indigonode.app.chat.grpc.Chat/Message", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) GetHistory(ctx context.Context, in *HistoryReq, opts ...grpc1.CallOption) (Chat_GetHistoryClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_Chat_serviceDesc.Streams[0], c.cc, "/stratumn.indigonode.app.chat.grpc.Chat/GetHistory", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatGetHistoryClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Chat_GetHistoryClient interface {
	Recv() (*DatedMessage, error)
	grpc1.ClientStream
}

type chatGetHistoryClient struct {
	grpc1.ClientStream
}

func (x *chatGetHistoryClient) Recv() (*DatedMessage, error) {
	m := new(DatedMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Chat service

type ChatServer interface {
	// Sends a message to a peer.
	Message(context.Context, *ChatMessage) (*Ack, error)
	// Gets the chat history with a peer.
	GetHistory(*HistoryReq, Chat_GetHistoryServer) error
}

func RegisterChatServer(s *grpc1.Server, srv ChatServer) {
	s.RegisterService(&_Chat_serviceDesc, srv)
}

func _Chat_Message_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).Message(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.indigonode.app.chat.grpc.Chat/Message",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).Message(ctx, req.(*ChatMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_GetHistory_Handler(srv interface{}, stream grpc1.ServerStream) error {
	m := new(HistoryReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatServer).GetHistory(m, &chatGetHistoryServer{stream})
}

type Chat_GetHistoryServer interface {
	Send(*DatedMessage) error
	grpc1.ServerStream
}

type chatGetHistoryServer struct {
	grpc1.ServerStream
}

func (x *chatGetHistoryServer) Send(m *DatedMessage) error {
	return x.ServerStream.SendMsg(m)
}

var _Chat_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "stratumn.indigonode.app.chat.grpc.Chat",
	HandlerType: (*ChatServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "Message",
			Handler:    _Chat_Message_Handler,
		},
	},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "GetHistory",
			Handler:       _Chat_GetHistory_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/go-indigonode/app/chat/grpc/chat.proto",
}

func (m *ChatMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ChatMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintChat(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if len(m.Message) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintChat(dAtA, i, uint64(len(m.Message)))
		i += copy(dAtA[i:], m.Message)
	}
	return i, nil
}

func (m *Ack) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Ack) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *HistoryReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HistoryReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintChat(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	return i, nil
}

func (m *DatedMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DatedMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.From) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintChat(dAtA, i, uint64(len(m.From)))
		i += copy(dAtA[i:], m.From)
	}
	if len(m.To) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintChat(dAtA, i, uint64(len(m.To)))
		i += copy(dAtA[i:], m.To)
	}
	if len(m.Content) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintChat(dAtA, i, uint64(len(m.Content)))
		i += copy(dAtA[i:], m.Content)
	}
	if m.Time != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintChat(dAtA, i, uint64(m.Time.Size()))
		n1, err := m.Time.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	return i, nil
}

func encodeVarintChat(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *ChatMessage) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovChat(uint64(l))
	}
	l = len(m.Message)
	if l > 0 {
		n += 1 + l + sovChat(uint64(l))
	}
	return n
}

func (m *Ack) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *HistoryReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovChat(uint64(l))
	}
	return n
}

func (m *DatedMessage) Size() (n int) {
	var l int
	_ = l
	l = len(m.From)
	if l > 0 {
		n += 1 + l + sovChat(uint64(l))
	}
	l = len(m.To)
	if l > 0 {
		n += 1 + l + sovChat(uint64(l))
	}
	l = len(m.Content)
	if l > 0 {
		n += 1 + l + sovChat(uint64(l))
	}
	if m.Time != nil {
		l = m.Time.Size()
		n += 1 + l + sovChat(uint64(l))
	}
	return n
}

func sovChat(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozChat(x uint64) (n int) {
	return sovChat(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ChatMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChat
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
			return fmt.Errorf("proto: ChatMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ChatMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
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
				return ErrInvalidLengthChat
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
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
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
				return ErrInvalidLengthChat
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Message = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipChat(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChat
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
func (m *Ack) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChat
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
			return fmt.Errorf("proto: Ack: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Ack: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipChat(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChat
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
func (m *HistoryReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChat
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
			return fmt.Errorf("proto: HistoryReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HistoryReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
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
				return ErrInvalidLengthChat
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
			skippy, err := skipChat(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChat
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
func (m *DatedMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowChat
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
			return fmt.Errorf("proto: DatedMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DatedMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field From", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
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
				return ErrInvalidLengthChat
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.From = append(m.From[:0], dAtA[iNdEx:postIndex]...)
			if m.From == nil {
				m.From = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field To", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
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
				return ErrInvalidLengthChat
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.To = append(m.To[:0], dAtA[iNdEx:postIndex]...)
			if m.To == nil {
				m.To = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Content", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
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
				return ErrInvalidLengthChat
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Content = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Time", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowChat
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthChat
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Time == nil {
				m.Time = &google_protobuf.Timestamp{}
			}
			if err := m.Time.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipChat(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthChat
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
func skipChat(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowChat
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
					return 0, ErrIntOverflowChat
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
					return 0, ErrIntOverflowChat
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
				return 0, ErrInvalidLengthChat
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowChat
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
				next, err := skipChat(dAtA[start:])
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
	ErrInvalidLengthChat = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowChat   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-indigonode/app/chat/grpc/chat.proto", fileDescriptorChat)
}

var fileDescriptorChat = []byte{
	// 491 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xbf, 0x6e, 0xd3, 0x40,
	0x1c, 0xc7, 0xb9, 0x34, 0xd4, 0xe2, 0x1a, 0x81, 0x30, 0x02, 0x0e, 0xab, 0x84, 0xab, 0x25, 0xa4,
	0xb4, 0x2a, 0xe7, 0x12, 0x36, 0x90, 0x90, 0x08, 0x95, 0x4a, 0x87, 0x4a, 0x28, 0x30, 0xb1, 0xa0,
	0x8b, 0xfd, 0xab, 0x7d, 0x10, 0xfb, 0x8c, 0xfd, 0x0b, 0x7f, 0xd6, 0xae, 0x7d, 0x01, 0xe6, 0x4c,
	0xbc, 0x40, 0x27, 0xbf, 0x00, 0x23, 0x8f, 0x80, 0xc2, 0xc0, 0xca, 0x23, 0xa0, 0x3b, 0xdb, 0x22,
	0x03, 0x44, 0x74, 0xb0, 0x65, 0xc9, 0x9f, 0xef, 0xef, 0x8f, 0x3f, 0x3e, 0x3a, 0x8c, 0x15, 0x26,
	0xb3, 0x89, 0x08, 0x75, 0x1a, 0x94, 0x58, 0x48, 0x9c, 0xa5, 0x59, 0x20, 0xa7, 0x2a, 0x84, 0x40,
	0xe6, 0x79, 0x10, 0x26, 0x12, 0x83, 0xb8, 0xc8, 0x43, 0xfb, 0x24, 0xf2, 0x42, 0xa3, 0x76, 0x37,
	0x5b, 0x50, 0x58, 0x50, 0xc8, 0x3c, 0x17, 0xf6, 0xb5, 0x01, 0xbd, 0x3b, 0xb1, 0xd6, 0xf1, 0x14,
	0x02, 0xcb, 0x4e, 0x66, 0xc7, 0x01, 0xaa, 0x14, 0x4a, 0x94, 0x69, 0x5e, 0xc7, 0xbd, 0xbd, 0x7f,
	0xb7, 0x0c, 0xa7, 0xaa, 0xee, 0x06, 0x1f, 0xd1, 0x5c, 0x75, 0xc2, 0x7f, 0x43, 0x37, 0x9e, 0x26,
	0x12, 0x8f, 0xa0, 0x2c, 0x65, 0x0c, 0xee, 0x2e, 0x75, 0x72, 0x80, 0xe2, 0xb5, 0x8a, 0x18, 0xe1,
	0x64, 0xd0, 0x1b, 0x5d, 0x9b, 0x57, 0xcc, 0x79, 0x0e, 0x50, 0xf0, 0xc3, 0xfd, 0x2f, 0x15, 0x23,
	0xbf, 0x2a, 0x46, 0xc6, 0xeb, 0x86, 0x39, 0x8c, 0xdc, 0x7b, 0xd4, 0x49, 0xeb, 0x20, 0xeb, 0x70,
	0x32, 0xb8, 0x64, 0xe9, 0x2b, 0x4d, 0x2d, 0x1e, 0xea, 0x0c, 0x21, 0xc3, 0x71, 0xcb, 0xf8, 0x17,
	0xe9, 0xda, 0x93, 0xf0, 0xad, 0xff, 0x90, 0xd2, 0x67, 0xaa, 0x44, 0x5d, 0x7c, 0x1a, 0xc3, 0xbb,
	0xf3, 0x75, 0xf4, 0x7f, 0x12, 0xda, 0xdb, 0x97, 0x08, 0x51, 0x3b, 0xf0, 0x0e, 0xed, 0x1e, 0x17,
	0x3a, 0x6d, 0xb2, 0x37, 0xe6, 0x15, 0xbb, 0xdc, 0xf6, 0x2f, 0x21, 0x8b, 0xa0, 0xb0, 0x71, 0xcb,
	0xb8, 0xdb, 0xb4, 0x83, 0xda, 0x4e, 0xda, 0x1b, 0xdd, 0x9a, 0x57, 0xec, 0x6a, 0x4b, 0x16, 0x10,
	0xaa, 0x5c, 0x41, 0x86, 0x16, 0xee, 0xa0, 0x76, 0xef, 0x53, 0xa7, 0x19, 0x9f, 0xad, 0xd9, 0xcd,
	0x6e, 0xfe, 0x65, 0x33, 0x33, 0xdd, 0xb8, 0xe5, 0xdc, 0x23, 0xda, 0x35, 0x3a, 0x58, 0x97, 0x93,
	0xc1, 0xc6, 0xd0, 0x13, 0xb5, 0x2b, 0xd1, 0xba, 0x12, 0x2f, 0x5b, 0x57, 0xa3, 0xdb, 0xf3, 0x8a,
	0x5d, 0x5f, 0xea, 0x0d, 0xea, 0x3d, 0x44, 0xdc, 0x84, 0x6d, 0x45, 0x5b, 0x66, 0x78, 0xda, 0xa1,
	0x5d, 0x63, 0xc6, 0x9d, 0x51, 0xa7, 0x5d, 0x76, 0x5b, 0xac, 0xfa, 0x3d, 0xc4, 0x92, 0x48, 0x6f,
	0x6b, 0x35, 0x6a, 0x3c, 0x6c, 0x9e, 0x9c, 0x31, 0xf6, 0x02, 0xb2, 0x88, 0x4b, 0xde, 0x28, 0xe2,
	0xa8, 0xb9, 0xe4, 0xe6, 0x5b, 0xbb, 0xa7, 0x84, 0xd2, 0x03, 0xc0, 0xc6, 0x94, 0x3b, 0x58, 0x5d,
	0xef, 0x8f, 0x50, 0x6f, 0x67, 0x35, 0xb9, 0x6c, 0xcf, 0xbf, 0x7b, 0x72, 0xc6, 0xb6, 0x0e, 0x00,
	0x4b, 0x8e, 0x09, 0x70, 0x03, 0xf1, 0xa4, 0xae, 0xc4, 0x3f, 0x28, 0x4c, 0x9a, 0x59, 0xf6, 0xc8,
	0xe8, 0xf1, 0xd7, 0x45, 0x9f, 0x7c, 0x5b, 0xf4, 0xc9, 0xf7, 0x45, 0x9f, 0x7c, 0xfe, 0xd1, 0xbf,
	0xf0, 0x6a, 0xf7, 0x3f, 0x4f, 0xd7, 0x23, 0x73, 0x9b, 0xac, 0x5b, 0x0d, 0x0f, 0x7e, 0x07, 0x00,
	0x00, 0xff, 0xff, 0x9b, 0x36, 0x33, 0x02, 0x94, 0x03, 0x00, 0x00,
}
