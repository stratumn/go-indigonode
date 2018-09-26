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

package coinutil

import (
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/merkle"

	"gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
)

// HashHeader computes the hash of a given header.
func HashHeader(header *pb.Header) (multihash.Multihash, error) {
	b, err := header.Marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// TODO: make hash algo configurable
	headerHash, err := multihash.Sum(b, multihash.SHA2_256, -1)
	return headerHash[:], err
}

// HashTransaction computes the hash of a given transaction.
func HashTransaction(tx *pb.Transaction) (multihash.Multihash, error) {
	b, err := tx.Marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// TODO: make hash algo configurable
	txHash, err := multihash.Sum(b, multihash.SHA2_256, -1)
	return txHash[:], err
}

// TransactionRoot computes the merkle root of a set of transactions.
func TransactionRoot(txs []*pb.Transaction) ([]byte, error) {
	var leaves [][]byte
	for _, tx := range txs {
		txHash, err := HashTransaction(tx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		leaves = append(leaves, txHash)
	}

	tree, err := merkle.NewStaticTree(leaves)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tree.Root(), nil
}
