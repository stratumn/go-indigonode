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

//go:generate mockgen -package mockcoin -destination mockcoin/mockmempool.go github.com/stratumn/alice/core/protocol/coin Mempool

package coin

import pb "github.com/stratumn/alice/pb/coin"

// Mempool stores transactions that need to be processed.
type Mempool interface {
	// AddTransaction adds a transaction to the mempool.
	// It assumes that the transaction has been validated.
	AddTransaction(tx *pb.Transaction) error
}
