// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-node/app/contacts/grpc/contacts.proto

package grpc // import "github.com/stratumn/go-node/app/contacts/grpc"

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

// The contact message containing the name and peer ID.
type Contact struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	PeerId               []byte   `protobuf:"bytes,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Contact) Reset()         { *m = Contact{} }
func (m *Contact) String() string { return proto.CompactTextString(m) }
func (*Contact) ProtoMessage()    {}
func (*Contact) Descriptor() ([]byte, []int) {
	return fileDescriptor_contacts_1513a9f9ad33889f, []int{0}
}
func (m *Contact) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Contact) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Contact.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *Contact) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Contact.Merge(dst, src)
}
func (m *Contact) XXX_Size() int {
	return m.Size()
}
func (m *Contact) XXX_DiscardUnknown() {
	xxx_messageInfo_Contact.DiscardUnknown(m)
}

var xxx_messageInfo_Contact proto.InternalMessageInfo

func (m *Contact) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Contact) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

// The list request message.
type ListReq struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListReq) Reset()         { *m = ListReq{} }
func (m *ListReq) String() string { return proto.CompactTextString(m) }
func (*ListReq) ProtoMessage()    {}
func (*ListReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_contacts_1513a9f9ad33889f, []int{1}
}
func (m *ListReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ListReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ListReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *ListReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListReq.Merge(dst, src)
}
func (m *ListReq) XXX_Size() int {
	return m.Size()
}
func (m *ListReq) XXX_DiscardUnknown() {
	xxx_messageInfo_ListReq.DiscardUnknown(m)
}

var xxx_messageInfo_ListReq proto.InternalMessageInfo

// The get request message.
type GetReq struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetReq) Reset()         { *m = GetReq{} }
func (m *GetReq) String() string { return proto.CompactTextString(m) }
func (*GetReq) ProtoMessage()    {}
func (*GetReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_contacts_1513a9f9ad33889f, []int{2}
}
func (m *GetReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GetReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GetReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *GetReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetReq.Merge(dst, src)
}
func (m *GetReq) XXX_Size() int {
	return m.Size()
}
func (m *GetReq) XXX_DiscardUnknown() {
	xxx_messageInfo_GetReq.DiscardUnknown(m)
}

var xxx_messageInfo_GetReq proto.InternalMessageInfo

func (m *GetReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// The set request message.
type SetReq struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	PeerId               []byte   `protobuf:"bytes,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SetReq) Reset()         { *m = SetReq{} }
func (m *SetReq) String() string { return proto.CompactTextString(m) }
func (*SetReq) ProtoMessage()    {}
func (*SetReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_contacts_1513a9f9ad33889f, []int{3}
}
func (m *SetReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SetReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SetReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *SetReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetReq.Merge(dst, src)
}
func (m *SetReq) XXX_Size() int {
	return m.Size()
}
func (m *SetReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SetReq.DiscardUnknown(m)
}

var xxx_messageInfo_SetReq proto.InternalMessageInfo

func (m *SetReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SetReq) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

// The delete request message.
type DeleteReq struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteReq) Reset()         { *m = DeleteReq{} }
func (m *DeleteReq) String() string { return proto.CompactTextString(m) }
func (*DeleteReq) ProtoMessage()    {}
func (*DeleteReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_contacts_1513a9f9ad33889f, []int{4}
}
func (m *DeleteReq) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DeleteReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DeleteReq.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (dst *DeleteReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteReq.Merge(dst, src)
}
func (m *DeleteReq) XXX_Size() int {
	return m.Size()
}
func (m *DeleteReq) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteReq.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteReq proto.InternalMessageInfo

func (m *DeleteReq) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*Contact)(nil), "stratumn.node.app.contacts.grpc.Contact")
	proto.RegisterType((*ListReq)(nil), "stratumn.node.app.contacts.grpc.ListReq")
	proto.RegisterType((*GetReq)(nil), "stratumn.node.app.contacts.grpc.GetReq")
	proto.RegisterType((*SetReq)(nil), "stratumn.node.app.contacts.grpc.SetReq")
	proto.RegisterType((*DeleteReq)(nil), "stratumn.node.app.contacts.grpc.DeleteReq")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Contacts service

type ContactsClient interface {
	// Streams all the contacts.
	List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (Contacts_ListClient, error)
	// Returns a contact.
	Get(ctx context.Context, in *GetReq, opts ...grpc.CallOption) (*Contact, error)
	// Sets a contact.
	Set(ctx context.Context, in *SetReq, opts ...grpc.CallOption) (*Contact, error)
	// Delete a contact.
	Delete(ctx context.Context, in *DeleteReq, opts ...grpc.CallOption) (*Contact, error)
}

type contactsClient struct {
	cc *grpc.ClientConn
}

func NewContactsClient(cc *grpc.ClientConn) ContactsClient {
	return &contactsClient{cc}
}

