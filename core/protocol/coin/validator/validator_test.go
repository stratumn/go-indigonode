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

	"github.com/stratumn/alice/core/protocol/coin/testutil"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

// TODO: test cases:
// Validate signature negative balance
// Validate block one invalid signature
// Validate valid block
// Validate block multiple txs, one becomes negative

type validateTxTestCase struct {
	name string
	tx   func() *pb.Transaction
	err  error
}

func TestValidateTx(t *testing.T) {
	testCases := []validateTxTestCase{{
		"empty-tx",
		func() *pb.Transaction { return nil },
		ErrEmptyTx,
	}, {
		"empty-value",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 0, 42)
			return tx
		},
		ErrInvalidTxValue,
	}, {
		"missing-to",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.To = nil
			return tx
		},
		ErrInvalidTxRecipient,
	}, {
		"missing-from",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.From = nil
			return tx
		},
		ErrInvalidTxSender,
	}, {
		"missing-signature",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.Signature = nil
			return tx
		},
		ErrMissingTxSignature,
	}, {
		"invalid-signature",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			tx.Signature.Signature[3] ^= 1
			return tx
		},
		ErrInvalidTxSignature,
	}, {
		"valid-tx",
		func() *pb.Transaction {
			tx := testutil.NewTransaction(t, 42, 42)
			return tx
		},
		nil,
	}}

	validator := NewBalanceValidator()

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTx(tt.tx(), nil)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
