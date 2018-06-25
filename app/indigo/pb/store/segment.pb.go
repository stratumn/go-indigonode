// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/stratumn/go-indigonode/app/indigo/pb/store/segment.proto

/*
	Package store is a generated protocol buffer package.

	It is generated from these files:
		github.com/stratumn/go-indigonode/app/indigo/pb/store/segment.proto

	It has these top-level messages:
		Link
		LinkHash
		LinkHashes
		Segment
		Segments
		Evidence
		Evidences
*/
package store

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/stratumn/go-indigonode/cli/grpc/ext"

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

type Link struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Link) Reset()                    { *m = Link{} }
func (m *Link) String() string            { return proto.CompactTextString(m) }
func (*Link) ProtoMessage()               {}
func (*Link) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{0} }

func (m *Link) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type LinkHash struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *LinkHash) Reset()                    { *m = LinkHash{} }
func (m *LinkHash) String() string            { return proto.CompactTextString(m) }
func (*LinkHash) ProtoMessage()               {}
func (*LinkHash) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{1} }

func (m *LinkHash) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type LinkHashes struct {
	LinkHashes []*LinkHash `protobuf:"bytes,1,rep,name=link_hashes,json=linkHashes" json:"link_hashes,omitempty"`
}

func (m *LinkHashes) Reset()                    { *m = LinkHashes{} }
func (m *LinkHashes) String() string            { return proto.CompactTextString(m) }
func (*LinkHashes) ProtoMessage()               {}
func (*LinkHashes) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{2} }

func (m *LinkHashes) GetLinkHashes() []*LinkHash {
	if m != nil {
		return m.LinkHashes
	}
	return nil
}

type Segment struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Segment) Reset()                    { *m = Segment{} }
func (m *Segment) String() string            { return proto.CompactTextString(m) }
func (*Segment) ProtoMessage()               {}
func (*Segment) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{3} }

func (m *Segment) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Segments struct {
	Segments []*Segment `protobuf:"bytes,1,rep,name=segments" json:"segments,omitempty"`
}

func (m *Segments) Reset()                    { *m = Segments{} }
func (m *Segments) String() string            { return proto.CompactTextString(m) }
func (*Segments) ProtoMessage()               {}
func (*Segments) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{4} }

func (m *Segments) GetSegments() []*Segment {
	if m != nil {
		return m.Segments
	}
	return nil
}

type Evidence struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Evidence) Reset()                    { *m = Evidence{} }
func (m *Evidence) String() string            { return proto.CompactTextString(m) }
func (*Evidence) ProtoMessage()               {}
func (*Evidence) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{5} }

func (m *Evidence) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Evidences struct {
	Evidences []*Evidence `protobuf:"bytes,1,rep,name=evidences" json:"evidences,omitempty"`
}

func (m *Evidences) Reset()                    { *m = Evidences{} }
func (m *Evidences) String() string            { return proto.CompactTextString(m) }
func (*Evidences) ProtoMessage()               {}
func (*Evidences) Descriptor() ([]byte, []int) { return fileDescriptorSegment, []int{6} }

func (m *Evidences) GetEvidences() []*Evidence {
	if m != nil {
		return m.Evidences
	}
	return nil
}

func init() {
	proto.RegisterType((*Link)(nil), "stratumn.indigonode.app.indigo.store.Link")
	proto.RegisterType((*LinkHash)(nil), "stratumn.indigonode.app.indigo.store.LinkHash")
	proto.RegisterType((*LinkHashes)(nil), "stratumn.indigonode.app.indigo.store.LinkHashes")
	proto.RegisterType((*Segment)(nil), "stratumn.indigonode.app.indigo.store.Segment")
	proto.RegisterType((*Segments)(nil), "stratumn.indigonode.app.indigo.store.Segments")
	proto.RegisterType((*Evidence)(nil), "stratumn.indigonode.app.indigo.store.Evidence")
	proto.RegisterType((*Evidences)(nil), "stratumn.indigonode.app.indigo.store.Evidences")
}
func (m *Link) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Link) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSegment(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *LinkHash) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LinkHash) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSegment(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *LinkHashes) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LinkHashes) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.LinkHashes) > 0 {
		for _, msg := range m.LinkHashes {
			dAtA[i] = 0xa
			i++
			i = encodeVarintSegment(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *Segment) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Segment) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSegment(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *Segments) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Segments) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Segments) > 0 {
		for _, msg := range m.Segments {
			dAtA[i] = 0xa
			i++
			i = encodeVarintSegment(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *Evidence) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Evidence) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSegment(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *Evidences) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Evidences) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Evidences) > 0 {
		for _, msg := range m.Evidences {
			dAtA[i] = 0xa
			i++
			i = encodeVarintSegment(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func encodeVarintSegment(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Link) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovSegment(uint64(l))
	}
	return n
}

func (m *LinkHash) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovSegment(uint64(l))
	}
	return n
}

