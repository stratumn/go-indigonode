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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNibs_String(t *testing.T) {
	tests := []struct {
		name     string
		n        nibs
		expected string
	}{{
		"even",
		newNibsFromNibs(1, 2, 3, 4),
		"1234",
	}, {
		"odd",
		newNibs([]byte{0xab, 0xcd}, true),
		"abc",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.String())
		})
	}
}

func TestNibs_Len(t *testing.T) {
	tests := []struct {
		name     string
		n        nibs
		expected int
	}{{
		"nil",
		newNibs(nil, false),
		0,
	}, {
		"even",
		newNibs([]byte{0x12, 0x34}, false),
		4,
	}, {
		"odd",
		newNibs([]byte{0xab, 0xcd}, true),
		3,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.Len())
		})
	}
}

func TestNibs_ByteLen(t *testing.T) {
	tests := []struct {
		name     string
		n        nibs
		expected int
	}{{
		"even",
		newNibs([]byte{0x12, 0x34}, false),
		2,
	}, {
		"odd",
		newNibs([]byte{0xab, 0xcd}, true),
		2,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.ByteLen())
		})
	}
}

func TestNibs_At(t *testing.T) {
	tests := []struct {
		name     string
		n        nibs
		index    int
		expected uint8
	}{{
		"zero",
		newNibs([]byte{0x12, 0xef}, false),
		0,
		1,
	}, {
		"one",
		newNibs([]byte{0x12, 0xef}, false),
		1,
		2,
	}, {
		"two",
		newNibs([]byte{0x12, 0xef}, false),
		2,
		14,
	}, {
		"three",
		newNibs([]byte{0x12, 0xef}, false),
		3,
		15,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.At(tt.index))
		})
	}
}

func TestNibs_Append(t *testing.T) {
	tests := []struct {
		name     string
		n        []nibs
		expected string
	}{{
		"even-even",
		[]nibs{
			newNibs([]byte{0x12, 0x34}, false),
			newNibs([]byte{0xab, 0xcd}, false),
		},
		"1234abcd",
	}, {
		"odd-odd",
		[]nibs{
			newNibs([]byte{0x12, 0x34}, true),
			newNibs([]byte{0xab, 0xcd}, true),
		},
		"123abc",
	}, {
		"even-odd-even",
		[]nibs{
			newNibs([]byte{0x12, 0x34}, false),
			newNibs([]byte{0x56, 0x78}, true),
			newNibs([]byte{0xab, 0xcd}, false),
		},
		"1234567abcd",
	}, {
		"odd-even-odd",
		[]nibs{
			newNibs([]byte{0x12, 0x34}, true),
			newNibs([]byte{0x56, 0x78}, false),
			newNibs([]byte{0xab, 0xcd}, true),
		},
		"1235678abc",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n[0].Append(tt.n[1:]...).String())
		})
	}
}

func TestNibs_Substr(t *testing.T) {
	tests := []struct {
		name     string
		n        nibs
		from     int
		to       int
		expected string
	}{{
		"even-to-even",
		newNibs([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, false),
		4,
		12,
		"456789ab",
	}, {
		"odd-to-odd",
		newNibs([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, false),
		7,
		11,
		"789a",
	}, {
		"even-to-odd",
		newNibs([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, false),
		6,
		11,
		"6789a",
	}, {
		"odd-to-even",
		newNibs([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, false),
		7,
		12,
		"789ab",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.Substr(tt.from, tt.to).String())
		})
	}
}

func TestNibs_Expand(t *testing.T) {
	tests := []struct {
		name     string
		n        nibs
		expected []uint8
	}{{
		"even",
		newNibs([]byte{0x12, 0x34}, false),
		[]uint8{1, 2, 3, 4},
	}, {
		"odd",
		newNibs([]byte{0xab, 0xcd}, true),
		[]uint8{0xa, 0xb, 0xc},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.n.Expand())
		})
	}
}
