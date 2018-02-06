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

package engine

import (
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutils"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// HashEngine is an engine that uses hashes of the block as a proof-of-work.
// It tries to find a Nonce so that the hash of the whole block starts with
// a given number of 0.
type HashEngine struct {
	difficulty uint64

	privKey *coinutils.PrivateKey
}

// NewHashEngine creates a PoW engine with an initial difficulty.
// The miner should specify its public key to receive block rewards.
func NewHashEngine(privKey *coinutils.PrivateKey, difficulty uint64) PoW {
	return &HashEngine{
		difficulty: difficulty,
		privKey:    privKey,
	}
}

// VerifyHeader verifies that the header met the PoW difficulty
// and correctly chains to a previous block.
func (e *HashEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	headerHash, err := coinutils.HashHeader(header)
	if err != nil {
		return errors.WithStack(err)
	}

	for i := uint64(0); i < e.difficulty; i++ {
		if headerHash[i] != 0 {
			return ErrDifficultyNotMet
		}
	}

	previousBlock, err := chain.GetBlock(header.PreviousHash, header.BlockNumber-1)
	if err != nil || previousBlock == nil {
		return ErrInvalidPreviousBlock
	}

	if previousBlock.Header.BlockNumber+1 != header.BlockNumber {
		return ErrInvalidBlockNumber
	}

	return nil
}

// Prepare initializes the fields of the given header to respect
// consensus rules.
func (e *HashEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	currentHeader, err := chain.CurrentHeader()
	if err != nil || currentHeader == nil {
		return ErrInvalidChain
	}

	previousHash, err := coinutils.HashHeader(currentHeader)
	if err != nil {
		return errors.WithStack(err)
	}

	header.BlockNumber = currentHeader.BlockNumber + 1
	header.MerkleRoot = nil
	header.Nonce = 0
	header.PreviousHash = previousHash
	header.Timestamp = ptypes.TimestampNow()
	header.Version = 1

	return nil
}

// Finalize verifies the header, create a reward for the miner and assembles
// the finalized block.
func (e *HashEngine) Finalize(chain chain.Reader, header *pb.Header, _ state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	previous, err := chain.GetHeaderByHash(header.PreviousHash)
	if err != nil || previous == nil {
		return nil, ErrInvalidPreviousBlock
	}

	if previous.BlockNumber+1 != header.BlockNumber {
		return nil, ErrInvalidBlockNumber
	}

	reward, err := e.createReward()
	if err != nil {
		return nil, err
	}

	blockTxs := append(txs, reward)
	blockRoot, err := coinutils.TransactionRoot(blockTxs)
	if err != nil {
		return nil, err
	}

	header.MerkleRoot = blockRoot

	block := &pb.Block{
		Header:       header,
		Transactions: blockTxs,
	}

	return block, nil
}

// Difficulty returns the number of leading 0 we expect from the header hash.
func (e *HashEngine) Difficulty() uint64 {
	return e.difficulty
}

// createReward creates a reward for the miner finalizing the block.
func (e *HashEngine) createReward() (*pb.Transaction, error) {
	to, err := e.privKey.GetPublicKey().Bytes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reward := &pb.Transaction{
		To:    to,
		Value: 5,
	}

	txBytes, err := reward.Marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sig, err := e.privKey.Sign(txBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reward.Signature = &pb.Signature{
		KeyType:   pb.KeyType_Ed25519,
		PublicKey: to,
		Signature: sig,
	}

	return reward, nil
}
