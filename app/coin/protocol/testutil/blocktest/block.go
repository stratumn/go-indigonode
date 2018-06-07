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

package blocktest

import (
	"testing"

	"github.com/stratumn/alice/app/coin/pb"
	"github.com/stratumn/alice/app/coin/protocol/coinutil"
	"github.com/stretchr/testify/require"
)

// NewBlock creates a new block containing a set of transactions.
func NewBlock(t testing.TB, txs []*pb.Transaction) *pb.Block {
	root, err := coinutil.TransactionRoot(txs)
	require.NoError(t, err, "coinutil.TransactionRoot()")

	block := &pb.Block{
		Header: &pb.Header{
			Version:     1,
			BlockNumber: 1,
			MerkleRoot:  root,
		},
		Transactions: txs,
	}

	return block
}
