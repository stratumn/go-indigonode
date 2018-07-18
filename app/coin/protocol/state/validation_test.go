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

package state

import (
	"testing"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/testutil/blocktest"
	txtest "github.com/stratumn/go-indigonode/app/coin/protocol/testutil/transaction"
	"github.com/stratumn/go-indigonode/core/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSimpleState(t *testing.T, opts ...Opt) State {
	memdb, err := db.NewMemDB(nil)
	require.NoError(t, err, "db.NewMemDB()")

	return NewState(memdb, opts...)
}

func TestValidateBalance(t *testing.T) {
	type validateTxTestCase struct {
		name  string
		tx    func() *pb.Transaction
		state func() State
		err   error
	}

	testCases := []validateTxTestCase{{
		"invalid-balance",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 0, 42)
			return tx
		},
		func() State { return newSimpleState(t) },
		ErrInsufficientBalance,
	}, {
		"invalid-balance-fee",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 40, 5, 3)
			return tx
		},
		func() State {
			s := newSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 41, Nonce: 1},
			)
			assert.NoError(t, err)
			return s
		},
		ErrInsufficientBalance,
	}, {
		"invalid-nonce",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 3, 42)
			return tx
		},
		func() State {
			s := newSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 42},
			)
			assert.NoError(t, err)
			return s
		},
		ErrInvalidTxNonce,
	}, {
		"valid-tx",
		func() *pb.Transaction {
			tx := txtest.NewTransaction(t, 42, 3, 42)
			return tx
		},
		func() State {
			s := newSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 80, Nonce: 40},
			)
			assert.NoError(t, err)
			return s
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBalance(tt.state(), tt.tx())
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBalances(t *testing.T) {
	testReward := uint64(5)

	type validateBlockTxTestCase struct {
		name  string
		block func() *pb.Block
		state func() State
		err   error
	}

	testCases := []validateBlockTxTestCase{{
		"invalid-balance",
		func() *pb.Block {
			return blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 3, 2, 5),
				txtest.NewTransaction(t, 3, 1, 6),
			})
		},
		func() State {
			s := newSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 8, Nonce: 1},
			)
			assert.NoError(t, err)
			return s
		},
		ErrInsufficientBalance,
	}, {
		"valid-block",
		func() *pb.Block {
			return blocktest.NewBlock(t, []*pb.Transaction{
				txtest.NewTransaction(t, 3, 1, 5),
				txtest.NewTransaction(t, 7, 1, 8),
				txtest.NewRewardTransaction(t, testReward+1+1),
			})
		},
		func() State {
			s := newSimpleState(t)
			err := s.UpdateAccount(
				[]byte(txtest.TxSenderPID),
				&pb.Account{Balance: 15, Nonce: 3},
			)
			assert.NoError(t, err)
			return s
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBalances(tt.state(), tt.block().GetTransactions())
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