func (m *LinkHashes) Size() (n int) {
	var l int
	_ = l
	if len(m.LinkHashes) > 0 {
		for _, e := range m.LinkHashes {
			l = e.Size()
			n += 1 + l + sovSegment(uint64(l))
		}
	}
	return n
}

func (m *Segment) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovSegment(uint64(l))
	}
	return n
}

func (m *Segments) Size() (n int) {
	var l int
	_ = l
	if len(m.Segments) > 0 {
		for _, e := range m.Segments {
			l = e.Size()
			n += 1 + l + sovSegment(uint64(l))
		}
	}
	return n
}

func (m *Evidence) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovSegment(uint64(l))
	}
	return n
}

func (m *Evidences) Size() (n int) {
	var l int
	_ = l
	if len(m.Evidences) > 0 {
		for _, e := range m.Evidences {
			l = e.Size()
			n += 1 + l + sovSegment(uint64(l))
		}
	}
	return n
}

func sovSegment(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozSegment(x uint64) (n int) {
	return sovSegment(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Link) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: Link: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Link: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
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
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func (m *LinkHash) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: LinkHash: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LinkHash: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
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
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func (m *LinkHashes) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: LinkHashes: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LinkHashes: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LinkHashes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LinkHashes = append(m.LinkHashes, &LinkHash{})
			if err := m.LinkHashes[len(m.LinkHashes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func (m *Segment) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: Segment: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Segment: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
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
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func (m *Segments) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: Segments: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Segments: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Segments", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Segments = append(m.Segments, &Segment{})
			if err := m.Segments[len(m.Segments)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func (m *Evidence) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: Evidence: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Evidence: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
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
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func (m *Evidences) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSegment
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
			return fmt.Errorf("proto: Evidences: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Evidences: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Evidences", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSegment
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
				return ErrInvalidLengthSegment
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Evidences = append(m.Evidences, &Evidence{})
			if err := m.Evidences[len(m.Evidences)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSegment(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSegment
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
func skipSegment(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSegment
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
					return 0, ErrIntOverflowSegment
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
					return 0, ErrIntOverflowSegment
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
				return 0, ErrInvalidLengthSegment
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowSegment
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
				next, err := skipSegment(dAtA[start:])
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
	ErrInvalidLengthSegment = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSegment   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/stratumn/go-indigonode/app/indigo/pb/store/segment.proto", fileDescriptorSegment)
}

var fileDescriptorSegment = []byte{
	// 392 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x41, 0x6a, 0xdb, 0x40,
	0x18, 0x85, 0x2b, 0x6a, 0x5a, 0x79, 0x5c, 0x28, 0xcc, 0xa2, 0x1d, 0xec, 0x22, 0x0b, 0xad, 0xdc,
	0x82, 0x67, 0x8a, 0x05, 0xed, 0xc2, 0x84, 0x80, 0xc0, 0x10, 0x82, 0x49, 0xc0, 0xde, 0x25, 0x8b,
	0x30, 0x92, 0x06, 0x69, 0xb0, 0x2c, 0x29, 0x9a, 0xb1, 0x49, 0x6e, 0x92, 0xb5, 0xae, 0xa1, 0x0b,
	0x64, 0x99, 0x23, 0x04, 0xe7, 0x22, 0xc1, 0x92, 0xc6, 0x41, 0xc4, 0xc1, 0xc9, 0x42, 0x82, 0x81,
	0xf7, 0xbd, 0xf7, 0xfe, 0x9f, 0x1f, 0x8c, 0x03, 0x2e, 0xc3, 0x95, 0x8b, 0xbd, 0x64, 0x49, 0x84,
	0xcc, 0xa8, 0x5c, 0x2d, 0x63, 0x42, 0x23, 0xee, 0x31, 0x42, 0xd3, 0x94, 0xf0, 0xd8, 0xe7, 0x41,
	0x42, 0x52, 0x97, 0x08, 0x99, 0x64, 0x8c, 0x08, 0x16, 0x2c, 0x59, 0x2c, 0x71, 0x9a, 0x25, 0x32,
	0x81, 0x7d, 0x45, 0xe0, 0x92, 0xc0, 0x34, 0x4d, 0x71, 0x45, 0xe0, 0x52, 0xde, 0xfd, 0xfb, 0xb6,
	0xbb, 0x17, 0x71, 0x12, 0x64, 0xa9, 0x47, 0xd8, 0x8d, 0xdc, 0x7e, 0x95, 0xa5, 0xf5, 0x1f, 0xb4,
	0xa6, 0x3c, 0x5e, 0x40, 0x02, 0x5a, 0x3e, 0x95, 0x14, 0x69, 0xa6, 0x36, 0xf8, 0xe6, 0xf4, 0xf2,
	0x02, 0xfd, 0x3c, 0x9d, 0x9f, 0x9f, 0x0d, 0x59, 0xec, 0x25, 0x3e, 0xf3, 0xcd, 0x88, 0xc7, 0x0b,
	0xd3, 0xbd, 0x95, 0x4c, 0xcc, 0x4a, 0xa1, 0xf5, 0x0f, 0xe8, 0x5b, 0xf0, 0x84, 0x8a, 0x10, 0xfe,
	0x69, 0xc0, 0x3f, 0xf2, 0x02, 0x41, 0x7b, 0x34, 0x2c, 0xd5, 0x15, 0x18, 0x52, 0x11, 0xd6, 0x1c,
	0x07, 0x40, 0x71, 0x4c, 0xc0, 0x4b, 0xd0, 0xd9, 0x0a, 0xae, 0xc2, 0xf2, 0x89, 0x34, 0xf3, 0xf3,
	0xa0, 0x33, 0xfa, 0x8d, 0x0f, 0xcc, 0x89, 0x95, 0x83, 0xf3, 0x3d, 0x2f, 0x50, 0x67, 0xaa, 0x22,
	0x98, 0x98, 0x81, 0x68, 0x67, 0x6e, 0x1d, 0x81, 0xaf, 0xf3, 0x6a, 0x7f, 0x70, 0xd4, 0x68, 0x68,
	0xe4, 0x05, 0xea, 0x36, 0xc6, 0xab, 0xf7, 0xdc, 0x98, 0xf0, 0x1a, 0xe8, 0x35, 0x2e, 0x20, 0x03,
	0x7a, 0x2d, 0x51, 0x25, 0x07, 0x07, 0x4b, 0xd6, 0xb0, 0xd3, 0xcf, 0x0b, 0xd4, 0xdb, 0x97, 0x26,
	0xea, 0xb8, 0x9d, 0xb5, 0x75, 0x0c, 0xf4, 0xc9, 0x9a, 0xfb, 0x2c, 0xf6, 0x18, 0xb4, 0x1b, 0x95,
	0x5f, 0x9b, 0xb0, 0x5a, 0xd8, 0xe8, 0xbc, 0x06, 0x6d, 0x65, 0x20, 0x20, 0x07, 0x6d, 0x25, 0x7a,
	0xff, 0x6a, 0x15, 0xee, 0x98, 0x79, 0x81, 0x7e, 0xed, 0x4d, 0x54, 0xbd, 0x5f, 0xdc, 0x9d, 0xc9,
	0xfd, 0xc6, 0xd0, 0x1e, 0x36, 0x86, 0xf6, 0xb8, 0x31, 0xb4, 0xbb, 0x27, 0xe3, 0xd3, 0x85, 0xfd,
	0xa1, 0x43, 0x1f, 0x97, 0x7f, 0xf7, 0x4b, 0x79, 0x94, 0xf6, 0x73, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xeb, 0x39, 0x03, 0x41, 0x26, 0x03, 0x00, 0x00,
}
