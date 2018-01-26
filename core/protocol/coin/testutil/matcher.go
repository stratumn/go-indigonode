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

package testutil

import (
	"fmt"

	pb "github.com/stratumn/alice/pb/coin"
)

// TxMatcher matches transactions.
type TxMatcher struct {
	value int64
	nonce int64
}

// NewTxMatcher creates a TxMatcher.
func NewTxMatcher(expectedTx *pb.Transaction) TxMatcher {
	return TxMatcher{
		value: expectedTx.Value,
		nonce: expectedTx.Nonce,
	}
}

// Matches returns whether x is a match.
// For unit tests, matching on value and nonce should be enough.
func (m TxMatcher) Matches(x interface{}) bool {
	return m.value == x.(*pb.Transaction).Value &&
		m.nonce == x.(*pb.Transaction).Nonce
}

// String describes what the matcher matches.
func (m TxMatcher) String() string {
	return fmt.Sprintf("Matching on value=%d and nonce=%d", m.value, m.nonce)
}

// HeaderMatcher matches headers.
type HeaderMatcher struct {
	blockNumber int64
}

// NewHeaderMatcher creates a HeaderMatcher.
func NewHeaderMatcher(expectedHeader *pb.Header) HeaderMatcher {
	return HeaderMatcher{
		blockNumber: expectedHeader.BlockNumber,
	}
}

// Matches returns whether x is a match.
// For unit tests, matching on block number should be enough.
func (m HeaderMatcher) Matches(x interface{}) bool {
	return m.blockNumber == x.(*pb.Header).BlockNumber
}

// String describes what the matcher matches.
func (m HeaderMatcher) String() string {
	return fmt.Sprintf("Matching on block number=%d", m.blockNumber)
}
