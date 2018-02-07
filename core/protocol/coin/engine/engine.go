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

//go:generate mockgen -package mockengine -destination mockengine/mockengine.go github.com/stratumn/alice/core/protocol/coin/engine Engine

package engine

import (
	"context"
	"errors"

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
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

	// Reward returns the reward that the miner will receive in a coinbase
	// transaction when producing a new valid block.
	Reward() uint64
}
