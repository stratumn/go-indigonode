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
	alice := []byte("alice")
	bob := []byte("bob")

	tests := []struct {
		name string
		run  func(*testing.T, State)
	}{{
		"set-get",
		func(t *testing.T, s State) {
			assert.NoError(t, s.SetBalance(alice, 10), "s.SetBalance()")
			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(10), v)

			assert.NoError(t, s.SetBalance(alice, 20), "s.SetBalance()")
			v, err = s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(20), v)
		},
	}, {
		"get-inexisting",
		func(t *testing.T, s State) {
			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(0), v)
		},
	}, {
		"add-sub",
		func(t *testing.T, s State) {
			v, err := s.AddBalance(alice, 10)
			assert.NoError(t, err, "s.AddBalance()")
			assert.Equal(t, uint64(10), v, "s.AddBalance()")

			v, err = s.SubBalance(alice, 4)
			assert.NoError(t, err, "s.SubBalance()")
			assert.Equal(t, uint64(6), v, "s.SubBalance()")

			v, err = s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(6), v, "s.GetBalance()")
		},
	}, {
		"sub-too-big",
		func(t *testing.T, s State) {
			_, err := s.AddBalance(alice, 1)
			assert.NoError(t, err, "s.AddBalance()")

			_, err = s.SubBalance(alice, 4)
			assert.EqualError(t, err, ErrAmountTooBig.Error())

			// Make sure balance hasn't changed.
			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance()")
			assert.Equal(t, uint64(1), v, "s.GetBalance()")
		},
	}, {
		"transfer",
		func(t *testing.T, s State) {
			_, err := s.AddBalance(alice, 10)
			assert.NoError(t, err, "s.AddBalance(alice)")

			err = s.Transfer(alice, bob, 4)
			assert.NoError(t, err)

			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.Equal(t, uint64(6), v, "s.GetBalance(alice)")

			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.Equal(t, uint64(4), v, "s.GetBalance(bob)")
		},
	}, {
		"transfer-too-big",
		func(t *testing.T, s State) {
			_, err := s.AddBalance(alice, 10)
			assert.NoError(t, err, "s.AddBalance(alice)")

			err = s.Transfer(alice, bob, 14)
			assert.EqualError(t, err, ErrAmountTooBig.Error())
		},
	}, {
		"write",
		func(t *testing.T, s State) {
			b := s.Batch()
			b.SetBalance(alice, 10)
			b.SetBalance(bob, 20)

			// Make sure state is not updated yet.
			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.Equal(t, uint64(0), v)
			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.Equal(t, uint64(0), v)

			assert.NoError(t, s.Write(b))

			// Make sure state was updated.
			v, err = s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.EqualValues(t, 10, v, "s.GetBalance(alice)")

			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.EqualValues(t, 20, v, "s.GetBalance(bob)")
		},
	}, {
		"commit",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			require.NoError(t, err, "s.Transaction()")

			err = tx.SetBalance(alice, 10)
			assert.NoError(t, err, "tx.SetBalance(alice)")

			err = tx.Transfer(alice, bob, 4)
			assert.NoError(t, err, "tx.Transfer(alice, bob)")

			// Make sure state is not updated yet.
			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.Equal(t, uint64(0), v)
			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.Equal(t, uint64(0), v)

			assert.NoError(t, tx.Commit())

			// Make sure state was updated.
			v, err = s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.EqualValues(t, 6, v, "s.GetBalance(alice)")

			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.EqualValues(t, 4, v, "s.GetBalance(bob)")
		},
	}, {
		"discard",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			require.NoError(t, err, "s.Transaction()")

			err = tx.SetBalance(alice, 10)
			assert.NoError(t, err, "tx.SetBalance(alice)")

			err = tx.Transfer(alice, bob, 4)
			assert.NoError(t, err, "tx.Transfer(alice, bob)")

			// Make sure state is not updated yet.
			v, err := s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.Equal(t, uint64(0), v)
			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.Equal(t, uint64(0), v)

			tx.Discard()

			// Make sure state was not updated.
			v, err = s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.Equal(t, uint64(0), v)
			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.Equal(t, uint64(0), v)
		},
	}, {
		"transaction-write",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			assert.NoError(t, err, "s.Transaction()")

			b := tx.Batch()
			b.SetBalance(alice, 10)
			b.SetBalance(bob, 20)

			// Make sure state is not updated yet.
			v, err := tx.GetBalance(alice)
			assert.NoError(t, err, "tx.GetBalance(alice)")
			assert.Equal(t, uint64(0), v)
			v, err = tx.GetBalance(bob)
			assert.NoError(t, err, "tx.GetBalance(bob)")
			assert.Equal(t, uint64(0), v)

			assert.NoError(t, tx.Write(b))

			assert.NoError(t, tx.Commit(), "tx.Commit()")

			// Make sure state was updated.
			v, err = s.GetBalance(alice)
			assert.NoError(t, err, "s.GetBalance(alice)")
			assert.EqualValues(t, 10, v, "s.GetBalance(alice)")

			v, err = s.GetBalance(bob)
			assert.NoError(t, err, "s.GetBalance(bob)")
			assert.EqualValues(t, 20, v, "s.GetBalance(bob)")
		},
	}, {
		"transaction-transfer-too-big",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			assert.NoError(t, err, "s.Transaction()")

			_, err = tx.AddBalance(alice, 10)
			assert.NoError(t, err, "tx.AddBalance(alice)")

			err = tx.Transfer(alice, bob, 14)
			assert.EqualError(t, err, ErrAmountTooBig.Error())
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
