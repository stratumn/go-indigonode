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
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// HashEngine is an engine that uses hashes of the block as a proof-of-work.
// It tries to find a Nonce so that the hash of the whole block starts with
// a given number of 0.
type HashEngine struct {
	difficulty uint64
}

// NewHashEngine creates a PoW engine with an initial difficulty.
func NewHashEngine(difficulty uint64) PoW {
	return &HashEngine{difficulty: difficulty}
}

func (e *HashEngine) VerifyHeader(chain chain.Reader, header *pb.Header) error {
	panic("not implemented")
}

func (e *HashEngine) Prepare(chain chain.Reader, header *pb.Header) error {
	panic("not implemented")
}

func (e *HashEngine) Finalize(chain chain.Reader, header *pb.Header, state state.Reader, txs []*pb.Transaction) (*pb.Block, error) {
	panic("not implemented")
}

// Difficulty returns the number of leading 0 we expect from the block hash.
func (e *HashEngine) Difficulty() uint64 {
	return e.difficulty
}
