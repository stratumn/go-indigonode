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

package coin

import (
	pb "github.com/stratumn/alice/pb/coin"
)

// Engine is an algorithm agnostic consensus engine.
type Engine interface {
	// VerifyHeader checks whether a header conforms to the consensus
	// rules of a given engine.
	VerifyHeader(chain ChainReader, header *pb.Header) error

	// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
	// concurrently. The method returns a quit channel to abort the operations and
	// a results channel to retrieve the async verifications (the order is that of
	// the input slice).
	VerifyHeaders(chain ChainReader, headers []*pb.Header) (chan<- struct{}, <-chan error)

	// Prepare initializes the consensus fields of a block header according to the
	// rules of a particular engine. The changes are executed inline.
	Prepare(chain ChainReader, header *pb.Header) error

	// Finalize verifies transactions, updates the state to apply them
	// and assembles the final block.
	// The block header might be updated (including a merkle root of the
	// transactions for example).
	Finalize(chain ChainReader, header *pb.Header, state *State, txs []*pb.Transaction) (*pb.Block, error)
}
