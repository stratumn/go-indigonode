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

package state

import (
	"testing"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/testutil/blocktest"
	txtest "github.com/stratumn/go-indigonode/app/coin/protocol/testutil/transaction"
	"github.com/stratumn/go-indigonode/core/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	computeTxKeys := func(txs []*pb.Transaction, blk *pb.Block) []*TxKey {
		h, err := coinutil.HashHeader(blk.Header)
		assert.NoError(t, err, "coinutil.HashHeader()")

		txKeys := make([]*TxKey, len(txs))
		for i := range txs {
			txKeys[i] = &TxKey{TxIdx: uint64(i), BlkHash: h}
		}
		return txKeys
	}

	_, _, pid, err := txtest.NewKeyPair()
	require.NoError(t, err)
	alice := []byte(pid)

	_, _, pid, err = txtest.NewKeyPair()
	require.NoError(t, err)
	bob := []byte(pid)

	_, _, pid, err = txtest.NewKeyPair()
	require.NoError(t, err)
	charlie := []byte(pid)

	_, _, pid, err = txtest.NewKeyPair()
	require.NoError(t, err)
	donald := []byte(pid)

	tests := []struct {
		name string
		run  func(*testing.T, State)
	}{{
		"set-get",
		func(t *testing.T, s State) {
			err := s.UpdateAccount(alice, &pb.Account{Balance: 10, Nonce: 1})
			assert.NoError(t, err, "s.UpdateAccount()")
			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount()")
			assert.Equal(t, &pb.Account{Balance: 10, Nonce: 1}, v)

			err = s.UpdateAccount(alice, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount()")
			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount()")
			assert.Equal(t, &pb.Account{Balance: 20}, v)
		},
	}, {
		"get-inexisting",
		func(t *testing.T, s State) {
			v, err := s.GetAccount(donald)
			assert.NoError(t, err, "s.GetAccount()")
			assert.Equal(t, &pb.Account{}, v)
		},
	}, {
		"process-transactions",
		func(t *testing.T, s State) {
			err := s.UpdateAccount(alice, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount()")

			err = s.UpdateAccount(bob, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount()")

			err = s.UpdateAccount(charlie, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount()")

			txs := []*pb.Transaction{{
				From:  alice,
				To:    bob,
				Value: 10,
				Fee:   2,
				Nonce: 1,
			}, {
				From:  bob,
				To:    charlie,
				Value: 5,
				Fee:   1,
				Nonce: 1,
			}, {
				From:  charlie,
				To:    alice,
				Value: 2,
				Fee:   1,
				Nonce: 1,
			}}
			blk := blocktest.NewBlock(t, txs)

			err = s.ProcessBlock(blk)
			assert.NoError(t, err)

			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, &pb.Account{Balance: 20 - 10 - 2 + 2, Nonce: 1}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, &pb.Account{Balance: 20 + 10 - 5 - 1, Nonce: 1}, v, "s.GetAccount(bob)")

			v, err = s.GetAccount(charlie)
			assert.NoError(t, err, "s.GetAccount(charlie)")
			assert.Equal(t, &pb.Account{Balance: 20 + 5 - 2 - 1, Nonce: 1}, v, "s.GetAccount(charlie)")
		},
	}, {
		"process-reward-transaction",
		func(t *testing.T, s State) {
			err := s.UpdateAccount(alice, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount()")

			txs := []*pb.Transaction{{
				From:  alice,
				To:    bob,
				Value: 10,
				Nonce: 1,
			}, {
				To:    alice,
				Value: 5,
			}}
			blk := blocktest.NewBlock(t, txs)

			err = s.ProcessBlock(blk)
			assert.NoError(t, err)

			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, &pb.Account{Balance: 20 - 10 + 5, Nonce: 1}, v, "s.GetAccount(alice)")

			// Verify that we didn't wrongfully take into account the miner reward
			// that has a nil sender address.
			v, err = s.GetAccount(nil)
			assert.NoError(t, err, "s.GetAccount(nil)")
			assert.Equal(t, &pb.Account{}, v, "s.GetAccount(nil)")
		},
	}, {
		"rollback-transactions",
		func(t *testing.T, s State) {
			// Process two states.

			txs1 := []*pb.Transaction{{
				From:  alice,
				To:    bob,
				Value: 10,
				Fee:   1,
				Nonce: 1,
			}, {
				From:  bob,
				To:    charlie,
				Value: 5,
				Fee:   2,
				Nonce: 2,
			}, {
				From:  charlie,
				To:    alice,
				Value: 2,
				Fee:   3,
				Nonce: 3,
			}}
			blk1 := blocktest.NewBlock(t, txs1)

			txs2 := []*pb.Transaction{{
				From:  bob,
				To:    charlie,
				Value: 2,
				Fee:   1,
				Nonce: 3,
			}, {
				From:  alice,
				To:    bob,
				Value: 3,
				Fee:   2,
				Nonce: 4,
			}}
			blk2 := blocktest.NewBlock(t, txs2)
			blk2.Header.BlockNumber = blk1.BlockNumber() + 1

			mr0a, err := s.MerkleRoot()
			assert.NoError(t, err, "s.MerkleRoot(blk0a)")

			err = s.ProcessBlock(blk1)
			assert.NoError(t, err, "s.ProcessBlock(blk1)")

			mr1a, err := s.MerkleRoot()
			assert.NoError(t, err, "s.MerkleRoot(blk1a)")

			err = s.ProcessBlock(blk2)
			assert.NoError(t, err, "s.ProcessBlock(blk2)")

			// Rollback blk2.
			err = s.RollbackBlock(blk2)
			assert.NoError(t, err, "s.RollbackBlock(blk2)")

			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, &pb.Account{Balance: 20 - 10 - 1 + 2, Nonce: 1}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, &pb.Account{Balance: 20 + 10 - 5 - 2, Nonce: 2}, v, "s.GetAccount(bob)")

			v, err = s.GetAccount(charlie)
			assert.NoError(t, err, "s.GetAccount(charlie)")
			assert.Equal(t, &pb.Account{Balance: 20 + 5 - 2 - 3, Nonce: 3}, v, "s.GetAccount(charlie)")

			mr1b, err := s.MerkleRoot()
			assert.NoError(t, err, "s.MerkleRoot(blk1b)")
			assert.Equal(t, mr1a, mr1b, "s.MerkleRoot(blk1b)")

			// Rollback blk1.
			err = s.RollbackBlock(blk1)
			assert.NoError(t, err, "s.RollbackBlock(blk1)")

			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, &pb.Account{Balance: 20}, v, "s.GetAccount(alice)")

			v, err = s.GetAccount(bob)
			assert.NoError(t, err, "s.GetAccount(bob)")
			assert.Equal(t, &pb.Account{Balance: 20}, v, "s.GetAccount(bob)")

			v, err = s.GetAccount(charlie)
			assert.NoError(t, err, "s.GetAccount(charlie)")
			assert.Equal(t, &pb.Account{Balance: 20}, v, "s.GetAccount(charlie)")

			mr0b, err := s.MerkleRoot()
			assert.NoError(t, err, "s.MerkleRoot(blk0b)")
			assert.Equal(t, mr0a, mr0b, "s.MerkleRoot(blk0b)")

			// Verify that we didn't wrongfully take into account the miner reward
			// that has a nil sender address.
			v, err = s.GetAccount(nil)
			assert.NoError(t, err, "s.GetAccount(nil)")
			assert.Equal(t, &pb.Account{}, v, "s.GetAccount(nil)")
		},
	}, {
		"process-transaction-nonces",
		func(t *testing.T, s State) {
			err := s.UpdateAccount(alice, &pb.Account{Balance: 42000})
			assert.NoError(t, err, "s.UpdateAccount()")

			txs1 := []*pb.Transaction{{
				From:  alice,
				To:    bob,
				Value: 3,
				Fee:   1,
				Nonce: 2,
			}, {
				From:  alice,
				To:    charlie,
				Value: 5,
				Fee:   1,
				Nonce: 3,
			}, {
				From:  alice,
				To:    bob,
				Value: 2,
				Fee:   1,
				Nonce: 1,
			}}
			blk1 := blocktest.NewBlock(t, txs1)

			txs2 := []*pb.Transaction{{
				From:  alice,
				To:    charlie,
				Value: 2,
				Fee:   1,
				Nonce: 6,
			}, {
				From:  alice,
				To:    bob,
				Value: 3,
				Fee:   1,
				Nonce: 5,
			}}
			blk2 := blocktest.NewBlock(t, txs2)

			err = s.ProcessBlock(blk1)
			assert.NoError(t, err, "s.ProcessBlock(blk1)")

			v, err := s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, uint64(3), v.Nonce, "s.GetAccount(alice).Nonce")

			err = s.ProcessBlock(blk2)
			assert.NoError(t, err, "s.ProcessBlock(blk2)")

			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, uint64(6), v.Nonce, "s.GetAccount(alice).Nonce")

			// Rollback blk2.
			err = s.RollbackBlock(blk2)
			assert.NoError(t, err, "s.RollbackBlock(blk2)")

			v, err = s.GetAccount(alice)
			assert.NoError(t, err, "s.GetAccount(alice)")
			assert.Equal(t, uint64(3), v.Nonce, "s.GetAccount(alice).Nonce")
		},
	}, {
		"stores-transaction-hashes",
		func(t *testing.T, s State) {
			txs1 := []*pb.Transaction{{
				From:  alice,
				To:    bob,
				Value: 3,
				Fee:   1,
				Nonce: 2,
			}, {
				From:  alice,
				To:    charlie,
				Value: 5,
				Fee:   1,
				Nonce: 3,
			}, {
				From:  alice,
				To:    bob,
				Value: 2,
				Fee:   1,
				Nonce: 1,
			}}
			blk1 := blocktest.NewBlock(t, txs1)
			txh1 := computeTxKeys(txs1, blk1)

			txs2 := []*pb.Transaction{{
				From:  alice,
				To:    charlie,
				Value: 2,
				Fee:   1,
				Nonce: 6,
			}, {
				From:  alice,
				To:    bob,
				Value: 3,
				Fee:   1,
				Nonce: 5,
			}}
			blk2 := blocktest.NewBlock(t, txs2)
			blk2.GetHeader().BlockNumber = blk1.BlockNumber() + 1
			txh2 := computeTxKeys(txs2, blk2)

			v, err := s.GetAccountTxKeys(alice)
			assert.NoError(t, err, "s.GetAccountTxHashes(alice)")
			assert.ElementsMatch(t, []*TxKey{}, v, "s.GetAccountTxHashes(alice)")

			err = s.ProcessBlock(blk1)
			assert.NoError(t, err, "s.ProcessBlock(blk1)")

			v, err = s.GetAccountTxKeys(alice)
			assert.NoError(t, err, "s.GetAccountTxHashes(alice)")
			assert.ElementsMatch(t, txh1, v, "s.GetAccountTxHashes(alice)")

			v, err = s.GetAccountTxKeys(bob)
			assert.NoError(t, err, "s.GetAccountTxHashes(bob)")
			assert.ElementsMatch(t, []*TxKey{txh1[0], txh1[2]}, v, "s.GetAccountTxHashes(bob)")

			v, err = s.GetAccountTxKeys(charlie)
			assert.NoError(t, err, "s.GetAccountTxHashes(charlie)")
			assert.ElementsMatch(t, []*TxKey{txh1[1]}, v, "s.GetAccountTxHashes(charlie)")

			err = s.ProcessBlock(blk2)
			assert.NoError(t, err, "s.ProcessBlock(blk2)")

			v, err = s.GetAccountTxKeys(alice)
			assert.NoError(t, err, "s.GetAccountTxHashes(alice)")
			assert.ElementsMatch(t, append(txh1, txh2...), v, "s.GetAccountTxHashes(alice)")

			// Rollback.
			err = s.RollbackBlock(blk2)
			assert.NoError(t, err, "s.RollbackBlock(blk2)")

			v, err = s.GetAccountTxKeys(alice)
			assert.NoError(t, err, "s.GetAccountTxHashes(alice)")
			assert.ElementsMatch(t, txh1, v, "s.GetAccountTxHashes(alice)")
		},
	}, {
		"orders-transaction-hashes",
		func(t *testing.T, s State) {
			txs1 := []*pb.Transaction{{
				From:  alice,
				To:    bob,
				Value: 3,
				Fee:   1,
				Nonce: 2,
			}}
			blk1 := blocktest.NewBlock(t, txs1)
			txh1 := computeTxKeys(txs1, blk1)[0]

			txs2 := []*pb.Transaction{{
				From:  alice,
				To:    charlie,
				Value: 2,
				Fee:   1,
				Nonce: 6,
			}}
			blk2 := blocktest.NewBlock(t, txs2)
			blk2.GetHeader().BlockNumber = blk1.BlockNumber() + 1
			txh2 := computeTxKeys(txs2, blk2)[0]

			err := s.ProcessBlock(blk1)
			assert.NoError(t, err, "s.ProcessBlock(blk1)")

			err = s.ProcessBlock(blk2)
			assert.NoError(t, err, "s.ProcessBlock(blk2)")

			v, err := s.GetAccountTxKeys(alice)
			assert.NoError(t, err, "s.GetAccountTxHashes(alice)")
			assert.Equal(t, []*TxKey{txh1, txh2}, v, "s.GetAccountTxHashes(alice)")
		},
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memdb, err := db.NewMemDB(nil)
			require.NoError(t, err, "db.NewMemDB()")
			defer memdb.Close()

			s := NewState(memdb, OptPrefix([]byte("test-")))

			err = s.UpdateAccount(alice, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount(alice)")

			err = s.UpdateAccount(bob, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount(alice)")

			err = s.UpdateAccount(charlie, &pb.Account{Balance: 20})
			assert.NoError(t, err, "s.UpdateAccount(alice)")

			tt.run(t, s)
		})
	}
}
