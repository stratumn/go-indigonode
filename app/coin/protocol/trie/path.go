// Copyright © 2017-2018 Stratumn SAS
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package trie

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

var (
	// ErrInvalidPathLen is returned when an encoded path has an invalid
	// number of bytes.
	ErrInvalidPathLen = errors.New("the encoded path has an invalid number of bytes")

	// ErrBufferToShort is returned when a given buffer is too short.
	ErrBufferToShort = errors.New("the buffer is too short")
)

// path represents the path of a node down the tree. Each nibble represents an
// index taken (0-15).
//
// A path has a compact binary encoding. Since a branch has sixteen children,
// two indices can be encoded into a single byte.
//
// path Encoding
//
//	10100110           [1010 ...]
//	depth (uvarint32)  indices  (one nibble for each depth)
type path nibs

// unmarshalPath unmarshals a path. It returns the path and the number of bytes
// read if no error occured.
func unmarshalPath(buf []byte) (path, int, error) {
	depth, read := binary.Uvarint(buf)

	if read <= 0 || len(buf) < read+int(depth+1)/2 {
		return path{}, 0, errors.WithStack(ErrInvalidPathLen)
	}

	buf = buf[read : read+(int(depth)+1)/2]
	nibs := newNibs(buf, depth%2 == 1)

	return path(nibs), read + nibs.ByteLen(), nil
}

// MarshalBinary marshals the path.
func (p path) MarshalBinary() ([]byte, error) {
	buf := make([]byte, binary.MaxVarintLen32+nibs(p).ByteLen())

	written, err := p.MarshalInto(buf)
	if err != nil {
		return nil, err
	}

	return buf[:written], nil
}

// MarshalInto marshals the path into an existing buffer. It returns the number
// of bytes written.
func (p path) MarshalInto(buf []byte) (int, error) {
	if len(buf) < binary.MaxVarintLen32 {
		return 0, errors.WithStack(ErrBufferToShort)
	}

	written := binary.PutUvarint(buf, uint64(nibs(p).Len()))

	if len(buf) < written+nibs(p).ByteLen() {
		return 0, errors.WithStack(ErrBufferToShort)
	}

	nibs(p).CopyToBuf(buf[written:])

	return written + nibs(p).ByteLen(), nil
}
