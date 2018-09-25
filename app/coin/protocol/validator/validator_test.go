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

package validator_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/chain/mockchain"
	"github.com/stratumn/go-node/app/coin/protocol/coinutil"
	"github.com/stratumn/go-node/app/coin/protocol/engine"
	"github.com/stratumn/go-node/app/coin/protocol/engine/mockengine"
	"github.com/stratumn/go-node/app/coin/protocol/state"
	"github.com/stratumn/go-node/app/coin/protocol/testutil"
	"github.com/stratumn/go-node/app/coin/protocol/testutil/blocktest"
	txtest "github.com/stratumn/go-node/app/coin/protocol/testutil/transaction"
	"github.com/stratumn/go-node/app/coin/protocol/validator"
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
			tx := txtest.NewTransaction(t, 0, 0, 42)
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxValue,
	}, {
		"missing-to",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			tx.To = nil
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxRecipient,
	}, {
		"missing-from",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			tx.From = nil
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxSender,
	}, {
		"send-to-self",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			tx.From = tx.To
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxRecipient,
	}, {
		"missing-signature",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			tx.Signature = nil
			return tx
		},
		func() state.State { return nil },
		validator.ErrMissingTxSignature,
	}, {
		"invalid-signature",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			tx.Signature.Signature[3] ^= 1
			return tx
		},
		func() state.State { return nil },
		validator.ErrInvalidTxSignature,
	}, {
		"invalid-balance",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			return tx
		},
		func() state.State { return testutil.NewSimpleState(t) },
		state.ErrInsufficientBalance,
	}, {
		"invalid-balance-fee",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 40, 5, 3)
			return tx
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 41, Nonce: 1},
			)
			assert.NoError(t, err)
			return s
		},
		state.ErrInsufficientBalance,
	}, {
		"invalid-nonce",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 3, 42)
			return tx
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 42},
			)
			assert.NoError(t, err)
			return s
		},
		state.ErrInvalidTxNonce,
	}, {
		"valid-tx",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 3, 42)
			return tx
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
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
		"genesis-block",
		func() *pb.Block { return &pb.Block{Header: &pb.Header{BlockNumber: 0}} },
		func() state.State { return nil },
		validator.ErrBlockHeightZero,
	}, {
		"empty-block",
		func() *pb.Block { return &pb.Block{Header: &pb.Header{BlockNumber: 42}} },
		func() state.State { return nil },
		nil,
	}, {
		"too-many-txs",
		func() *pb.Block {
			var txs []*pb.Transaction
			for i := uint32(0); i < testMaxTxCount+1; i++ {
				txs = append(txs, txtest.NewTransaction(t, 1, 1, uint64(i)))
			}

			return blocktest.NewBlock(t, txs)
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
				txs = append(txs, txtest.NewTransaction(t, 1, 1, uint64(i+1)))
			}

			txs = append(txs, txtest.NewRewardTransaction(t, 1))
			return blocktest.NewBlock(t, txs)
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 0},
			)
			assert.NoError(t, err)
			return s
		},
		nil,
	}, {
		"invalid-signature",
		func() *pb.Block {
			block := blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 3, 1, 5),
				txtest.NewTransaction(t, 7, 2, 1),
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
			return blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 3, 2, 5),
				txtest.NewTransaction(t, 3, 1, 6),
			})
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 8, Nonce: 1},
			)
			assert.NoError(t, err)
			return s
		},
		state.ErrInsufficientBalance,
	}, {
		"invalid-merkle-root",
		func() *pb.Block {
			block := blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 3, 1, 5),
				txtest.NewTransaction(t, 7, 1, 6),
			})

			block.Header.MerkleRoot = []byte("does not look valid")

			return block
		},
		func() state.State { return nil },
		validator.ErrInvalidMerkleRoot,
	}, {"multiple-rewards",
		func() *pb.Block {
			return blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewRewardTransaction(t, 3),
				txtest.NewRewardTransaction(t, 4),
			})
		},
		func() state.State { return nil },
		validator.ErrMultipleMinerRewards,
	}, {
		"invalid-reward",
		func() *pb.Block {
			return blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 1, 5, 1),
				txtest.NewRewardTransaction(t, 5+testReward+1),
			})
		},
		func() state.State { return nil },
		validator.ErrInvalidMinerReward,
	}, {
		"valid-block",
		func() *pb.Block {
			return blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 3, 1, 5),
				txtest.NewTransaction(t, 7, 1, 8),
				txtest.NewRewardTransaction(t, testReward+1+1),
			})
		},
		func() state.State {
			s := testutil.NewSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
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

func TestGossipValidator(t *testing.T) {
	testReward := uint64(5)
	difficulty := uint64(1)
	testMaxTxCount := uint32(3)

	mockCtrl := gomock.NewController(t)
	mockPoW := mockengine.NewMockPoW(mockCtrl)
	mockChain := mockchain.NewMockChain(mockCtrl)

	mockPoW.EXPECT().Difficulty().AnyTimes().Return(difficulty)
	mockPoW.EXPECT().Reward().AnyTimes().Return(testReward)

	validator := validator.NewGossipValidator(testMaxTxCount, mockPoW, mockChain)

	validBlock := blocktest.NewBlock(t, []*pb.Transaction{
		txtest.NewTransaction(t, 3, 1, 5),
		txtest.NewTransaction(t, 7, 1, 8),
		txtest.NewRewardTransaction(t, testReward+1+1),
	})

	s := testutil.NewSimpleState(t)
	err := s.UpdateAccount(
		[]byte(txtest.TxSenderPID),
		&pb.Account{Balance: 15, Nonce: 3},
	)
	assert.NoError(t, err)

	t.Run("ValidHeader", func(t *testing.T) {
		mockPoW.EXPECT().VerifyHeader(gomock.Any(), gomock.Any()).Times(1).Return(nil)

		err := validator.ValidateBlock(validBlock, s)
		assert.NoError(t, err)
	})

	t.Run("InvalidHeader", func(t *testing.T) {
		mockPoW.EXPECT().VerifyHeader(gomock.Any(), gomock.Any()).Times(1).Return(engine.ErrInvalidBlockNumber)

		err := validator.ValidateBlock(validBlock, s)
		assert.EqualError(t, engine.ErrInvalidBlockNumber, err.Error())
	})
}
