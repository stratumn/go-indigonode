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

package blocktest

import (
	"testing"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"
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
