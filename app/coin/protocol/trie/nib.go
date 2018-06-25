// Copyright © 2017-2018 Stratumn SAS
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
	"encoding/hex"
)

// nibs deals with nibbles (four bits) in a buffer.
type nibs struct {
	buf []byte
	odd bool
}

// newNibs creates a Nibs from a buffer. If odd is true, the last nibble of the
// buffer is ignored.
func newNibs(buf []byte, odd bool) nibs {
	b := make([]byte, len(buf))
	copy(b, buf)

	return newNibsWithoutCopy(b, odd)
}

// newNibsWithoutCopy creates a Nibs from a buffer without copying the buffer.
func newNibsWithoutCopy(buf []byte, odd bool) nibs {
	if odd {
		buf[len(buf)-1] &= 0xF0
	}

	return nibs{buf: buf, odd: odd}
}

// newNibsFromNibs creates a Nibs from a slice of nibbles.
func newNibsFromNibs(nbs ...uint8) nibs {
	buf := make([]byte, (len(nbs)+1)/2)
	n := newNibsWithoutCopy(buf, len(nbs)%2 == 1)

	for i, v := range nbs {
		n.Put(i, v)
	}

	return n
}

// String returns a string representation of the nibbles.
func (n nibs) String() string {
	return hex.EncodeToString(n.buf)[:n.Len()]
}

// Len returns the number of nibbles in the buffer.
func (n nibs) Len() int {
	l := len(n.buf)
	if l < 1 {
		return 0
	}

	if n.odd {
		return l*2 - 1
	}

	return l * 2
}

// ByteLen returns the number of bytes in the buffer.
func (n nibs) ByteLen() int {
	return len(n.buf)
}

// Odd returns whether the number of nibbles is odd.
func (n nibs) Odd() bool {
	return n.odd
}

// At returns the nth nibble.
func (n nibs) At(index int) uint8 {
	b := n.buf[index/2]

	if index%2 == 0 {
		return b >> 4
	}

	return b & 0x0F
}

// Put sets the nibble at the specified nibble index.
func (n nibs) Put(index int, nib uint8) {
	i := index / 2

	if index%2 == 0 {
		n.buf[i] &= 0x0F
		n.buf[i] |= nib << 4
		return
	}

	n.buf[i] &= 0xF0
	n.buf[i] |= nib & 0x0F
}

// Append appends nibbles.
func (n nibs) Append(nbs ...nibs) nibs {
	l := n.Len()

	for _, nib := range nbs {
		l += nib.Len()
	}

	appended := newNibsWithoutCopy(make([]byte, (l+1)/2), l%2 == 1)
	n.CopyToBuf(appended.buf)
	i := n.Len()

	for _, nib := range nbs {
		for j := 0; j < nib.Len(); j++ {
			appended.Put(i, nib.At(j))
			i++
		}
	}

	return appended
}

// Substr returns a substring of the nibbles. The indexes are given in number
// of nibbles.
func (n nibs) Substr(from, to int) nibs {
	l := to - from
	sub := newNibsWithoutCopy(make([]byte, (l+1)/2), l%2 == 1)

	for i, j := 0, from; j < to; i, j = i+1, j+1 {
		sub.Put(i, n.At(j))
	}

	return sub
}

// CopyToBuf copies the nibbles to a buffer.
func (n nibs) CopyToBuf(buf []byte) {
	if n.odd && len(n.buf) > 0 {
		copy(buf, n.buf[:len(n.buf)-1])
		buf[len(n.buf)-1] &= 0x0F
		buf[len(n.buf)-1] |= n.buf[len(n.buf)-1] & 0xF0
		return
	}

	copy(buf, n.buf)
}

// Expand returns a slice containing the individual nibbles.
func (n nibs) Expand() []uint8 {
	s := make([]uint8, n.Len())

	for i := 0; i < n.Len(); i++ {
		s[i] = n.At(i)
	}

	return s
}