func (c *contactsClient) List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (Contacts_ListClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Contacts_serviceDesc.Streams[0], "/stratumn.node.app.contacts.grpc.Contacts/List", opts...)
	if err != nil {
		return nil, err
	}
	x := &contactsListClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Contacts_ListClient interface {
	Recv() (*Contact, error)
	grpc.ClientStream
}

type contactsListClient struct {
	grpc.ClientStream
}

func (x *contactsListClient) Recv() (*Contact, error) {
	m := new(Contact)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *contactsClient) Get(ctx context.Context, in *GetReq, opts ...grpc.CallOption) (*Contact, error) {
	out := new(Contact)
	err := c.cc.Invoke(ctx, "/stratumn.node.app.contacts.grpc.Contacts/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contactsClient) Set(ctx context.Context, in *SetReq, opts ...grpc.CallOption) (*Contact, error) {
	out := new(Contact)
	err := c.cc.Invoke(ctx, "/stratumn.node.app.contacts.grpc.Contacts/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contactsClient) Delete(ctx context.Context, in *DeleteReq, opts ...grpc.CallOption) (*Contact, error) {
	out := new(Contact)
	err := c.cc.Invoke(ctx, "/stratumn.node.app.contacts.grpc.Contacts/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Contacts service

type ContactsServer interface {
	// Streams all the contacts.
	List(*ListReq, Contacts_ListServer) error
	// Returns a contact.
	Get(context.Context, *GetReq) (*Contact, error)
	// Sets a contact.
	Set(context.Context, *SetReq) (*Contact, error)
	// Delete a contact.
	Delete(context.Context, *DeleteReq) (*Contact, error)
}

func RegisterContactsServer(s *grpc.Server, srv ContactsServer) {
	s.RegisterService(&_Contacts_serviceDesc, srv)
}

func _Contacts_List_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ContactsServer).List(m, &contactsListServer{stream})
}

type Contacts_ListServer interface {
	Send(*Contact) error
	grpc.ServerStream
}

type contactsListServer struct {
	grpc.ServerStream
}

func (x *contactsListServer) Send(m *Contact) error {
	return x.ServerStream.SendMsg(m)
}

func _Contacts_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContactsServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.app.contacts.grpc.Contacts/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContactsServer).Get(ctx, req.(*GetReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Contacts_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContactsServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.app.contacts.grpc.Contacts/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContactsServer).Set(ctx, req.(*SetReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Contacts_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContactsServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.node.app.contacts.grpc.Contacts/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContactsServer).Delete(ctx, req.(*DeleteReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Contacts_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.node.app.contacts.grpc.Contacts",
	HandlerType: (*ContactsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Contacts_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _Contacts_Set_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Contacts_Delete_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "List",
			Handler:       _Contacts_List_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/stratumn/go-node/app/contacts/grpc/contacts.proto",
}

func (m *Contact) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Contact) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintContacts(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.PeerId) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintContacts(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *ListReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ListReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *GetReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintContacts(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *SetReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintContacts(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if len(m.PeerId) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintContacts(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *DeleteReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DeleteReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Name) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintContacts(dAtA, i, uint64(len(m.Name)))
		i += copy(dAtA[i:], m.Name)
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintContacts(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Contact) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovContacts(uint64(l))
	}
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovContacts(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *ListReq) Size() (n int) {
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *GetReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovContacts(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *SetReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovContacts(uint64(l))
	}
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovContacts(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *DeleteReq) Size() (n int) {
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovContacts(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovContacts(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozContacts(x uint64) (n int) {
	return sovContacts(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Contact) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContacts
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
			return fmt.Errorf("proto: Contact: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Contact: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContacts
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
				return ErrInvalidLengthContacts
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContacts
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
				return ErrInvalidLengthContacts
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
			skippy, err := skipContacts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthContacts
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
func (m *ListReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContacts
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
			return fmt.Errorf("proto: ListReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ListReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipContacts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthContacts
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
func (m *GetReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContacts
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
			return fmt.Errorf("proto: GetReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContacts
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
				return ErrInvalidLengthContacts
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipContacts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthContacts
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
func (m *SetReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContacts
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
			return fmt.Errorf("proto: SetReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContacts
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
				return ErrInvalidLengthContacts
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContacts
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
				return ErrInvalidLengthContacts
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
			skippy, err := skipContacts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthContacts
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
func (m *DeleteReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowContacts
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
			return fmt.Errorf("proto: DeleteReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DeleteReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowContacts
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
				return ErrInvalidLengthContacts
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipContacts(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthContacts
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
func skipContacts(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowContacts
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
					return 0, ErrIntOverflowContacts
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
					return 0, ErrIntOverflowContacts
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
				return 0, ErrInvalidLengthContacts
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowContacts
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
				next, err := skipContacts(dAtA[start:])
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
	ErrInvalidLengthContacts = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowContacts   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-node/app/contacts/grpc/contacts.proto", fileDescriptor_contacts_1513a9f9ad33889f)
}

var fileDescriptor_contacts_1513a9f9ad33889f = []byte{
	// 403 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xb2, 0x49, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x2f, 0x2e, 0x29, 0x4a, 0x2c, 0x29, 0xcd, 0xcd, 0xd3,
	0x4f, 0xcf, 0xd7, 0xcd, 0xcb, 0x4f, 0x49, 0xd5, 0x4f, 0x2c, 0x28, 0xd0, 0x4f, 0xce, 0xcf, 0x2b,
	0x49, 0x4c, 0x2e, 0x29, 0xd6, 0x4f, 0x2f, 0x2a, 0x48, 0x86, 0xf3, 0xf4, 0x0a, 0x8a, 0xf2, 0x4b,
	0xf2, 0x85, 0xe4, 0x61, 0x5a, 0xf4, 0x40, 0xea, 0xf5, 0x12, 0x0b, 0x0a, 0xf4, 0xe0, 0x2a, 0x40,
	0xea, 0xa5, 0x8c, 0xf0, 0x19, 0x9f, 0x9c, 0x93, 0x09, 0x31, 0x35, 0xb5, 0xa2, 0x04, 0x84, 0x21,
	0x86, 0x2a, 0xd9, 0x70, 0xb1, 0x3b, 0x43, 0x0c, 0x11, 0x12, 0xe2, 0x62, 0xc9, 0x4b, 0xcc, 0x4d,
	0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0xb3, 0x85, 0x64, 0xb9, 0xd8, 0x0b, 0x52, 0x53,
	0x8b, 0xe2, 0x33, 0x53, 0x24, 0x98, 0x14, 0x18, 0x35, 0x78, 0x9c, 0x58, 0x3e, 0xec, 0x96, 0x60,
	0x0c, 0x62, 0x03, 0x09, 0x7a, 0xa6, 0x28, 0x71, 0x72, 0xb1, 0xfb, 0x64, 0x16, 0x97, 0x04, 0xa5,
	0x16, 0x2a, 0x19, 0x71, 0xb1, 0xb9, 0xa7, 0x82, 0x58, 0x42, 0x1a, 0xc8, 0xe6, 0x38, 0x89, 0x2c,
	0xda, 0x2d, 0xc1, 0x03, 0xb5, 0x42, 0x01, 0x24, 0xbe, 0x02, 0x64, 0x00, 0x58, 0x85, 0x52, 0x02,
	0x17, 0x5b, 0x30, 0x89, 0x7a, 0x84, 0x74, 0xd0, 0x5d, 0x24, 0xbc, 0x68, 0xb7, 0x04, 0x7b, 0x40,
	0x6a, 0x6a, 0x91, 0x82, 0xa7, 0x0b, 0x48, 0x1d, 0x8a, 0x03, 0x4d, 0xb9, 0x38, 0x5d, 0x52, 0x73,
	0x52, 0x4b, 0x52, 0x49, 0xb2, 0xc4, 0xe8, 0x3a, 0x33, 0x17, 0x07, 0x54, 0xaa, 0x58, 0xa8, 0x9c,
	0x8b, 0x05, 0xe4, 0x49, 0x21, 0x0d, 0x3d, 0x02, 0x11, 0xa0, 0x07, 0x0d, 0x0b, 0x29, 0xc2, 0x2a,
	0xa1, 0x86, 0x2b, 0x49, 0x36, 0x6d, 0x95, 0x10, 0x05, 0x69, 0x53, 0x48, 0xcc, 0xc9, 0x51, 0x28,
	0xc9, 0x48, 0x55, 0x80, 0xa9, 0x33, 0x60, 0x14, 0xca, 0xe7, 0x62, 0x76, 0x4f, 0x2d, 0x11, 0x52,
	0x27, 0x68, 0x1a, 0x24, 0xe0, 0x49, 0xb0, 0x56, 0xb4, 0x69, 0xab, 0x84, 0xa0, 0x4b, 0x66, 0x71,
	0x41, 0x4e, 0x62, 0xa5, 0x42, 0x22, 0xcc, 0x4e, 0xa1, 0x6c, 0x2e, 0xe6, 0x60, 0xa2, 0x2c, 0x0c,
	0x26, 0xd5, 0x42, 0xc1, 0xa6, 0xad, 0x12, 0xbc, 0xc1, 0xa9, 0x25, 0x48, 0x96, 0x95, 0x70, 0xb1,
	0x41, 0xa2, 0x46, 0x48, 0x8b, 0xa0, 0x31, 0xf0, 0x38, 0x24, 0xc1, 0x4a, 0x91, 0xa6, 0xad, 0x12,
	0x02, 0x10, 0x8d, 0x08, 0x5b, 0x9d, 0x5c, 0x4e, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1,
	0xc1, 0x23, 0x39, 0xc6, 0x19, 0x8f, 0xe5, 0x18, 0xa2, 0x8c, 0x48, 0xca, 0x94, 0xd6, 0x20, 0x22,
	0x89, 0x0d, 0x9c, 0x79, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x49, 0xc8, 0x0d, 0xde, 0xd1,
	0x03, 0x00, 0x00,
}
