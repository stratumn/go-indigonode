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

package state

import (
	"testing"

	db "github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, State)
	}{{
		"set-get",
		func(t *testing.T, s State) {
			k := []byte("alice")

			assert.NoError(t, s.SetBalance(k, 10), "s.SetBalance()")
			v, err := s.GetBalance(k)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(10), v)

			assert.NoError(t, s.SetBalance(k, 20), "s.SetBalance()")
			v, err = s.GetBalance(k)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(20), v)
		},
	}, {
		"get-inexisting",
		func(t *testing.T, s State) {
			v, err := s.GetBalance([]byte("alice"))
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(0), v)
		},
	}, {
		"add-sub",
		func(t *testing.T, s State) {
			k := []byte("alice")

			v, err := s.AddBalance(k, 10)
			assert.NoError(t, err, "s.AddBalance()")
			assert.Equal(t, uint64(10), v, "s.AddBalance()")

			v, err = s.SubBalance(k, 4)
			assert.NoError(t, err, "s.SubBalance()")
			assert.Equal(t, uint64(6), v, "s.SubBalance()")

			v, err = s.GetBalance(k)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(6), v, "s.GetBalance()")
		},
	}, {
		"sub-too-big",
		func(t *testing.T, s State) {
			k := []byte("alice")

			_, err := s.AddBalance(k, 1)
			assert.NoError(t, err, "s.AddBalance()")

			_, err = s.SubBalance(k, 4)
			assert.EqualError(t, err, ErrAmountTooBig.Error())

			// Make sure balance hasn't changed.
			v, err := s.GetBalance(k)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(1), v, "s.GetBalance()")
		},
	}, {
		"write",
		func(t *testing.T, s State) {
			b := s.Batch()
			b.SetBalance([]byte("alice"), 10)
			b.SetBalance([]byte("bob"), 20)

			// Make sure state is not updated yet.
			v, err := s.GetBalance([]byte("alice"))
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.Equal(t, uint64(0), v)
			v, err = s.GetBalance([]byte("bob"))
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.Equal(t, uint64(0), v)

			assert.NoError(t, s.Write(b))

			// Make sure state was updated.
			v, err = s.GetBalance([]byte("alice"))
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.EqualValues(t, 10, v, "s.GetBalance(alice)")

			v, err = s.GetBalance([]byte("bob"))
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.EqualValues(t, 20, v, "s.GetBalance(bob)")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memdb, err := db.NewMemDB(nil)
			require.NoError(t, err, "db.NewMemDB()")
			defer memdb.Close()

			tt.run(t, NewState(memdb, []byte("test-")))
		})
	}
}

// Whitebox test to make sure it works.
func TestPrefix(t *testing.T) {
	k := prefixKey([]byte("hello-"), []byte("world"))
	assert.Equal(t, "hello-world", string(k))
}
