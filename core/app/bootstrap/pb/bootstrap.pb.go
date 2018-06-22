// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/alice/core/app/bootstrap/pb/bootstrap.proto

/*
	Package pb is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/alice/core/app/bootstrap/pb/bootstrap.proto

	It has these top-level messages:
		Hello
		Ack
		CompleteReq
		Filter
		PeerID
		NodeIdentity
		UpdateProposal
		Vote
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/alice/grpc/ext"
import stratumn_alice_core_crypto "github.com/stratumn/alice/core/crypto"

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

// Type of network updates that can be proposed.
type UpdateType int32

const (
	UpdateType_None       UpdateType = 0
	UpdateType_AddNode    UpdateType = 1
	UpdateType_RemoveNode UpdateType = 2
)

var UpdateType_name = map[int32]string{
	0: "None",
	1: "AddNode",
	2: "RemoveNode",
}
var UpdateType_value = map[string]int32{
	"None":       0,
	"AddNode":    1,
	"RemoveNode": 2,
}

func (x UpdateType) String() string {
	return proto.EnumName(UpdateType_name, int32(x))
}
func (UpdateType) EnumDescriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{0} }

// A simple Hello handshake message.
type Hello struct {
}

func (m *Hello) Reset()                    { *m = Hello{} }
func (m *Hello) String() string            { return proto.CompactTextString(m) }
func (*Hello) ProtoMessage()               {}
func (*Hello) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{0} }

// A message ack.
type Ack struct {
	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (m *Ack) Reset()                    { *m = Ack{} }
func (m *Ack) String() string            { return proto.CompactTextString(m) }
func (*Ack) ProtoMessage()               {}
func (*Ack) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{1} }

func (m *Ack) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// A request to complete the bootstrap phase.
type CompleteReq struct {
}

func (m *CompleteReq) Reset()                    { *m = CompleteReq{} }
func (m *CompleteReq) String() string            { return proto.CompactTextString(m) }
func (*CompleteReq) ProtoMessage()               {}
func (*CompleteReq) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{2} }

// A results filter.
type Filter struct {
}

func (m *Filter) Reset()                    { *m = Filter{} }
func (m *Filter) String() string            { return proto.CompactTextString(m) }
func (*Filter) ProtoMessage()               {}
func (*Filter) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{3} }

// A base58-encoded PeerId.
type PeerID struct {
	PeerId []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
}

func (m *PeerID) Reset()                    { *m = PeerID{} }
func (m *PeerID) String() string            { return proto.CompactTextString(m) }
func (*PeerID) ProtoMessage()               {}
func (*PeerID) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{4} }

func (m *PeerID) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

// A message containing a proof of a node's identity.
type NodeIdentity struct {
	PeerId        []byte `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	PeerAddr      []byte `protobuf:"bytes,2,opt,name=peer_addr,json=peerAddr,proto3" json:"peer_addr,omitempty"`
	IdentityProof []byte `protobuf:"bytes,3,opt,name=identity_proof,json=identityProof,proto3" json:"identity_proof,omitempty"`
}

func (m *NodeIdentity) Reset()                    { *m = NodeIdentity{} }
func (m *NodeIdentity) String() string            { return proto.CompactTextString(m) }
func (*NodeIdentity) ProtoMessage()               {}
func (*NodeIdentity) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{5} }

func (m *NodeIdentity) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

func (m *NodeIdentity) GetPeerAddr() []byte {
	if m != nil {
		return m.PeerAddr
	}
	return nil
}

func (m *NodeIdentity) GetIdentityProof() []byte {
	if m != nil {
		return m.IdentityProof
	}
	return nil
}

// Proposal to update the network.
type UpdateProposal struct {
	UpdateType  UpdateType    `protobuf:"varint,1,opt,name=update_type,json=updateType,proto3,enum=stratumn.alice.core.app.bootstrap.pb.UpdateType" json:"update_type,omitempty"`
	NodeDetails *NodeIdentity `protobuf:"bytes,2,opt,name=node_details,json=nodeDetails" json:"node_details,omitempty"`
	Challenge   []byte        `protobuf:"bytes,3,opt,name=challenge,proto3" json:"challenge,omitempty"`
}

func (m *UpdateProposal) Reset()                    { *m = UpdateProposal{} }
func (m *UpdateProposal) String() string            { return proto.CompactTextString(m) }
func (*UpdateProposal) ProtoMessage()               {}
func (*UpdateProposal) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{6} }

func (m *UpdateProposal) GetUpdateType() UpdateType {
	if m != nil {
		return m.UpdateType
	}
	return UpdateType_None
}

func (m *UpdateProposal) GetNodeDetails() *NodeIdentity {
	if m != nil {
		return m.NodeDetails
	}
	return nil
}

func (m *UpdateProposal) GetChallenge() []byte {
	if m != nil {
		return m.Challenge
	}
	return nil
}

type Vote struct {
	UpdateType UpdateType                            `protobuf:"varint,1,opt,name=update_type,json=updateType,proto3,enum=stratumn.alice.core.app.bootstrap.pb.UpdateType" json:"update_type,omitempty"`
	PeerId     []byte                                `protobuf:"bytes,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	Challenge  []byte                                `protobuf:"bytes,3,opt,name=challenge,proto3" json:"challenge,omitempty"`
	Signature  *stratumn_alice_core_crypto.Signature `protobuf:"bytes,4,opt,name=signature" json:"signature,omitempty"`
}

func (m *Vote) Reset()                    { *m = Vote{} }
func (m *Vote) String() string            { return proto.CompactTextString(m) }
func (*Vote) ProtoMessage()               {}
func (*Vote) Descriptor() ([]byte, []int) { return fileDescriptorBootstrap, []int{7} }

func (m *Vote) GetUpdateType() UpdateType {
	if m != nil {
		return m.UpdateType
	}
	return UpdateType_None
}

func (m *Vote) GetPeerId() []byte {
	if m != nil {
		return m.PeerId
	}
	return nil
}

func (m *Vote) GetChallenge() []byte {
	if m != nil {
		return m.Challenge
	}
	return nil
}

func (m *Vote) GetSignature() *stratumn_alice_core_crypto.Signature {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*Hello)(nil), "stratumn.alice.core.app.bootstrap.pb.Hello")
	proto.RegisterType((*Ack)(nil), "stratumn.alice.core.app.bootstrap.pb.Ack")
	proto.RegisterType((*CompleteReq)(nil), "stratumn.alice.core.app.bootstrap.pb.CompleteReq")
	proto.RegisterType((*Filter)(nil), "stratumn.alice.core.app.bootstrap.pb.Filter")
	proto.RegisterType((*PeerID)(nil), "stratumn.alice.core.app.bootstrap.pb.PeerID")
	proto.RegisterType((*NodeIdentity)(nil), "stratumn.alice.core.app.bootstrap.pb.NodeIdentity")
	proto.RegisterType((*UpdateProposal)(nil), "stratumn.alice.core.app.bootstrap.pb.UpdateProposal")
	proto.RegisterType((*Vote)(nil), "stratumn.alice.core.app.bootstrap.pb.Vote")
	proto.RegisterEnum("stratumn.alice.core.app.bootstrap.pb.UpdateType", UpdateType_name, UpdateType_value)
}
func (m *Hello) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Hello) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
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
	if len(m.Error) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.Error)))
		i += copy(dAtA[i:], m.Error)
	}
	return i, nil
}

func (m *CompleteReq) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CompleteReq) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *Filter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Filter) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *PeerID) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeerID) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	return i, nil
}

func (m *NodeIdentity) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NodeIdentity) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PeerId) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if len(m.PeerAddr) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.PeerAddr)))
		i += copy(dAtA[i:], m.PeerAddr)
	}
	if len(m.IdentityProof) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.IdentityProof)))
		i += copy(dAtA[i:], m.IdentityProof)
	}
	return i, nil
}

func (m *UpdateProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UpdateProposal) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.UpdateType != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(m.UpdateType))
	}
	if m.NodeDetails != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(m.NodeDetails.Size()))
		n1, err := m.NodeDetails.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if len(m.Challenge) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.Challenge)))
		i += copy(dAtA[i:], m.Challenge)
	}
	return i, nil
}

func (m *Vote) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Vote) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.UpdateType != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(m.UpdateType))
	}
	if len(m.PeerId) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.PeerId)))
		i += copy(dAtA[i:], m.PeerId)
	}
	if len(m.Challenge) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(len(m.Challenge)))
		i += copy(dAtA[i:], m.Challenge)
	}
	if m.Signature != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintBootstrap(dAtA, i, uint64(m.Signature.Size()))
		n2, err := m.Signature.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func encodeVarintBootstrap(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Hello) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *Ack) Size() (n int) {
	var l int
	_ = l
	l = len(m.Error)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	return n
}

func (m *CompleteReq) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *Filter) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *PeerID) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	return n
}

func (m *NodeIdentity) Size() (n int) {
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	l = len(m.PeerAddr)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	l = len(m.IdentityProof)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	return n
}

func (m *UpdateProposal) Size() (n int) {
	var l int
	_ = l
	if m.UpdateType != 0 {
		n += 1 + sovBootstrap(uint64(m.UpdateType))
	}
	if m.NodeDetails != nil {
		l = m.NodeDetails.Size()
		n += 1 + l + sovBootstrap(uint64(l))
	}
	l = len(m.Challenge)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	return n
}

func (m *Vote) Size() (n int) {
	var l int
	_ = l
	if m.UpdateType != 0 {
		n += 1 + sovBootstrap(uint64(m.UpdateType))
	}
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	l = len(m.Challenge)
	if l > 0 {
		n += 1 + l + sovBootstrap(uint64(l))
	}
	if m.Signature != nil {
		l = m.Signature.Size()
		n += 1 + l + sovBootstrap(uint64(l))
	}
	return n
}

func sovBootstrap(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozBootstrap(x uint64) (n int) {
	return sovBootstrap(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Hello) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: Hello: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Hello: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
				return ErrIntOverflowBootstrap
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
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Error", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Error = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func (m *CompleteReq) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: CompleteReq: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CompleteReq: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func (m *Filter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: Filter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Filter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func (m *PeerID) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: PeerID: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeerID: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
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
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func (m *NodeIdentity) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: NodeIdentity: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NodeIdentity: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
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
				return fmt.Errorf("proto: wrong wireType = %d for field PeerAddr", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PeerAddr = append(m.PeerAddr[:0], dAtA[iNdEx:postIndex]...)
			if m.PeerAddr == nil {
				m.PeerAddr = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IdentityProof", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IdentityProof = append(m.IdentityProof[:0], dAtA[iNdEx:postIndex]...)
			if m.IdentityProof == nil {
				m.IdentityProof = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func (m *UpdateProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: UpdateProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UpdateProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateType", wireType)
			}
			m.UpdateType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateType |= (UpdateType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeDetails", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.NodeDetails == nil {
				m.NodeDetails = &NodeIdentity{}
			}
			if err := m.NodeDetails.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Challenge", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Challenge = append(m.Challenge[:0], dAtA[iNdEx:postIndex]...)
			if m.Challenge == nil {
				m.Challenge = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func (m *Vote) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBootstrap
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
			return fmt.Errorf("proto: Vote: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Vote: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateType", wireType)
			}
			m.UpdateType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateType |= (UpdateType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Challenge", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Challenge = append(m.Challenge[:0], dAtA[iNdEx:postIndex]...)
			if m.Challenge == nil {
				m.Challenge = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBootstrap
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
				return ErrInvalidLengthBootstrap
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Signature == nil {
				m.Signature = &stratumn_alice_core_crypto.Signature{}
			}
			if err := m.Signature.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBootstrap(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBootstrap
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
func skipBootstrap(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowBootstrap
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
					return 0, ErrIntOverflowBootstrap
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
					return 0, ErrIntOverflowBootstrap
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
				return 0, ErrInvalidLengthBootstrap
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowBootstrap
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
				next, err := skipBootstrap(dAtA[start:])
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
	ErrInvalidLengthBootstrap = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowBootstrap   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/alice/core/app/bootstrap/pb/bootstrap.proto", fileDescriptorBootstrap)
}

var fileDescriptorBootstrap = []byte{
	// 530 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x93, 0xcd, 0x8a, 0x13, 0x41,
	0x10, 0xc7, 0xb7, 0xf7, 0x23, 0x1f, 0x95, 0x6c, 0x08, 0xad, 0xe0, 0xb0, 0x4a, 0x08, 0x83, 0xc2,
	0x2a, 0xeb, 0x8c, 0x64, 0x41, 0x10, 0x0f, 0x92, 0x0f, 0xc4, 0x20, 0x2c, 0x71, 0xdc, 0x15, 0xf4,
	0x12, 0x7a, 0xa6, 0xcb, 0xec, 0xe0, 0x64, 0xba, 0xed, 0xe9, 0x88, 0x79, 0x13, 0xcf, 0x39, 0x88,
	0xef, 0x90, 0x17, 0xf0, 0xe8, 0x45, 0xf0, 0x28, 0xf1, 0x05, 0x7c, 0x04, 0xe9, 0x9e, 0x64, 0xb3,
	0x07, 0x3f, 0x76, 0x0f, 0x1e, 0x86, 0x99, 0xea, 0xfe, 0xfd, 0xab, 0xff, 0x55, 0x35, 0x0d, 0x8f,
	0x46, 0xb1, 0x3e, 0x9d, 0x84, 0x5e, 0x24, 0xc6, 0x7e, 0xa6, 0x15, 0xd3, 0x93, 0x71, 0xea, 0xb3,
	0x24, 0x8e, 0xd0, 0x8f, 0x84, 0x42, 0x9f, 0x49, 0xe9, 0x87, 0x42, 0x68, 0xb3, 0x27, 0x7d, 0x19,
	0xae, 0x03, 0x4f, 0x2a, 0xa1, 0x05, 0xbd, 0xb9, 0x52, 0x79, 0x56, 0xe5, 0x19, 0x95, 0xc7, 0xa4,
	0xf4, 0xce, 0x81, 0xe1, 0xde, 0xc1, 0x9f, 0x8f, 0x19, 0x29, 0x19, 0xf9, 0xf8, 0x5e, 0x9b, 0x27,
	0xcf, 0xb9, 0xd7, 0xfa, 0x87, 0xa9, 0x48, 0x4d, 0xa5, 0x16, 0xcb, 0x57, 0xae, 0x71, 0x8b, 0xb0,
	0xf3, 0x04, 0x93, 0x44, 0xb8, 0xd7, 0x61, 0xab, 0x1d, 0xbd, 0xa1, 0x57, 0x61, 0x07, 0x95, 0x12,
	0xca, 0x21, 0x4d, 0xb2, 0x5f, 0x0e, 0xf2, 0xc0, 0xdd, 0x85, 0x4a, 0x57, 0x8c, 0x65, 0x82, 0x1a,
	0x03, 0x7c, 0xeb, 0x96, 0xa0, 0xf0, 0x38, 0x4e, 0x34, 0x2a, 0xf7, 0x3e, 0x14, 0x06, 0x88, 0xaa,
	0xdf, 0xa3, 0x07, 0x50, 0x94, 0x88, 0x6a, 0x18, 0x73, 0x2b, 0xad, 0x76, 0xae, 0xcc, 0xe6, 0x4e,
	0xd1, 0x6c, 0x36, 0xfb, 0xbd, 0x4f, 0x73, 0x87, 0xfc, 0x9c, 0x3b, 0x24, 0x28, 0x18, 0xa6, 0xcf,
	0xdd, 0x8f, 0x04, 0xaa, 0x47, 0x82, 0x63, 0x9f, 0x63, 0xaa, 0x63, 0x3d, 0xbd, 0x9c, 0x9c, 0xde,
	0x85, 0xb2, 0xa5, 0x19, 0xe7, 0xca, 0xd9, 0xb4, 0x7c, 0x7d, 0x36, 0x77, 0xaa, 0x96, 0x6f, 0x73,
	0xae, 0x30, 0xcb, 0x82, 0x92, 0x41, 0x4c, 0x40, 0x1f, 0x40, 0x2d, 0x5e, 0x1e, 0x34, 0x94, 0x4a,
	0x88, 0xd7, 0xce, 0x96, 0xd5, 0xd0, 0xd9, 0xdc, 0xa9, 0x3d, 0x7d, 0xd9, 0x6d, 0x72, 0xa6, 0x59,
	0x73, 0xff, 0xb8, 0xd3, 0xbb, 0x1d, 0xec, 0xae, 0xc8, 0x81, 0x01, 0xdd, 0x6f, 0x04, 0x6a, 0x27,
	0x92, 0x33, 0x8d, 0x03, 0x25, 0xa4, 0xc8, 0x58, 0x42, 0x9f, 0x41, 0x65, 0x62, 0x57, 0x86, 0x7a,
	0x2a, 0xd1, 0xda, 0xad, 0xb5, 0xee, 0x79, 0x17, 0x19, 0xa8, 0x97, 0xa7, 0x3a, 0x9e, 0x4a, 0x0c,
	0x60, 0x72, 0xf6, 0x4d, 0x4f, 0xa0, 0x9a, 0x0a, 0x8e, 0x43, 0x8e, 0x9a, 0xc5, 0x49, 0x66, 0x4b,
	0xaa, 0xb4, 0x5a, 0x17, 0xcb, 0x79, 0xbe, 0x8f, 0x41, 0xc5, 0xe4, 0xe9, 0xe5, 0x69, 0xe8, 0x0d,
	0x28, 0x47, 0xa7, 0x2c, 0x49, 0x30, 0x1d, 0x61, 0x5e, 0x72, 0xb0, 0x5e, 0x70, 0xbf, 0x12, 0xd8,
	0x7e, 0x21, 0x34, 0xfe, 0x8f, 0x82, 0xae, 0xad, 0xc7, 0x69, 0xc7, 0x73, 0x36, 0xb9, 0xbf, 0x5a,
	0xa2, 0x5d, 0x28, 0x67, 0xf1, 0x28, 0x65, 0x7a, 0xa2, 0xd0, 0xd9, 0xb6, 0x4d, 0xb8, 0xf5, 0x5b,
	0x1f, 0xcb, 0x7f, 0xf8, 0xf9, 0x0a, 0x0e, 0xd6, 0xba, 0x3b, 0x87, 0x00, 0x6b, 0x57, 0xb4, 0x04,
	0xdb, 0x47, 0x22, 0xc5, 0xfa, 0x06, 0xad, 0x40, 0xb1, 0xcd, 0xb9, 0xe9, 0x56, 0x9d, 0xd0, 0x1a,
	0x40, 0x80, 0x63, 0xf1, 0x0e, 0x6d, 0xbc, 0xd9, 0xe9, 0x7d, 0x5e, 0x34, 0xc8, 0x97, 0x45, 0x83,
	0x7c, 0x5f, 0x34, 0xc8, 0x87, 0x1f, 0x8d, 0x8d, 0x57, 0xad, 0x4b, 0x5e, 0xf1, 0x87, 0x32, 0x0c,
	0x0b, 0xf6, 0x52, 0x1d, 0xfe, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x2e, 0xd6, 0xb6, 0x9c, 0x1f, 0x04,
	0x00, 0x00,
}
