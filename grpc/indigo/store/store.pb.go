// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/alice/grpc/indigo/store/store.proto

/*
	Package store is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/alice/grpc/indigo/store/store.proto

	It has these top-level messages:
		InfoReq
		InfoResp
		SegmentFilter
		MapIDs
		MapFilter
		AddEvidenceReq
		AddEvidenceResp
*/
package store

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/grpc/ext"
import stratumn_alice_pb_indigo_store "github.com/stratumn/alice/pb/indigo/store"

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

type InfoReq struct {
}

func (m *InfoReq) Reset()                    { *m = InfoReq{} }
func (m *InfoReq) String() string            { return proto.CompactTextString(m) }
func (*InfoReq) ProtoMessage()               {}
func (*InfoReq) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{0} }

type InfoResp struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *InfoResp) Reset()                    { *m = InfoResp{} }
func (m *InfoResp) String() string            { return proto.CompactTextString(m) }
func (*InfoResp) ProtoMessage()               {}
func (*InfoResp) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{1} }

func (m *InfoResp) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type SegmentFilter struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *SegmentFilter) Reset()                    { *m = SegmentFilter{} }
func (m *SegmentFilter) String() string            { return proto.CompactTextString(m) }
func (*SegmentFilter) ProtoMessage()               {}
func (*SegmentFilter) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{2} }

func (m *SegmentFilter) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type MapIDs struct {
	MapIds []string `protobuf:"bytes,1,rep,name=map_ids,json=mapIds" json:"map_ids,omitempty"`
}

func (m *MapIDs) Reset()                    { *m = MapIDs{} }
func (m *MapIDs) String() string            { return proto.CompactTextString(m) }
func (*MapIDs) ProtoMessage()               {}
func (*MapIDs) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{3} }

func (m *MapIDs) GetMapIds() []string {
	if m != nil {
		return m.MapIds
	}
	return nil
}

type MapFilter struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *MapFilter) Reset()                    { *m = MapFilter{} }
func (m *MapFilter) String() string            { return proto.CompactTextString(m) }
func (*MapFilter) ProtoMessage()               {}
func (*MapFilter) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{4} }

func (m *MapFilter) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type AddEvidenceReq struct {
	LinkHash *stratumn_alice_pb_indigo_store.LinkHash `protobuf:"bytes,1,opt,name=link_hash,json=linkHash" json:"link_hash,omitempty"`
	Evidence *stratumn_alice_pb_indigo_store.Evidence `protobuf:"bytes,2,opt,name=evidence" json:"evidence,omitempty"`
}

func (m *AddEvidenceReq) Reset()                    { *m = AddEvidenceReq{} }
func (m *AddEvidenceReq) String() string            { return proto.CompactTextString(m) }
func (*AddEvidenceReq) ProtoMessage()               {}
func (*AddEvidenceReq) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{5} }

func (m *AddEvidenceReq) GetLinkHash() *stratumn_alice_pb_indigo_store.LinkHash {
	if m != nil {
		return m.LinkHash
	}
	return nil
}

func (m *AddEvidenceReq) GetEvidence() *stratumn_alice_pb_indigo_store.Evidence {
	if m != nil {
		return m.Evidence
	}
	return nil
}

type AddEvidenceResp struct {
}

func (m *AddEvidenceResp) Reset()                    { *m = AddEvidenceResp{} }
func (m *AddEvidenceResp) String() string            { return proto.CompactTextString(m) }
func (*AddEvidenceResp) ProtoMessage()               {}
func (*AddEvidenceResp) Descriptor() ([]byte, []int) { return fileDescriptorStore, []int{6} }

