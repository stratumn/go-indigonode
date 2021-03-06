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

package engine

import (
	"context"
	"math"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/chain"
	"github.com/stratumn/go-node/app/coin/protocol/coinutil"
	"github.com/stratumn/go-node/app/coin/protocol/state"

	"github.com/multiformats/go-multihash"
	peer "github.com/libp2p/go-libp2p-peer"
)

// HashEngine is an engine that uses hashes of the block as a proof-of-work.
// It tries to find a Nonce so that the hash of the whole block starts with
// a given number of 0.
type HashEngine struct {
	difficulty uint64
	reward     uint64

	minerID peer.ID
}

// NewHashEngine creates a PoW engine with an initial difficulty.
// The miner should specify its public key to receive block rewards.
func NewHashEngine(minerID peer.ID, difficulty, reward uint64) PoW {
	return &HashEngine{
		difficulty: difficulty,
		reward:     reward,
		minerID:    minerID,
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

	if previousBlock.BlockNumber()+1 != header.BlockNumber {
		return ErrInvalidBlockNumber
	}

	return nil
}

// Prepare initializes the fields of the given header to respect
// consensus rules.
func (e *HashEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	currentHeader, err := chain.CurrentHeader()
	if err != nil {
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
	reward := &pb.Transaction{
		To:    []byte(e.minerID),
		Value: e.reward + coinutil.GetBlockFees(&pb.Block{Transactions: txs}),
	}

	return reward, nil
}
