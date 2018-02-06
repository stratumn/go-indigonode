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
	"crypto/sha256"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/types"
)

// HashHeader computes the hash of a given header.
func HashHeader(header *pb.Header) ([]byte, error) {
	b, err := header.Marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	headerHash := sha256.Sum256(b)
	return headerHash[:], nil
}

// HashBlock computes the hash of a given block.
func HashBlock(block *pb.Block) ([]byte, error) {
	b, err := block.Marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	blockHash := sha256.Sum256(b)
	return blockHash[:], nil
}

// HashTransaction computes the hash of a given transaction.
func HashTransaction(tx *pb.Transaction) ([]byte, error) {
	b, err := tx.Marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	txHash := sha256.Sum256(b)
	return txHash[:], nil
}

// TransactionRoot computes the merkle root of a set of transactions.
func TransactionRoot(txs []*pb.Transaction) ([]byte, error) {
	var leaves []types.Bytes32
	for _, tx := range txs {
		txHash, err := HashTransaction(tx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		leaves = append(leaves, *types.NewBytes32FromBytes(txHash))
	}

	tree, err := merkle.NewStaticTree(leaves)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tree.Root()[:], nil
}
