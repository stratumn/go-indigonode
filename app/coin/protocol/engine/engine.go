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
	"errors"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/state"
)

var (
	// ErrInvalidPreviousBlock is the error returned by VerifyHeader
	// when the previous block referenced can't be found.
	ErrInvalidPreviousBlock = errors.New("invalid header: previous block could not be found")

	// ErrInvalidBlockNumber is the error returned  by VerifyHeader
	// when the block number is invalid.
	ErrInvalidBlockNumber = errors.New("invalid header: block number is invalid")

	// ErrDifficultyNotMet is the error returned  by VerifyHeader
	// when the block's difficulty has not been met (for PoW).
	ErrDifficultyNotMet = errors.New("invalid header: difficulty not met")

	// ErrInvalidChain is the error returned when the input chain
	// returns invalid or unactionable results.
	ErrInvalidChain = errors.New("invalid chain")
)

// Engine is an algorithm agnostic consensus engine.
// It verifies that headers are valid and properly assembles blocks
// according to the consensus rules.
// It doesn't change any state, it's the responsibility of the caller
// to process the blocks assembled by the engine.
type Engine interface {
	// VerifyHeader checks whether a header conforms to the consensus
	// rules of a given engine.
	VerifyHeader(chain chain.Reader, header *pb.Header) error

	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain chain.Reader, header *pb.Header) error

	// Finalize assumes that transactions have already been validated and don't
	// violate state rules.
	// It assembles the transactions in a final block, potentially adding more
	// transactions than those from the input (for example, block rewards).
	// The block header might be updated (including a merkle root of the
	// transactions, or computing a valid nonce for a proof-of-work algorithm
	// for example).
	Finalize(ctx context.Context, chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error)
}

// ProofOfWait is a consensus engine based on the powerful proof-of-wait © algorithm.
// It is only meant to be used as a PoC and in tests.
type ProofOfWait interface {
	Engine

	// Interval returns the minimum and maximum number of
	// milliseconds necessary to produce a block.
	Interval() (int, int)
}

// PoW is a consensus engine based on a proof-of-work algorithm.
type PoW interface {
	Engine

	// Difficulty returns the difficulty of the proof-of-work.
	// Blocks that don't conform to the current difficulty will be rejected.
	Difficulty() uint64

	// Reward returns the reward that the miner will receive in a reward
	// transaction when producing a new valid block.
	Reward() uint64
}
