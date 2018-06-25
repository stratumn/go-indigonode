// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/core/crypto/crypto.proto

/*
	Package crypto is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/core/crypto/crypto.proto

	It has these top-level messages:
		Signature
*/
package crypto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

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

// Types of digital keys supported.
type KeyType int32

const (
	KeyType_RSA       KeyType = 0
	KeyType_ECDSA     KeyType = 1
	KeyType_Ed25519   KeyType = 2
	KeyType_Secp256k1 KeyType = 3
)

var KeyType_name = map[int32]string{
	0: "RSA",
	1: "ECDSA",
	2: "Ed25519",
	3: "Secp256k1",
}
var KeyType_value = map[string]int32{
	"RSA":       0,
	"ECDSA":     1,
	"Ed25519":   2,
	"Secp256k1": 3,
}

func (x KeyType) String() string {
	return proto.EnumName(KeyType_name, int32(x))
}
func (KeyType) EnumDescriptor() ([]byte, []int) { return fileDescriptorCrypto, []int{0} }

// A digital signature.
type Signature struct {
	KeyType   KeyType `protobuf:"varint,1,opt,name=key_type,json=keyType,proto3,enum=stratumn.indigonode.core.crypto.KeyType" json:"key_type,omitempty"`
	PublicKey []byte  `protobuf:"bytes,2,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Signature []byte  `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *Signature) Reset()                    { *m = Signature{} }
func (m *Signature) String() string            { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()               {}
func (*Signature) Descriptor() ([]byte, []int) { return fileDescriptorCrypto, []int{0} }

func (m *Signature) GetKeyType() KeyType {
	if m != nil {
		return m.KeyType
	}
	return KeyType_RSA
}

func (m *Signature) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *Signature) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*Signature)(nil), "stratumn.indigonode.core.crypto.Signature")
	proto.RegisterEnum("stratumn.indigonode.core.crypto.KeyType", KeyType_name, KeyType_value)
}
func (m *Signature) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Signature) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.KeyType != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintCrypto(dAtA, i, uint64(m.KeyType))
	}
	if len(m.PublicKey) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintCrypto(dAtA, i, uint64(len(m.PublicKey)))
		i += copy(dAtA[i:], m.PublicKey)
	}
	if len(m.Signature) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintCrypto(dAtA, i, uint64(len(m.Signature)))
		i += copy(dAtA[i:], m.Signature)
	}
	return i, nil
}

func encodeVarintCrypto(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Signature) Size() (n int) {
	var l int
	_ = l
	if m.KeyType != 0 {
		n += 1 + sovCrypto(uint64(m.KeyType))
	}
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovCrypto(uint64(l))
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovCrypto(uint64(l))
	}
	return n
}

func sovCrypto(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozCrypto(x uint64) (n int) {
	return sovCrypto(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Signature) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCrypto
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
			return fmt.Errorf("proto: Signature: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Signature: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field KeyType", wireType)
			}
			m.KeyType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrypto
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.KeyType |= (KeyType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrypto
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
				return ErrInvalidLengthCrypto
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = append(m.PublicKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PublicKey == nil {
				m.PublicKey = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrypto
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
				return ErrInvalidLengthCrypto
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCrypto(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCrypto
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
func skipCrypto(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCrypto
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
					return 0, ErrIntOverflowCrypto
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
					return 0, ErrIntOverflowCrypto
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
				return 0, ErrInvalidLengthCrypto
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowCrypto
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
				next, err := skipCrypto(dAtA[start:])
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
	ErrInvalidLengthCrypto = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCrypto   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-indigonode/core/crypto/crypto.proto", fileDescriptorCrypto)
}

var fileDescriptorCrypto = []byte{
	// 267 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xb2, 0x4a, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x2f, 0x2e, 0x29, 0x4a, 0x2c, 0x29, 0xcd, 0xcd, 0xd3,
	0x4f, 0xcf, 0xd7, 0xcd, 0xcc, 0x4b, 0xc9, 0x4c, 0xcf, 0xcf, 0xcb, 0x4f, 0x49, 0xd5, 0x4f, 0xce,
	0x2f, 0x4a, 0xd5, 0x4f, 0x2e, 0xaa, 0x2c, 0x28, 0xc9, 0x87, 0x52, 0x7a, 0x05, 0x45, 0xf9, 0x25,
	0xf9, 0x42, 0xf2, 0x30, 0x0d, 0x7a, 0x08, 0xd5, 0x7a, 0x20, 0xd5, 0x7a, 0x10, 0x65, 0x4a, 0xbd,
	0x8c, 0x5c, 0x9c, 0xc1, 0x99, 0xe9, 0x79, 0x89, 0x25, 0xa5, 0x45, 0xa9, 0x42, 0xce, 0x5c, 0x1c,
	0xd9, 0xa9, 0x95, 0xf1, 0x25, 0x95, 0x05, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x7c, 0x46, 0x1a,
	0x7a, 0x04, 0x4c, 0xd0, 0xf3, 0x4e, 0xad, 0x0c, 0xa9, 0x2c, 0x48, 0x0d, 0x62, 0xcf, 0x86, 0x30,
	0x84, 0x64, 0xb9, 0xb8, 0x0a, 0x4a, 0x93, 0x72, 0x32, 0x93, 0xe3, 0xb3, 0x53, 0x2b, 0x25, 0x98,
	0x14, 0x18, 0x35, 0x78, 0x82, 0x38, 0x21, 0x22, 0xde, 0xa9, 0x95, 0x42, 0x32, 0x5c, 0x9c, 0xc5,
	0x30, 0x0b, 0x25, 0x98, 0x21, 0xb2, 0x70, 0x01, 0x2d, 0x4b, 0x2e, 0x76, 0xa8, 0x81, 0x42, 0xec,
	0x5c, 0xcc, 0x41, 0xc1, 0x8e, 0x02, 0x0c, 0x42, 0x9c, 0x5c, 0xac, 0xae, 0xce, 0x2e, 0xc1, 0x8e,
	0x02, 0x8c, 0x42, 0xdc, 0x5c, 0xec, 0xae, 0x29, 0x46, 0xa6, 0xa6, 0x86, 0x96, 0x02, 0x4c, 0x42,
	0xbc, 0x5c, 0x9c, 0xc1, 0xa9, 0xc9, 0x05, 0x46, 0xa6, 0x66, 0xd9, 0x86, 0x02, 0xcc, 0x4e, 0x6e,
	0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3, 0x8c, 0xc7, 0x72,
	0x0c, 0x51, 0x26, 0x24, 0x85, 0x9c, 0x35, 0x84, 0x4a, 0x62, 0x03, 0x07, 0x9d, 0x31, 0x20, 0x00,
	0x00, 0xff, 0xff, 0xb9, 0xae, 0x9a, 0x0b, 0x78, 0x01, 0x00, 0x00,
}
