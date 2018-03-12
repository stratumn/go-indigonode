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

package coinutil

import (
	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stratumn/merkle"

	"gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
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