func init() {
	proto.RegisterType((*InfoReq)(nil), "stratumn.alice.grpc.indigo.store.InfoReq")
	proto.RegisterType((*InfoResp)(nil), "stratumn.alice.grpc.indigo.store.InfoResp")
	proto.RegisterType((*SegmentFilter)(nil), "stratumn.alice.grpc.indigo.store.SegmentFilter")
	proto.RegisterType((*MapIDs)(nil), "stratumn.alice.grpc.indigo.store.MapIDs")
	proto.RegisterType((*MapFilter)(nil), "stratumn.alice.grpc.indigo.store.MapFilter")
	proto.RegisterType((*AddEvidenceReq)(nil), "stratumn.alice.grpc.indigo.store.AddEvidenceReq")
	proto.RegisterType((*AddEvidenceResp)(nil), "stratumn.alice.grpc.indigo.store.AddEvidenceResp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for IndigoStore service

type IndigoStoreClient interface {
	// Get store information.
	GetInfo(ctx context.Context, in *InfoReq, opts ...grpc.CallOption) (*InfoResp, error)
	// Create a link.
	CreateLink(ctx context.Context, in *stratumn_alice_pb_indigo_store.Link, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.LinkHash, error)
	// Get a segment.
	GetSegment(ctx context.Context, in *stratumn_alice_pb_indigo_store.LinkHash, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.Segment, error)
	// Find segments.
	FindSegments(ctx context.Context, in *SegmentFilter, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.Segments, error)
	// Get map IDs.
	GetMapIDs(ctx context.Context, in *MapFilter, opts ...grpc.CallOption) (*MapIDs, error)
	// Add evidence to a segment.
	AddEvidence(ctx context.Context, in *AddEvidenceReq, opts ...grpc.CallOption) (*AddEvidenceResp, error)
	// Get segment evidences.
	GetEvidences(ctx context.Context, in *stratumn_alice_pb_indigo_store.LinkHash, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.Evidences, error)
}

type indigoStoreClient struct {
	cc *grpc.ClientConn
}

func NewIndigoStoreClient(cc *grpc.ClientConn) IndigoStoreClient {
	return &indigoStoreClient{cc}
}

func (c *indigoStoreClient) GetInfo(ctx context.Context, in *InfoReq, opts ...grpc.CallOption) (*InfoResp, error) {
	out := new(InfoResp)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/GetInfo", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indigoStoreClient) CreateLink(ctx context.Context, in *stratumn_alice_pb_indigo_store.Link, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.LinkHash, error) {
	out := new(stratumn_alice_pb_indigo_store.LinkHash)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/CreateLink", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indigoStoreClient) GetSegment(ctx context.Context, in *stratumn_alice_pb_indigo_store.LinkHash, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.Segment, error) {
	out := new(stratumn_alice_pb_indigo_store.Segment)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/GetSegment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indigoStoreClient) FindSegments(ctx context.Context, in *SegmentFilter, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.Segments, error) {
	out := new(stratumn_alice_pb_indigo_store.Segments)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/FindSegments", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indigoStoreClient) GetMapIDs(ctx context.Context, in *MapFilter, opts ...grpc.CallOption) (*MapIDs, error) {
	out := new(MapIDs)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/GetMapIDs", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indigoStoreClient) AddEvidence(ctx context.Context, in *AddEvidenceReq, opts ...grpc.CallOption) (*AddEvidenceResp, error) {
	out := new(AddEvidenceResp)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/AddEvidence", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indigoStoreClient) GetEvidences(ctx context.Context, in *stratumn_alice_pb_indigo_store.LinkHash, opts ...grpc.CallOption) (*stratumn_alice_pb_indigo_store.Evidences, error) {
	out := new(stratumn_alice_pb_indigo_store.Evidences)
	err := grpc.Invoke(ctx, "/stratumn.alice.grpc.indigo.store.IndigoStore/GetEvidences", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for IndigoStore service

type IndigoStoreServer interface {
	// Get store information.
	GetInfo(context.Context, *InfoReq) (*InfoResp, error)
	// Create a link.
	CreateLink(context.Context, *stratumn_alice_pb_indigo_store.Link) (*stratumn_alice_pb_indigo_store.LinkHash, error)
	// Get a segment.
	GetSegment(context.Context, *stratumn_alice_pb_indigo_store.LinkHash) (*stratumn_alice_pb_indigo_store.Segment, error)
	// Find segments.
	FindSegments(context.Context, *SegmentFilter) (*stratumn_alice_pb_indigo_store.Segments, error)
	// Get map IDs.
	GetMapIDs(context.Context, *MapFilter) (*MapIDs, error)
	// Add evidence to a segment.
	AddEvidence(context.Context, *AddEvidenceReq) (*AddEvidenceResp, error)
	// Get segment evidences.
	GetEvidences(context.Context, *stratumn_alice_pb_indigo_store.LinkHash) (*stratumn_alice_pb_indigo_store.Evidences, error)
}

func RegisterIndigoStoreServer(s *grpc.Server, srv IndigoStoreServer) {
	s.RegisterService(&_IndigoStore_serviceDesc, srv)
}

func _IndigoStore_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).GetInfo(ctx, req.(*InfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndigoStore_CreateLink_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_pb_indigo_store.Link)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).CreateLink(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/CreateLink",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).CreateLink(ctx, req.(*stratumn_alice_pb_indigo_store.Link))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndigoStore_GetSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_pb_indigo_store.LinkHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).GetSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/GetSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).GetSegment(ctx, req.(*stratumn_alice_pb_indigo_store.LinkHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndigoStore_FindSegments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SegmentFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).FindSegments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/FindSegments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).FindSegments(ctx, req.(*SegmentFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndigoStore_GetMapIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MapFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).GetMapIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/GetMapIDs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).GetMapIDs(ctx, req.(*MapFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndigoStore_AddEvidence_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddEvidenceReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).AddEvidence(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/AddEvidence",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).AddEvidence(ctx, req.(*AddEvidenceReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _IndigoStore_GetEvidences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(stratumn_alice_pb_indigo_store.LinkHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndigoStoreServer).GetEvidences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/stratumn.alice.grpc.indigo.store.IndigoStore/GetEvidences",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndigoStoreServer).GetEvidences(ctx, req.(*stratumn_alice_pb_indigo_store.LinkHash))
	}
	return interceptor(ctx, in, info, handler)
}

var _IndigoStore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "stratumn.alice.grpc.indigo.store.IndigoStore",
	HandlerType: (*IndigoStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetInfo",
			Handler:    _IndigoStore_GetInfo_Handler,
		},
		{
			MethodName: "CreateLink",
			Handler:    _IndigoStore_CreateLink_Handler,
		},
		{
			MethodName: "GetSegment",
			Handler:    _IndigoStore_GetSegment_Handler,
		},
		{
			MethodName: "FindSegments",
			Handler:    _IndigoStore_FindSegments_Handler,
		},
		{
			MethodName: "GetMapIDs",
			Handler:    _IndigoStore_GetMapIDs_Handler,
		},
		{
			MethodName: "AddEvidence",
			Handler:    _IndigoStore_AddEvidence_Handler,
		},
		{
			MethodName: "GetEvidences",
			Handler:    _IndigoStore_GetEvidences_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/stratumn/alice/grpc/indigo/store/store.proto",
}

func (m *InfoReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InfoReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *InfoResp) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InfoResp) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintStore(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *SegmentFilter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SegmentFilter) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintStore(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *MapIDs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MapIDs) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.MapIds) > 0 {
		for _, s := range m.MapIds {
			dAtA[i] = 0xa
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	return i, nil
}

func (m *MapFilter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MapFilter) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintStore(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *AddEvidenceReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AddEvidenceReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.LinkHash != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintStore(dAtA, i, uint64(m.LinkHash.Size()))
		n1, err := m.LinkHash.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.Evidence != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintStore(dAtA, i, uint64(m.Evidence.Size()))
		n2, err := m.Evidence.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func (m *AddEvidenceResp) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AddEvidenceResp) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func encodeVarintStore(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *InfoReq) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *InfoResp) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovStore(uint64(l))
	}
	return n
}

