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

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
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

	// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
	// concurrently. The input context can be used for cancellation.
	// The method returns a results channel to retrieve the async verifications
	// (the order is that of the input slice).
	VerifyHeaders(ctx context.Context, chain chain.Reader, headers []*pb.Header) <-chan error

	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain chain.Reader, header *pb.Header) error

	// Finalize assumes that transactions have already been validated and don't
	// violate state rules.
	// It assembles the transactions in a final block, potentially adding more
	// transactions than those from the input (for example, block rewards).
	// The block header might be updated (including a merkle root of the
	// transactions for example).
	Finalize(chain chain.Reader, header *pb.Header, state *state.Reader, txs []*pb.Transaction) (*pb.Block, error)
}
