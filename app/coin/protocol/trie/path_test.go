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
