// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/core/protector/pb/protector.proto

/*
	Package pb is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/core/protector/pb/protector.proto

	It has these top-level messages:
		PeerAddrs
		NetworkConfig
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import stratumn_indigonode_core_crypto "github.com/stratumn/go-indigonode/core/crypto"

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

type NetworkState int32

const (
	// The network is in the bootstrap phase (not fully private yet).
	NetworkState_BOOTSTRAP NetworkState = 0
	// The network bootstrap phase is complete and the network is now protected.
	NetworkState_PROTECTED NetworkState = 1
)

var NetworkState_name = map[int32]string{
	0: "BOOTSTRAP",
	1: "PROTECTED",
}
var NetworkState_value = map[string]int32{
	"BOOTSTRAP": 0,
	"PROTECTED": 1,
}

func (x NetworkState) String() string {
	return proto.EnumName(NetworkState_name, int32(x))
}
func (NetworkState) EnumDescriptor() ([]byte, []int) { return fileDescriptorProtector, []int{0} }

// A list of multiaddresses.
type PeerAddrs struct {
	Addresses []string `protobuf:"bytes,1,rep,name=addresses" json:"addresses,omitempty"`
}

func (m *PeerAddrs) Reset()                    { *m = PeerAddrs{} }
func (m *PeerAddrs) String() string            { return proto.CompactTextString(m) }
func (*PeerAddrs) ProtoMessage()               {}
func (*PeerAddrs) Descriptor() ([]byte, []int) { return fileDescriptorProtector, []int{0} }

func (m *PeerAddrs) GetAddresses() []string {
	if m != nil {
		return m.Addresses
	}
	return nil
}

// The NetworkConfig message contains the network state
// and its participants (signed by a trusted node).
type NetworkConfig struct {
	NetworkState NetworkState                               `protobuf:"varint,1,opt,name=network_state,json=networkState,proto3,enum=stratumn.indigonode.core.protector.NetworkState" json:"network_state,omitempty"`
	Participants map[string]*PeerAddrs                      `protobuf:"bytes,2,rep,name=participants" json:"participants,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value"`
	Signature    *stratumn_indigonode_core_crypto.Signature `protobuf:"bytes,10,opt,name=signature" json:"signature,omitempty"`
}

func (m *NetworkConfig) Reset()                    { *m = NetworkConfig{} }
func (m *NetworkConfig) String() string            { return proto.CompactTextString(m) }
func (*NetworkConfig) ProtoMessage()               {}
func (*NetworkConfig) Descriptor() ([]byte, []int) { return fileDescriptorProtector, []int{1} }

func (m *NetworkConfig) GetNetworkState() NetworkState {
	if m != nil {
		return m.NetworkState
	}
	return NetworkState_BOOTSTRAP
}

func (m *NetworkConfig) GetParticipants() map[string]*PeerAddrs {
	if m != nil {
		return m.Participants
	}
	return nil
}

func (m *NetworkConfig) GetSignature() *stratumn_indigonode_core_crypto.Signature {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*PeerAddrs)(nil), "stratumn.indigonode.core.protector.PeerAddrs")
	proto.RegisterType((*NetworkConfig)(nil), "stratumn.indigonode.core.protector.NetworkConfig")
	proto.RegisterEnum("stratumn.indigonode.core.protector.NetworkState", NetworkState_name, NetworkState_value)
}
func (m *PeerAddrs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeerAddrs) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Addresses) > 0 {
		for _, s := range m.Addresses {
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

func (m *NetworkConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NetworkConfig) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.NetworkState != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintProtector(dAtA, i, uint64(m.NetworkState))
	}
	if len(m.Participants) > 0 {
		for k, _ := range m.Participants {
			dAtA[i] = 0x12
			i++
			v := m.Participants[k]
			msgSize := 0
			if v != nil {
				msgSize = v.Size()
				msgSize += 1 + sovProtector(uint64(msgSize))
			}
			mapSize := 1 + len(k) + sovProtector(uint64(len(k))) + msgSize
			i = encodeVarintProtector(dAtA, i, uint64(mapSize))
			dAtA[i] = 0xa
			i++
			i = encodeVarintProtector(dAtA, i, uint64(len(k)))
			i += copy(dAtA[i:], k)
			if v != nil {
				dAtA[i] = 0x12
				i++
				i = encodeVarintProtector(dAtA, i, uint64(v.Size()))
				n1, err := v.MarshalTo(dAtA[i:])
				if err != nil {
					return 0, err
				}
				i += n1
			}
		}
	}
	if m.Signature != nil {
		dAtA[i] = 0x52
		i++
		i = encodeVarintProtector(dAtA, i, uint64(m.Signature.Size()))
		n2, err := m.Signature.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n2
	}
	return i, nil
}

func encodeVarintProtector(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *PeerAddrs) Size() (n int) {
	var l int
	_ = l
	if len(m.Addresses) > 0 {
		for _, s := range m.Addresses {
			l = len(s)
			n += 1 + l + sovProtector(uint64(l))
		}
	}
	return n
}

func (m *NetworkConfig) Size() (n int) {
	var l int
	_ = l
	if m.NetworkState != 0 {
		n += 1 + sovProtector(uint64(m.NetworkState))
	}
	if len(m.Participants) > 0 {
		for k, v := range m.Participants {
			_ = k
			_ = v
			l = 0
			if v != nil {
				l = v.Size()
				l += 1 + sovProtector(uint64(l))
			}
			mapEntrySize := 1 + len(k) + sovProtector(uint64(len(k))) + l
			n += mapEntrySize + 1 + sovProtector(uint64(mapEntrySize))
		}
	}
	if m.Signature != nil {
		l = m.Signature.Size()
		n += 1 + l + sovProtector(uint64(l))
	}
	return n
}

func sovProtector(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozProtector(x uint64) (n int) {
	return sovProtector(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PeerAddrs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtector
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
			return fmt.Errorf("proto: PeerAddrs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeerAddrs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Addresses", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtector
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
				return ErrInvalidLengthProtector
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Addresses = append(m.Addresses, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtector(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtector
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
func (m *NetworkConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtector
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
			return fmt.Errorf("proto: NetworkConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NetworkConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NetworkState", wireType)
			}
			m.NetworkState = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtector
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NetworkState |= (NetworkState(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Participants", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtector
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
				return ErrInvalidLengthProtector
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Participants == nil {
				m.Participants = make(map[string]*PeerAddrs)
			}
			var mapkey string
			var mapvalue *PeerAddrs
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowProtector
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
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowProtector
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= (uint64(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthProtector
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowProtector
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= (int(b) & 0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthProtector
					}
					postmsgIndex := iNdEx + mapmsglen
					if mapmsglen < 0 {
						return ErrInvalidLengthProtector
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &PeerAddrs{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipProtector(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthProtector
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Participants[mapkey] = mapvalue
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtector
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
				return ErrInvalidLengthProtector
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Signature == nil {
				m.Signature = &stratumn_indigonode_core_crypto.Signature{}
			}
			if err := m.Signature.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtector(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtector
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
func skipProtector(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProtector
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
					return 0, ErrIntOverflowProtector
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
					return 0, ErrIntOverflowProtector
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
				return 0, ErrInvalidLengthProtector
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowProtector
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
				next, err := skipProtector(dAtA[start:])
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
	ErrInvalidLengthProtector = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProtector   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-indigonode/core/protector/pb/protector.proto", fileDescriptorProtector)
}

var fileDescriptorProtector = []byte{
	// 379 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0xdf, 0x6e, 0xda, 0x30,
	0x14, 0xc6, 0x31, 0xd1, 0x26, 0xc5, 0xc0, 0xc4, 0x7c, 0x15, 0xa1, 0x29, 0x8a, 0xb8, 0xca, 0xd0,
	0x70, 0x26, 0x26, 0x4d, 0x13, 0xbb, 0x82, 0x0c, 0x89, 0xab, 0x11, 0x99, 0xf4, 0xa6, 0x37, 0x55,
	0xfe, 0xb8, 0x69, 0x44, 0xb1, 0x23, 0xdb, 0x69, 0xc5, 0x23, 0xf4, 0x0d, 0xfa, 0x48, 0xbd, 0xec,
	0x23, 0x54, 0xf4, 0x45, 0xaa, 0x10, 0x20, 0x54, 0x55, 0x55, 0xda, 0x2b, 0x9f, 0x73, 0xe4, 0xdf,
	0xf7, 0x1d, 0x1f, 0x1f, 0xe8, 0x26, 0xa9, 0xba, 0xc8, 0x43, 0x1c, 0xf1, 0xa5, 0x23, 0x95, 0x08,
	0x54, 0xbe, 0x64, 0x4e, 0xc2, 0xfb, 0x29, 0x8b, 0xd3, 0x84, 0x33, 0x1e, 0x53, 0x27, 0xe2, 0x82,
	0x3a, 0x99, 0xe0, 0x8a, 0x46, 0x8a, 0x0b, 0x27, 0x0b, 0xab, 0x04, 0x17, 0x11, 0x47, 0xdd, 0x1d,
	0x89, 0x2b, 0x0c, 0x17, 0x18, 0xde, 0xdf, 0xec, 0x0c, 0x8f, 0x34, 0x8a, 0xc4, 0x2a, 0x53, 0x7c,
	0x7b, 0x94, 0xfa, 0xdd, 0xef, 0x50, 0xf7, 0x28, 0x15, 0xa3, 0x38, 0x16, 0x12, 0x7d, 0x83, 0x7a,
	0x10, 0xc7, 0x82, 0x4a, 0x49, 0xa5, 0x01, 0x2c, 0xcd, 0xd6, 0x49, 0x55, 0xe8, 0xde, 0x68, 0xb0,
	0xf5, 0x9f, 0xaa, 0x6b, 0x2e, 0x16, 0x2e, 0x67, 0xe7, 0x69, 0x82, 0x4e, 0x60, 0x8b, 0x95, 0x85,
	0x33, 0xa9, 0x02, 0x45, 0x0d, 0x60, 0x01, 0xfb, 0xcb, 0xe0, 0x27, 0x7e, 0xbb, 0x69, 0xbc, 0x55,
	0x9a, 0x17, 0x1c, 0x69, 0xb2, 0x83, 0x0c, 0x25, 0xb0, 0x99, 0x05, 0x42, 0xa5, 0x51, 0x9a, 0x05,
	0x4c, 0x49, 0xa3, 0x6e, 0x69, 0x76, 0x63, 0xe0, 0xbe, 0x43, 0xb5, 0xec, 0x0f, 0x7b, 0x07, 0x2a,
	0x13, 0xa6, 0xc4, 0x8a, 0x3c, 0x13, 0x46, 0x53, 0xa8, 0xcb, 0x34, 0x61, 0x81, 0xca, 0x05, 0x35,
	0xa0, 0x05, 0xec, 0xc6, 0xa0, 0xf7, 0xba, 0xcb, 0x76, 0x6e, 0xf3, 0x1d, 0x41, 0x2a, 0xb8, 0xc3,
	0xe0, 0xd7, 0x17, 0x66, 0xa8, 0x0d, 0xb5, 0x05, 0x5d, 0x6d, 0x86, 0xa2, 0x93, 0x22, 0x44, 0x2e,
	0xfc, 0x74, 0x15, 0x5c, 0xe6, 0xd4, 0xa8, 0x6f, 0xcc, 0xfa, 0xc7, 0x3c, 0x69, 0xff, 0x3d, 0xa4,
	0x64, 0x87, 0xf5, 0x3f, 0xa0, 0xf7, 0x03, 0x36, 0x0f, 0x07, 0x88, 0x5a, 0x50, 0x1f, 0xcf, 0x66,
	0xfe, 0xdc, 0x27, 0x23, 0xaf, 0x5d, 0x2b, 0x52, 0x8f, 0xcc, 0xfc, 0x89, 0xeb, 0x4f, 0xfe, 0xb5,
	0xc1, 0x78, 0x7a, 0xb7, 0x36, 0xc1, 0xfd, 0xda, 0x04, 0x0f, 0x6b, 0x13, 0xdc, 0x3e, 0x9a, 0xb5,
	0xd3, 0xdf, 0x1f, 0xd8, 0xcd, 0xbf, 0x59, 0x18, 0x7e, 0xde, 0x6c, 0xcd, 0xaf, 0xa7, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x72, 0x94, 0xc9, 0xde, 0xdc, 0x02, 0x00, 0x00,
}
