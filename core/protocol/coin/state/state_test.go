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
			err := s.UpdateAccount(alice, Account{Balance: 10, Nonce: 1})
			assert.NoError(t, err, "s.UpdateAccount()")
			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount()")
			assert.Equal(t, Account{Balance: 10, Nonce: 1}, v)

			assert.NoError(t, s.UpdateAccount(alice, Account{Balance: 20}), "s.UpdateAccount()")
			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount()")
			assert.Equal(t, Account{Balance: 20}, v)
		},
	}, {
		"get-inexisting",
		func(t *testing.T, s State) {
			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount()")
			assert.Equal(t, Account{}, v)
		},
	}, {
		"transfer",
		func(t *testing.T, s State) {
			err := s.UpdateAccount(alice, Account{Balance: 10})
			assert.NoError(t, err, "s.UpdateAccount(alice)")

			err = s.Transfer(alice, bob, 4, 3)
			assert.NoError(t, err)

			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{Balance: 6, Nonce: 3}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{Balance: 4}, v, "s.GetAccount(bob)")
		},
	}, {
		"transfer-too-big",
		func(t *testing.T, s State) {
			err := s.UpdateAccount(alice, Account{Balance: 10})
			assert.NoError(t, err, "s.UpdateAccount(alice)")

			err = s.Transfer(alice, bob, 14, 0)
			assert.EqualError(t, err, ErrAmountTooBig.Error())
		},
	}, {
		"write",
		func(t *testing.T, s State) {
			b := s.Batch()
			b.UpdateAccount(alice, Account{Balance: 10})
			b.UpdateAccount(bob, Account{Balance: 20})

			// Make sure state is not updated yet.
			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{}, v)
			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{}, v)

			assert.NoError(t, s.Write(b))

			// Make sure state was updated.
			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{Balance: 10}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{Balance: 20}, v, "s.GetAccount(bob)")
		},
	}, {
		"commit",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			require.NoError(t, err, "s.Transaction()")

			err = tx.UpdateAccount(alice, Account{Balance: 10})
			assert.NoError(t, err, "tx.UpdateAccount(alice)")

			err = tx.Transfer(alice, bob, 4, 2)
			assert.NoError(t, err, "tx.Transfer(alice, bob)")

			// Make sure state is not updated yet.
			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{}, v)
			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{}, v)

			assert.NoError(t, tx.Commit())

			// Make sure state was updated.
			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{Balance: 6, Nonce: 2}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{Balance: 4}, v, "s.GetAccount(bob)")
		},
	}, {
		"discard",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			require.NoError(t, err, "s.Transaction()")

			err = tx.UpdateAccount(alice, Account{Balance: 10})
			assert.NoError(t, err, "tx.UpdateAccount(alice)")

			err = tx.Transfer(alice, bob, 4, 2)
			assert.NoError(t, err, "tx.Transfer(alice, bob)")

			// Make sure state is not updated yet.
			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{}, v)
			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{}, v)

			tx.Discard()

			// Make sure state was not updated.
			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{}, v)
			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{}, v)
		},
	}, {
		"transaction-write",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			assert.NoError(t, err, "s.Transaction()")

			b := tx.Batch()
			b.UpdateAccount(alice, Account{Balance: 10, Nonce: 1})
			b.UpdateAccount(bob, Account{Balance: 20})

			// Make sure state is not updated yet.
			v, err := tx.GetAccount(alice)
			assert.NoError(t, err, "tx.GetAccount(alice)")
			assert.Equal(t, Account{}, v)
			v, err = tx.GetAccount(bob)
			assert.NoError(t, err, "tx.GetAccount(bob)")
			assert.Equal(t, Account{}, v)

			assert.NoError(t, tx.Write(b))

			assert.NoError(t, tx.Commit(), "tx.Commit()")

			// Make sure state was updated.
			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, Account{Balance: 10, Nonce: 1}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, Account{Balance: 20}, v, "s.GetAccount(bob)")
		},
	}, {
		"transaction-transfer-too-big",
		func(t *testing.T, s State) {
			tx, err := s.Transaction()
			assert.NoError(t, err, "s.Transaction()")

			tx.UpdateAccount(alice, Account{Balance: 10})
			assert.NoError(t, err, "tx.UpdateAccount(alice)")

			err = tx.Transfer(alice, bob, 14, 2)
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

func TestDecodeAccount_invalid(t *testing.T) {
	_, err := DecodeAccount([]byte("alice"))
	assert.EqualError(t, err, ErrInvalidAccount.Error())
}
