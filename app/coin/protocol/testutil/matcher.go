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

package testutil

import (
	"fmt"

	"github.com/stratumn/go-node/app/coin/pb"
)

// TxMatcher matches transactions.
type TxMatcher struct {
	value uint64
	nonce uint64
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
	blockNumber uint64
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
