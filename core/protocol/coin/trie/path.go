// Copyright © 2017-2018  Stratumn SAS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
