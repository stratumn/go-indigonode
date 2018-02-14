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

	block1 := &pb.Block{
		Header: &pb.Header{
			BlockNumber: 0,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    bob,
			Value: 10,
			Nonce: 1,
		}},
	}
	h1, err := coinutil.HashHeader(block1.Header)
	assert.NoError(t, err, "HashHeader(block1)")

	block2 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  1,
			PreviousHash: h1,
		},
		Transactions: []*pb.Transaction{{
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
	h2, err := coinutil.HashHeader(block2.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	assert.NoError(t, p.Process(block1, s, c))
	assert.NoError(t, p.Process(block2, s, c))
	assert.NoError(t, p.Process(block2, s, c))

	// Check chain
	b, err := c.GetBlock(h1, 0)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block1, b, "GetBlock()")

	b, err = c.GetBlock(h2, 1)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block2, b, "GetBlock()")

	header, err := c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block2.Header, header, "CurrentHeader()")

	headers, err := c.GetHeadersByNumber(1)
	assert.NoError(t, err, "GetHeadersByNumber()")
	assert.Len(t, headers, 1, "GetHeadersByNumber()")

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

func TestProcessor_ProcessWithReorg(t *testing.T) {
	/* Test a chain reorg
	In this case, we start with a chain containing block1->block21->block31 (same as above).
	We then add other blocks forking the chain from block1 until it is higher that the main chain:

	block1  --->  block21  --->  block31
				﹨
				 ﹨-->  block22  ---> block32 ---> block4

	The test checks that transactions from block21 and block31 are rollbacked in right order
	and that transactions from block22, block32 and block4 are applied to the state.
	*/

	alice := []byte("alice")
	bob := []byte("bob")
	charlie := []byte("charlie")
	dexter := []byte("dexter")

	p := processor.NewProcessor()
	s := testutil.NewSimpleState(t, state.OptPrefix([]byte("s")))
	c := &testutil.SimpleChain{}

	err := s.UpdateAccount(alice, &pb.Account{Balance: 20})
	assert.NoError(t, err, "s.UpdateAccount(alice)")

	block1 := &pb.Block{
		Header: &pb.Header{
			BlockNumber: 0,
			Nonce:       1,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    bob,
			Value: 10,
			Nonce: 1,
		}},
	}
	h1, err := coinutil.HashHeader(block1.Header)
	assert.NoError(t, err, "HashHeader(block1)")

	block21 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  1,
			PreviousHash: h1,
			Nonce:        21,
		},
		Transactions: []*pb.Transaction{{
			From:  bob,
			To:    charlie,
			Value: 7,
			Nonce: 2,
		}, {
			From:  charlie,
			To:    alice,
			Value: 2,
			Nonce: 3,
		}},
	}
	h21, err := coinutil.HashHeader(block21.Header)
	assert.NoError(t, err, "HashHeader(block21)")

	block31 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  2,
			PreviousHash: h21,
			Nonce:        31,
		},
		Transactions: []*pb.Transaction{{
			From:  bob,
			To:    dexter,
			Value: 3,
			Nonce: 2,
		}},
	}
	h31, err := coinutil.HashHeader(block31.Header)
	assert.NoError(t, err, "HashHeader(block31)")

	block22 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  1,
			PreviousHash: h1,
			Nonce:        22,
		},
		Transactions: []*pb.Transaction{{
			From:  bob,
			To:    dexter,
			Value: 5,
			Nonce: 2,
		}},
	}
	h22, err := coinutil.HashHeader(block22.Header)
	assert.NoError(t, err, "HashHeader(block22)")

	block32 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  2,
			PreviousHash: h22,
			Nonce:        32,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    dexter,
			Value: 4,
			Nonce: 2,
		}},
	}
	h32, err := coinutil.HashHeader(block32.Header)
	assert.NoError(t, err, "HashHeader(block32)")

	block4 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  3,
			PreviousHash: h32,
			Nonce:        4,
		},
		Transactions: []*pb.Transaction{{
			From:  dexter,
			To:    bob,
			Value: 1,
			Nonce: 42,
		}},
	}
	h4, err := coinutil.HashHeader(block4.Header)
	assert.NoError(t, err, "HashHeader(block3)")

	assert.NoError(t, p.Process(block1, s, c))
	assert.NoError(t, p.Process(block21, s, c))
	assert.NoError(t, p.Process(block31, s, c))
	assert.NoError(t, p.Process(block22, s, c))
	assert.NoError(t, p.Process(block32, s, c))
	assert.NoError(t, p.Process(block4, s, c))

	// Check chain
	b, err := c.GetBlock(h1, 0)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block1, b, "GetBlock()")

	b, err = c.GetBlock(h21, 1)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block21, b, "GetBlock()")

	b, err = c.GetBlock(h22, 1)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block22, b, "GetBlock()")

	b, err = c.GetBlock(h31, 2)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block31, b, "GetBlock()")

	b, err = c.GetBlock(h32, 2)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block32, b, "GetBlock()")

	b, err = c.GetBlock(h4, 2)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block4, b, "GetBlock()")

	header, err := c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block4.Header, header, "CurrentHeader()")

	// Check state
	v, err := s.GetAccount(alice)
	assert.NoError(t, err, "s.GetAccount(alice)")
	assert.Equal(t, &pb.Account{Balance: 20 - 10 - 4, Nonce: 2}, v, "s.GetAccount(alice)")

	v, err = s.GetAccount(bob)
	assert.NoError(t, err, "s.GetAccount(bob)")
	assert.Equal(t, &pb.Account{Balance: 10 - 5 + 1, Nonce: 2}, v, "s.GetAccount(bob)")

	v, err = s.GetAccount(charlie)
	assert.NoError(t, err, "s.GetAccount(charlie)")
	assert.Equal(t, &pb.Account{}, v, "s.GetAccount(charlie)")

	v, err = s.GetAccount(dexter)
	assert.NoError(t, err, "s.GetAccount(dexter)")
	assert.Equal(t, &pb.Account{Balance: 5 + 4 - 1, Nonce: 42}, v, "s.GetAccount(dexter)")
}
