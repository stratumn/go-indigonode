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

package processor_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_Process(t *testing.T) {
	alice := []byte("alice")
	bob := []byte("bob")
	charlie := []byte("charlie")

	p := processor.NewProcessor()
	s := testutil.NewSimpleState(t, state.OptPrefix([]byte("s")))
	c := &testutil.SimpleChain{}

	err := s.UpdateAccount(alice, &pb.Account{Balance: 20})
	assert.NoError(t, err, "s.UpdateAccount(alice)")

	block := &pb.Block{
		Header: &pb.Header{
			BlockNumber: 0,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    bob,
			Value: 10,
			Nonce: 1,
		}, {
			From:  bob,
			To:    charlie,
			Value: 5,
			Nonce: 2,
		}, {
			From:  charlie,
			To:    alice,
			Value: 2,
			Nonce: 3,
		}},
	}

	assert.NoError(t, p.Process(block, s, c))

	// Check chain
	h, err := coinutil.HashHeader(block.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	b, err := c.GetBlock(h[:], 0)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block, b, "GetBlock()")

	header, err := c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block.Header, header, "CurrentHeader()")

	// Check state
	v, err := s.GetAccount(alice)
	assert.NoError(t, err, "s.GetAccount(alice)")
	assert.Equal(t, &pb.Account{Balance: 20 - 10 + 2, Nonce: 1}, v, "s.GetAccount(alice)")

	v, err = s.GetAccount(bob)
	assert.NoError(t, err, "s.GetAccount(bob)")
	assert.Equal(t, &pb.Account{Balance: 10 - 5, Nonce: 2}, v, "s.GetAccount(bob)")

	v, err = s.GetAccount(charlie)
	assert.NoError(t, err, "s.GetAccount(charlie)")
	assert.Equal(t, &pb.Account{Balance: 5 - 2, Nonce: 3}, v, "s.GetAccount(charlie)")
}
