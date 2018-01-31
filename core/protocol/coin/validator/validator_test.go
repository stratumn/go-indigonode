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

package validator

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
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
		ErrEmptyTx,
	}, {
		"empty-value",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 0, 42)
			return tx
		},
		func() state.State { return nil },
		ErrInvalidTxValue,
	}, {
		"missing-to",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.To = nil
			return tx
		},
		func() state.State { return nil },
		ErrInvalidTxRecipient,
	}, {
		"missing-from",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.From = nil
			return tx
		},
		func() state.State { return nil },
		ErrInvalidTxSender,
	}, {
		"send-to-self",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.From = tx.To
			return tx
		},
		func() state.State { return nil },
		ErrInvalidTxRecipient,
	}, {
		"missing-signature",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.Signature = nil
			return tx
		},
		func() state.State { return nil },
		ErrMissingTxSignature,
	}, {
		"invalid-signature",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.Signature.Signature[3] ^= 1
			return tx
		},
		func() state.State { return nil },
		ErrInvalidTxSignature,
	}, {
		"invalid-balance",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			return tx
		},
		func() state.State { return testutil.NewSimpleState(t) },
		ErrInsufficientBalance,
	}, {
		"valid-tx",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			return tx
		},
		func() state.State {
			state := testutil.NewSimpleState(t)
			_, err := state.AddBalance([]byte(testutil.TxSenderPID), 80)
			assert.NoError(t, err)
			return state
		},
		nil,
	}}

	validator := NewBalanceValidator(DefaultMaxTxPerBlock)

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
			return &pb.Block{
				Transactions: []*pb.Transaction{
					testutil.NewTransaction(t, 1, 1),
					testutil.NewTransaction(t, 2, 2),
					testutil.NewTransaction(t, 1, 3),
					testutil.NewTransaction(t, 3, 4),
				},
			}
		},
		func() state.State { return nil },
		ErrTooManyTxs,
	}, {
		"invalid-signature",
		func() *pb.Block {
			block := &pb.Block{
				Transactions: []*pb.Transaction{
					testutil.NewTransaction(t, 3, 5),
					testutil.NewTransaction(t, 7, 1),
				},
			}

			block.Transactions[1].Signature.Signature[5] ^= block.Transactions[1].Signature.Signature[5]

			return block
		},
		func() state.State { return nil },
		ErrInvalidTxSignature,
	}, {
		"invalid-balance",
		func() *pb.Block {
			return &pb.Block{
				Transactions: []*pb.Transaction{
					testutil.NewTransaction(t, 3, 5),
					testutil.NewTransaction(t, 7, 1),
				},
			}
		},
		func() state.State {
			state := testutil.NewSimpleState(t)
			_, err := state.AddBalance([]byte(testutil.TxSenderPID), 8)
			assert.NoError(t, err)
			return state
		},
		ErrInsufficientBalance,
	}, {
		"valid-block",
		func() *pb.Block {
			return &pb.Block{
				Transactions: []*pb.Transaction{
					testutil.NewTransaction(t, 3, 5),
					testutil.NewTransaction(t, 7, 1),
				},
			}
		},
		func() state.State {
			state := testutil.NewSimpleState(t)
			_, err := state.AddBalance([]byte(testutil.TxSenderPID), 10)
			assert.NoError(t, err)
			return state
		},
		nil,
	}}

	validator := NewBalanceValidator(3)

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
