// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/core/app/bootstrap/pb/bootstrap.proto

/*
	Package pb is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/core/app/bootstrap/pb/bootstrap.proto

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
import _ "github.com/stratumn/go-indigonode/cli/grpc/ext"
import stratumn_alice_core_crypto "github.com/stratumn/go-indigonode/core/crypto"

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
	UpdateType  UpdateType    `protobuf:"varint,1,opt,name=update_type,json=updateType,proto3,enum=stratumn.alice.core.app.bootstrap.UpdateType" json:"update_type,omitempty"`
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
	UpdateType UpdateType                            `protobuf:"varint,1,opt,name=update_type,json=updateType,proto3,enum=stratumn.alice.core.app.bootstrap.UpdateType" json:"update_type,omitempty"`
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
	proto.RegisterType((*Hello)(nil), "stratumn.alice.core.app.bootstrap.Hello")
	proto.RegisterType((*Ack)(nil), "stratumn.alice.core.app.bootstrap.Ack")
	proto.RegisterType((*CompleteReq)(nil), "stratumn.alice.core.app.bootstrap.CompleteReq")
	proto.RegisterType((*Filter)(nil), "stratumn.alice.core.app.bootstrap.Filter")
	proto.RegisterType((*PeerID)(nil), "stratumn.alice.core.app.bootstrap.PeerID")
	proto.RegisterType((*NodeIdentity)(nil), "stratumn.alice.core.app.bootstrap.NodeIdentity")
	proto.RegisterType((*UpdateProposal)(nil), "stratumn.alice.core.app.bootstrap.UpdateProposal")
	proto.RegisterType((*Vote)(nil), "stratumn.alice.core.app.bootstrap.Vote")
	proto.RegisterEnum("stratumn.alice.core.app.bootstrap.UpdateType", UpdateType_name, UpdateType_value)
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
	proto.RegisterFile("github.com/stratumn/go-indigonode/core/app/bootstrap/pb/bootstrap.proto", fileDescriptorBootstrap)
}

var fileDescriptorBootstrap = []byte{
	// 526 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0xdd, 0x8a, 0xd3, 0x40,
	0x14, 0xde, 0xec, 0x4f, 0x7f, 0x4e, 0xbb, 0xa5, 0x8c, 0x82, 0x61, 0x95, 0x52, 0x03, 0xc2, 0x2a,
	0x6e, 0x22, 0x5d, 0x10, 0xc4, 0x0b, 0xe9, 0x0f, 0x62, 0x11, 0x4a, 0x19, 0x57, 0x41, 0x6f, 0xca,
	0x34, 0x73, 0xec, 0x06, 0xd3, 0xcc, 0x38, 0x99, 0x8a, 0x7d, 0x13, 0xaf, 0x7b, 0x21, 0xbe, 0x43,
	0x5f, 0xc0, 0x4b, 0xf1, 0x09, 0xa4, 0xbe, 0x80, 0x8f, 0x20, 0x33, 0x69, 0x1b, 0x2f, 0x64, 0xd7,
	0x05, 0x2f, 0x42, 0x72, 0x66, 0xbe, 0xef, 0x9b, 0xef, 0x9b, 0x93, 0x03, 0x4f, 0x26, 0x91, 0x3e,
	0x9f, 0x8d, 0xfd, 0x50, 0x4c, 0x83, 0x54, 0x2b, 0xa6, 0x67, 0xd3, 0x24, 0x60, 0x71, 0x14, 0x62,
	0x10, 0x0a, 0x85, 0x01, 0x93, 0x32, 0x18, 0x0b, 0xa1, 0xcd, 0x9e, 0x0c, 0xe4, 0x38, 0x2f, 0x7c,
	0xa9, 0x84, 0x16, 0xe4, 0xf6, 0x86, 0xe5, 0x5b, 0x96, 0x6f, 0x58, 0x3e, 0x93, 0xd2, 0xdf, 0x02,
	0x8f, 0x1e, 0x5c, 0x70, 0x46, 0x1c, 0x05, 0x13, 0x25, 0xc3, 0x00, 0x3f, 0x6a, 0xf3, 0x64, 0xa2,
	0x47, 0xad, 0x4b, 0x5c, 0x85, 0x6a, 0x2e, 0xb5, 0x58, 0xbf, 0x32, 0x8e, 0x57, 0x84, 0x83, 0x67,
	0x18, 0xc7, 0xc2, 0xbb, 0x09, 0x7b, 0xed, 0xf0, 0x1d, 0xb9, 0x0e, 0x07, 0xa8, 0x94, 0x50, 0xae,
	0xd3, 0x74, 0x8e, 0xcb, 0x34, 0x2b, 0xbc, 0x43, 0xa8, 0x74, 0xc5, 0x54, 0xc6, 0xa8, 0x91, 0xe2,
	0x7b, 0xaf, 0x04, 0x85, 0xa7, 0x51, 0xac, 0x51, 0x79, 0x0f, 0xa1, 0x30, 0x44, 0x54, 0xfd, 0x1e,
	0xb9, 0x0f, 0x45, 0x89, 0xa8, 0x46, 0x11, 0xb7, 0xd4, 0x6a, 0xe7, 0xda, 0x62, 0xe9, 0x16, 0xcd,
	0x66, 0xb3, 0xdf, 0xfb, 0xb2, 0x74, 0x9d, 0x5f, 0x4b, 0xd7, 0xa1, 0x05, 0x83, 0xe9, 0x73, 0xef,
	0xb3, 0x03, 0xd5, 0x81, 0xe0, 0xd8, 0xe7, 0x98, 0xe8, 0x48, 0xcf, 0xaf, 0x46, 0x27, 0x27, 0x50,
	0xb6, 0x68, 0xc6, 0xb9, 0x72, 0x77, 0x2d, 0xbe, 0xbe, 0x58, 0xba, 0x55, 0x8b, 0x6f, 0x73, 0xae,
	0x30, 0x4d, 0x69, 0xc9, 0x40, 0x4c, 0x41, 0x1e, 0x41, 0x2d, 0x5a, 0x1f, 0x34, 0x92, 0x4a, 0x88,
	0xb7, 0xee, 0x9e, 0xe5, 0x90, 0xc5, 0xd2, 0xad, 0x3d, 0x7f, 0xdd, 0x6d, 0x72, 0xa6, 0x59, 0xf3,
	0xf8, 0xac, 0xd3, 0xbb, 0x4b, 0x0f, 0x37, 0xc8, 0xa1, 0x01, 0x7a, 0xdf, 0x1d, 0xa8, 0xbd, 0x94,
	0x9c, 0x69, 0x1c, 0x2a, 0x21, 0x45, 0xca, 0x62, 0x32, 0x80, 0xca, 0xcc, 0xae, 0x8c, 0xf4, 0x5c,
	0xa2, 0xb5, 0x5b, 0x6b, 0x9d, 0xf8, 0x97, 0x76, 0xd4, 0xcf, 0x74, 0xce, 0xe6, 0x12, 0x29, 0xcc,
	0xb6, 0xdf, 0x84, 0x42, 0x35, 0x11, 0x1c, 0x47, 0x1c, 0x35, 0x8b, 0xe2, 0xd4, 0xe6, 0xa9, 0xb4,
	0x82, 0x7f, 0x10, 0xfc, 0xf3, 0x06, 0x69, 0xc5, 0x88, 0xf4, 0x32, 0x0d, 0x72, 0x0b, 0xca, 0xe1,
	0x39, 0x8b, 0x63, 0x4c, 0x26, 0x98, 0x85, 0xa5, 0xf9, 0x82, 0x09, 0xb5, 0xff, 0x4a, 0x68, 0xfc,
	0xef, 0x51, 0x6e, 0xe4, 0x5d, 0xb4, 0x5d, 0xd9, 0x36, 0xec, 0x42, 0x3f, 0xa4, 0x0b, 0xe5, 0x34,
	0x9a, 0x24, 0x4c, 0xcf, 0x14, 0xba, 0xfb, 0x36, 0xfe, 0x9d, 0xbf, 0x9a, 0x58, 0xff, 0xba, 0x2f,
	0x36, 0x60, 0x9a, 0xf3, 0xee, 0x9d, 0x02, 0xe4, 0xae, 0x48, 0x09, 0xf6, 0x07, 0x22, 0xc1, 0xfa,
	0x0e, 0xa9, 0x40, 0xb1, 0xcd, 0xb9, 0xb9, 0xaa, 0xba, 0x43, 0x6a, 0x00, 0x14, 0xa7, 0xe2, 0x03,
	0xda, 0x7a, 0xb7, 0xd3, 0xfb, 0xba, 0x6a, 0x38, 0xdf, 0x56, 0x0d, 0xe7, 0xc7, 0xaa, 0xe1, 0x7c,
	0xfa, 0xd9, 0xd8, 0x79, 0xd3, 0xba, 0xe2, 0x68, 0x3f, 0x96, 0xe3, 0x71, 0xc1, 0xce, 0xd2, 0xe9,
	0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0xa0, 0x0a, 0x6c, 0x32, 0x17, 0x04, 0x00, 0x00,
}
