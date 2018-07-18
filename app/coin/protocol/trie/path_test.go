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
	"github.com/stretchr/testify/require"
)

// newPath creates a new path from a buffer. If odd is true, then the number of
// nibbles is odd.
func newPath(buf []byte, odd bool) path {
	return path(newNibs(buf, odd))
}

func TestPath_Encoding(t *testing.T) {
	tests := []struct {
		name         string
		path         path
		marshalErr   error
		unmarshalErr error
	}{{
		"empty",
		newPath(nil, false),
		nil,
		nil,
	}, {
		"0-to-2",
		newPath([]byte{0x01, 0x20}, true),
		nil,
		nil,
	}, {
		"0-to-15",
		newPath([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, false),
		nil,
		nil,
	}, {
		"15-to-0",
		newPath([]byte{0xfe, 0xdc, 0xba, 0x98, 0x76, 0x54, 0x32, 0x10}, false),
		nil,
		nil,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := tt.path.MarshalBinary()

			if tt.marshalErr != nil {
				assert.EqualError(t, err, tt.marshalErr.Error(), "tt.Path.MarshalBinary()")
				return
			}

			require.NoError(t, err, "tt.path.MarshalBinary()")

			path, read, err := unmarshalPath(buf)

			if tt.unmarshalErr != nil {
				assert.EqualError(t, err, tt.unmarshalErr.Error(), "UnmarshalPath()")
				return
			}

			require.NoError(t, err, "UnmarshalPath()")
			assert.Equal(t, len(buf), read, "UnmarshalPath().read")
			assert.Equal(t, tt.path, path, "UnmarshalPath().path")
		})
	}
}
