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

package validator_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

func TestValidateTx(t *testing.T) {
	type validateTxTestCase struct {
		name  string
		tx    func() *pb.Transaction
		state func() state.State
		err   error
	}

	testCases := []validateTxTestCase{{
		"empty-tx",
		func() *pb.Transaction { return nil },
		func() state.State { return nil },
		validator.ErrEmptyTx,
	}, {
		"empty-value",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 0, 0, 42)
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxValue,
	}, {
		"missing-to",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 0, 42)
			tx.To = nil
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxRecipient,
	}, {
		"missing-from",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 0, 42)
			tx.From = nil
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxSender,
	}, {
		"send-to-self",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 0, 42)
			tx.From = tx.To
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxRecipient,
	}, {
		"missing-signature",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 0, 42)
			tx.Signature = nil
			return tx
		},
		func() state.State { return nil },
		validator.ErrMissingTxSignature,
	}, {
		"invalid-signature",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 0, 42)
			tx.Signature.Signature[3] ^= 1
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxSignature,
	}, {
		"invalid-balance",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 0, 42)
			return tx
		},
		func() state.State { return testutil.NewSimpleState(t) },
		validator.ErrInsufficientBalance,
	}, {
		"invalid-balance-fee",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 40, 5, 3)
			return tx
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(testutil.TxSenderPID),
				&pb.Account{Balance: 41, Nonce: 1},
			)
			assert.NoError(t, err)
			return s
		},
		validator.ErrInsufficientBalance,
	}, {
		"invalid-nonce",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 3, 42)
			return tx
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(testutil.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 42},
			)
			assert.NoError(t, err)
			return s
		},
		validator.ErrInvalidTxNonce,
	}, {
		"valid-tx",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 3, 42)
			return tx
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(testutil.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 40},
			)
			assert.NoError(t, err)
			return s
		},
		nil,
	}}

	validator := validator.NewBalanceValidator(
		100,
		testutil.NewDummyPoW(&testutil.DummyEngine{}, 2, 5),
	)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTx(tt.tx(), tt.state())
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBlock(t *testing.T) {
	testReward := uint64(5)
	testMaxTxCount := uint32(3)

	type validateBlockTxTestCase struct {
		name  string
		block func() *pb.Block
		state func() state.State
		err   error
	}

	testCases := []validateBlockTxTestCase{{
		"empty-block",
		func() *pb.Block { return &pb.Block{} },
		func() state.State { return nil },
		nil,
	}, {
		"too-many-txs",
		func() *pb.Block {
			var txs []*pb.Transaction
			for i := uint32(0); i < testMaxTxCount+1; i++ {
				txs = append(txs, testutil.NewTransaction(t, 1, 1, uint64(i)))
			}

			return testutil.NewBlock(t, txs)
		},
		func() state.State { return nil },
		validator.ErrTooManyTxs,
	}, {
		"miner-reward-tx-count",
		// The miner reward shouldn't be taken into account
		// when validating the number of transactions per block.
		func() *pb.Block {
			var txs []*pb.Transaction
			for i := uint32(0); i < testMaxTxCount; i++ {
				txs = append(txs, testutil.NewTransaction(t, 1, 1, uint64(i+1)))
			}

			txs = append(txs, testutil.NewRewardTransaction(t, 1))
			return testutil.NewBlock(t, txs)
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(testutil.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 0},
			)
			assert.NoError(t, err)
			return s
		},
		nil,
	}, {
		"invalid-signature",
		func() *pb.Block {
			block := testutil.NewBlock(t, []*pb.Transaction{
				testutil.NewTransaction(t, 3, 1, 5),
				testutil.NewTransaction(t, 7, 2, 1),
			})

			block.Transactions[1].Signature.Signature[5] ^= block.Transactions[1].Signature.Signature[5]
			newRoot, err := coinutil.TransactionRoot(block.Transactions)
			assert.NoError(t, err, "coinutil.TransactionRoot()")

			block.Header.MerkleRoot = newRoot

			return block
		},
		func() state.State { return nil },
		validator.ErrInvalidTxSignature,
	}, {
		"invalid-balance",
		func() *pb.Block {
			return testutil.NewBlock(t, []*pb.Transaction{
				testutil.NewTransaction(t, 3, 2, 5),
				testutil.NewTransaction(t, 3, 1, 6),
			})
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(testutil.TxSenderPID),
				&pb.Account{Balance: 8, Nonce: 1},
			)
			assert.NoError(t, err)
			return s
		},
		validator.ErrInsufficientBalance,
	}, {
		"invalid-merkle-root",
		func() *pb.Block {
			block := testutil.NewBlock(t, []*pb.Transaction{
				testutil.NewTransaction(t, 3, 1, 5),
				testutil.NewTransaction(t, 7, 1, 6),
			})

			block.Header.MerkleRoot = []byte("does not look valid")

			return block
		},
		func() state.State { return nil },
		validator.ErrInvalidMerkleRoot,
	}, {"multiple-rewards",
		func() *pb.Block {
			return testutil.NewBlock(t, []*pb.Transaction{
				testutil.NewRewardTransaction(t, 3),
				testutil.NewRewardTransaction(t, 4),
			})
		},
		func() state.State { return nil },
		validator.ErrMultipleMinerRewards,
	}, {
		"invalid-reward",
		func() *pb.Block {
			return testutil.NewBlock(t, []*pb.Transaction{
				testutil.NewTransaction(t, 1, 5, 1),
				testutil.NewRewardTransaction(t, 5+testReward+1),
			})
		},
		func() state.State { return nil },
		validator.ErrInvalidMinerReward,
	}, {
		"valid-block",
		func() *pb.Block {
			return testutil.NewBlock(t, []*pb.Transaction{
				testutil.NewTransaction(t, 3, 1, 5),
				testutil.NewTransaction(t, 7, 1, 8),
				testutil.NewRewardTransaction(t, testReward+1+1),
			})
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(testutil.TxSenderPID),
				&pb.Account{Balance: 15, Nonce: 3},
			)
			assert.NoError(t, err)
			return s
		},
		nil,
	}}

	validator := validator.NewBalanceValidator(
		testMaxTxCount,
		testutil.NewDummyPoW(&testutil.DummyEngine{}, 1, testReward),
	)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBlock(tt.block(), tt.state())
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
