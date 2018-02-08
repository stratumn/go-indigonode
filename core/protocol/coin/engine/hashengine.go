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
	"context"
	"math"

	ptypes "github.com/gogo/protobuf/types"
	multihash "github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// HashEngine is an engine that uses hashes of the block as a proof-of-work.
// It tries to find a Nonce so that the hash of the whole block starts with
// a given number of 0.
type HashEngine struct {
	difficulty uint64
	reward     uint64

	pubKey *coinutil.PublicKey
}

// NewHashEngine creates a PoW engine with an initial difficulty.
// The miner should specify its public key to receive block rewards.
func NewHashEngine(pubKey *coinutil.PublicKey, difficulty, reward uint64) PoW {
	return &HashEngine{
		difficulty: difficulty,
		reward:     reward,
		pubKey:     pubKey,
	}
}

// VerifyHeader verifies that the header met the PoW difficulty
// and correctly chains to a previous block.
func (e *HashEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	if err := e.verifyPow(header); err != nil {
		return err
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

	previousHash, err := coinutil.HashHeader(currentHeader)
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

// Finalize verifies the header, creates a reward for the miner, finds a nonce
// that meets the PoW difficulty and assembles the finalized block.
// You can abort Finalize by cancelling the context (if for example a new block
// arrives and you want to stop mining on an outdated block).
func (e *HashEngine) Finalize(ctx context.Context, chain chain.Reader, header *pb.Header, _ state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	previous, err := chain.GetHeaderByHash(header.PreviousHash)
	if err != nil || previous == nil {
		return nil, ErrInvalidPreviousBlock
	}

	if previous.BlockNumber+1 != header.BlockNumber {
		return nil, ErrInvalidBlockNumber
	}

	reward, err := e.createReward(txs)
	if err != nil {
		return nil, err
	}

	blockTxs := append(txs, reward)
	blockRoot, err := coinutil.TransactionRoot(blockTxs)
	if err != nil {
		return nil, err
	}

	header.MerkleRoot = blockRoot

	block := &pb.Block{
		Header:       header,
		Transactions: blockTxs,
	}

	err = e.pow(ctx, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// Difficulty returns the number of leading 0 we expect from the header hash.
func (e *HashEngine) Difficulty() uint64 {
	return e.difficulty
}

// Reward returns the reward a miner will receive when producing
// a new valid block.
func (e *HashEngine) Reward() uint64 {
	return e.reward
}

// verifyPow verifies if the proof-of-work is valid.
func (e *HashEngine) verifyPow(header *pb.Header) error {
	headerHash, err := coinutil.HashHeader(header)
	if err != nil {
		return errors.WithStack(err)
	}

	// headerHash is a Multihash. Get the actual digest.
	mh, err := multihash.Decode(headerHash)
	if err != nil {
		return err
	}
	for i := uint64(0); i < e.difficulty; i++ {
		if mh.Digest[i/8]&(1<<(7-(i%8))) != 0 {
			return ErrDifficultyNotMet
		}
	}

	return nil
}

// pow finds a solution to the proof-of-work problem.
// It sets the block header's nonce appropriately.
func (e *HashEngine) pow(ctx context.Context, block *pb.Block) error {
	for i := uint64(0); i < math.MaxUint64; i++ {
		// Stop if the caller doesn't want to mine this block anymore.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		block.Header.Nonce = i
		if err := e.verifyPow(block.Header); err == nil {
			break
		}
	}

	return nil
}

// createReward creates a reward for the miner finalizing the block.
func (e *HashEngine) createReward(txs []*pb.Transaction) (*pb.Transaction, error) {
	to, err := e.pubKey.Bytes()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reward := &pb.Transaction{
		To:    to,
		Value: e.reward + coinutil.GetBlockFees(&pb.Block{Transactions: txs}),
	}

	return reward, nil
}
