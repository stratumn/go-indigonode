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

// Protocol describes the interface exposed to other nodes in the network.
type Protocol interface {
	// AddTransaction validates a transaction and adds it to the mempool.
	AddTransaction(tx *pb.Transaction) error

	// AppendBlock validates the incoming block and adds it at the end of
	// the chain, updating internal state to reflect the block's transactions.
	// Note: it might cause the consensus engine to stop mining on an outdated
	// block and mine on top of the newly added block.
	AppendBlock(block *pb.Block) error
}