func (m *SegmentFilter) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovStore(uint64(l))
	}
	return n
}

func (m *MapIDs) Size() (n int) {
	var l int
	_ = l
	if len(m.MapIds) > 0 {
		for _, s := range m.MapIds {
			l = len(s)
			n += 1 + l + sovStore(uint64(l))
		}
	}
	return n
}

func (m *MapFilter) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovStore(uint64(l))
	}
	return n
}

func (m *AddEvidenceReq) Size() (n int) {
	var l int
	_ = l
	if m.LinkHash != nil {
		l = m.LinkHash.Size()
		n += 1 + l + sovStore(uint64(l))
	}
	if m.Evidence != nil {
		l = m.Evidence.Size()
		n += 1 + l + sovStore(uint64(l))
	}
	return n
}

func (m *AddEvidenceResp) Size() (n int) {
	var l int
	_ = l
	return n
}

func sovStore(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozStore(x uint64) (n int) {
	return sovStore(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *InfoReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: InfoReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InfoReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func (m *InfoResp) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: InfoResp: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InfoResp: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStore
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
				return ErrInvalidLengthStore
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func (m *SegmentFilter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: SegmentFilter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SegmentFilter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStore
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
				return ErrInvalidLengthStore
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func (m *MapIDs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: MapIDs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MapIDs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MapIds", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStore
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
				return ErrInvalidLengthStore
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MapIds = append(m.MapIds, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func (m *MapFilter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: MapFilter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MapFilter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStore
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
				return ErrInvalidLengthStore
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func (m *AddEvidenceReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: AddEvidenceReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AddEvidenceReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LinkHash", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStore
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
				return ErrInvalidLengthStore
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.LinkHash == nil {
				m.LinkHash = &stratumn_alice_pb_indigo_store.LinkHash{}
			}
			if err := m.LinkHash.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Evidence", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStore
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
				return ErrInvalidLengthStore
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Evidence == nil {
				m.Evidence = &stratumn_alice_pb_indigo_store.Evidence{}
			}
			if err := m.Evidence.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func (m *AddEvidenceResp) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStore
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
			return fmt.Errorf("proto: AddEvidenceResp: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AddEvidenceResp: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipStore(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStore
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
func skipStore(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowStore
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
					return 0, ErrIntOverflowStore
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
					return 0, ErrIntOverflowStore
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
				return 0, ErrInvalidLengthStore
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowStore
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
				next, err := skipStore(dAtA[start:])
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
	ErrInvalidLengthStore = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowStore   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/alice/grpc/indigo/store/store.proto", fileDescriptorStore)
}

var fileDescriptorStore = []byte{
	// 697 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x95, 0xcf, 0x4f, 0x13, 0x4f,
	0x18, 0xc6, 0xbf, 0xfb, 0xd5, 0x50, 0x3a, 0x45, 0x0d, 0x13, 0x21, 0x43, 0x0d, 0x75, 0xd8, 0x60,
	0x04, 0x94, 0xad, 0x94, 0xa8, 0x31, 0x9e, 0x28, 0x48, 0x2d, 0x11, 0x4d, 0xca, 0xcd, 0xc4, 0x90,
	0x69, 0xf7, 0xa5, 0x9d, 0xd0, 0x9d, 0x1d, 0x76, 0x06, 0xc4, 0xc4, 0x78, 0x68, 0xe2, 0xd9, 0x98,
	0x78, 0xf0, 0xdc, 0xff, 0x61, 0x4f, 0x3d, 0x9b, 0x78, 0xf4, 0xe4, 0xd9, 0xe0, 0x3f, 0x62, 0xf6,
	0x17, 0x76, 0x81, 0xb8, 0xe5, 0xd0, 0x4d, 0x0f, 0xcf, 0x7c, 0x9e, 0x77, 0xde, 0xf7, 0x79, 0x33,
	0xe8, 0x71, 0x9b, 0xeb, 0xce, 0x61, 0xd3, 0x6a, 0xb9, 0x4e, 0x59, 0x69, 0x8f, 0xe9, 0x43, 0x47,
	0x94, 0x59, 0x97, 0xb7, 0xa0, 0xdc, 0xf6, 0x64, 0xab, 0xcc, 0x85, 0xcd, 0xdb, 0x6e, 0x59, 0x69,
	0xd7, 0x83, 0xe8, 0x6b, 0x49, 0xcf, 0xd5, 0x2e, 0xa6, 0x89, 0xda, 0x0a, 0xd5, 0x56, 0xa0, 0xb6,
	0x22, 0xb5, 0x15, 0xea, 0x8a, 0xf7, 0x33, 0xd0, 0x70, 0xac, 0x83, 0x5f, 0xc4, 0x2b, 0xfe, 0xa3,
	0x10, 0xd9, 0x3c, 0x53, 0x06, 0xb4, 0x1d, 0x10, 0xf1, 0x41, 0x33, 0x8f, 0x72, 0x75, 0xb1, 0xe7,
	0x36, 0xe0, 0xc0, 0xac, 0xa2, 0xf1, 0xe8, 0xaf, 0x92, 0xf8, 0x11, 0xba, 0x6a, 0x33, 0xcd, 0x88,
	0x41, 0x8d, 0x85, 0x89, 0xaa, 0xd9, 0x1f, 0x90, 0xd2, 0xd6, 0xce, 0xab, 0x97, 0xcb, 0x20, 0x5a,
	0xae, 0x0d, 0x36, 0x0d, 0x51, 0x94, 0x8b, 0x3d, 0xd7, 0x73, 0x98, 0xe6, 0xae, 0x68, 0x84, 0x7a,
	0x73, 0x0b, 0x5d, 0xdb, 0x89, 0xf8, 0x9b, 0xbc, 0xab, 0xc1, 0xc3, 0x4f, 0x52, 0xa0, 0x3b, 0xfd,
	0x01, 0x99, 0x4b, 0x83, 0x22, 0x35, 0xdd, 0x0b, 0xe5, 0xb4, 0xf9, 0x4e, 0x83, 0x8a, 0x59, 0x16,
	0x1a, 0xdb, 0x66, 0xb2, 0xbe, 0xa1, 0xf0, 0x3c, 0xca, 0x39, 0x4c, 0xee, 0x72, 0x5b, 0x11, 0x83,
	0x5e, 0x59, 0xc8, 0x57, 0x0b, 0xfd, 0x01, 0xc9, 0x6d, 0x33, 0x49, 0xeb, 0x1b, 0xaa, 0x31, 0xe6,
	0x30, 0x59, 0xb7, 0x95, 0x59, 0x45, 0xf9, 0x6d, 0x26, 0x63, 0xdf, 0x87, 0x29, 0xdf, 0xb9, 0xfe,
	0x80, 0xcc, 0xa6, 0x7c, 0x1d, 0x26, 0x2f, 0xf2, 0xfc, 0x69, 0xa0, 0xeb, 0x6b, 0xb6, 0xfd, 0xec,
	0x88, 0xdb, 0x20, 0x5a, 0xd0, 0x80, 0x03, 0xfc, 0x06, 0xe5, 0xbb, 0x5c, 0xec, 0xef, 0x76, 0x98,
	0xea, 0x84, 0xb8, 0x42, 0x65, 0xc1, 0x3a, 0x33, 0x3e, 0xd9, 0x4c, 0x0d, 0xcf, 0x7a, 0xc1, 0xc5,
	0xfe, 0x73, 0xa6, 0x3a, 0xd5, 0xe9, 0xfe, 0x80, 0xe0, 0xd5, 0xca, 0x72, 0xe8, 0x41, 0x03, 0x0e,
	0x0d, 0x38, 0x8d, 0xf1, 0x6e, 0xac, 0xc0, 0x80, 0xc6, 0x21, 0x76, 0x23, 0xff, 0x8f, 0x46, 0x4f,
	0xaa, 0xab, 0xde, 0xee, 0x0f, 0xc8, 0xad, 0xd4, 0xb5, 0x12, 0x54, 0x7c, 0xa9, 0x53, 0xb4, 0x39,
	0x89, 0x6e, 0xa4, 0xee, 0xa5, 0x64, 0xe5, 0x5b, 0x0e, 0x15, 0xea, 0x21, 0x77, 0x27, 0xc0, 0xe2,
	0x2f, 0x06, 0xca, 0xd5, 0x40, 0x07, 0x19, 0xc0, 0x8b, 0x56, 0x56, 0x40, 0xad, 0x38, 0x36, 0xc5,
	0xa5, 0x51, 0xa5, 0x4a, 0x9a, 0x95, 0x9e, 0x4f, 0xac, 0x1a, 0xe8, 0xe1, 0xe4, 0x50, 0xd6, 0x74,
	0x0f, 0x35, 0xd5, 0x1d, 0xa0, 0x51, 0x35, 0x34, 0x2c, 0x87, 0x2a, 0xf0, 0x8e, 0x78, 0x0b, 0xb0,
	0x42, 0x68, 0xdd, 0x03, 0xa6, 0x21, 0x68, 0x2a, 0x9e, 0x1f, 0xa5, 0xf5, 0xc5, 0x91, 0x07, 0x64,
	0x4e, 0xf5, 0x7c, 0x32, 0x19, 0xf1, 0x29, 0xa3, 0x02, 0xde, 0x86, 0x43, 0xc2, 0x9f, 0x0c, 0x84,
	0x6a, 0xa0, 0xe3, 0x2c, 0xe3, 0x91, 0x79, 0xc5, 0xbb, 0x59, 0xca, 0x18, 0x69, 0x5a, 0x3d, 0x9f,
	0x2c, 0x05, 0xad, 0x60, 0x82, 0xc2, 0x31, 0x57, 0x9a, 0x8b, 0xf6, 0xdf, 0x6d, 0xf0, 0x5c, 0x87,
	0x72, 0x3d, 0x94, 0x18, 0xfc, 0xd1, 0x40, 0x13, 0x9b, 0x5c, 0xd8, 0xf1, 0x79, 0x85, 0xcb, 0xd9,
	0x7d, 0x4f, 0xad, 0x62, 0x76, 0x53, 0x12, 0xb4, 0x59, 0xec, 0xf9, 0x64, 0x3a, 0x30, 0x3b, 0x57,
	0x99, 0xc2, 0xef, 0x51, 0xbe, 0x06, 0x3a, 0x5e, 0xcc, 0x7b, 0xd9, 0x35, 0x9c, 0xae, 0xe4, 0x79,
	0xff, 0x0b, 0xc5, 0xf5, 0x0d, 0x65, 0x92, 0x9e, 0x4f, 0x6e, 0x06, 0xbd, 0x39, 0xb5, 0x77, 0xa2,
	0x5d, 0xc7, 0x9f, 0x0d, 0x54, 0x18, 0xca, 0x31, 0x7e, 0x90, 0xcd, 0x4c, 0xaf, 0x73, 0x71, 0xe5,
	0x92, 0x27, 0x94, 0x34, 0x67, 0x7b, 0x3e, 0x99, 0x59, 0xb3, 0x87, 0xb6, 0x4b, 0xbb, 0x94, 0x25,
	0x2d, 0xc1, 0x1f, 0xd0, 0x44, 0x0d, 0x74, 0x72, 0x42, 0x5d, 0x22, 0x2c, 0x8b, 0xa3, 0x6e, 0xba,
	0x32, 0x67, 0x7a, 0x3e, 0x99, 0x0a, 0x5a, 0x92, 0x44, 0x24, 0xa9, 0x45, 0x55, 0xd7, 0xbf, 0x9f,
	0x94, 0x8c, 0x1f, 0x27, 0x25, 0xe3, 0xd7, 0x49, 0xc9, 0xf8, 0xfa, 0xbb, 0xf4, 0xdf, 0xeb, 0x95,
	0x4b, 0x3c, 0x4b, 0x4f, 0xc3, 0x6f, 0x73, 0x2c, 0x7c, 0x0e, 0x56, 0xff, 0x04, 0x00, 0x00, 0xff,
	0xff, 0xef, 0xbb, 0xbb, 0x4f, 0xd2, 0x06, 0x00, 0x00,
}
