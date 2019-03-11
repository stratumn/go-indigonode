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

package processor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/coinutil"
	"github.com/stratumn/go-node/app/coin/protocol/processor"
	"github.com/stratumn/go-node/app/coin/protocol/state"
	"github.com/stratumn/go-node/app/coin/protocol/testutil"
	txtest "github.com/stratumn/go-node/app/coin/protocol/testutil/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cid "github.com/ipfs/go-cid"
)

var (
	alice, bob, charlie, dexter []byte
	accounts                    map[string]*pb.Account
)

func init() {
	accounts = make(map[string]*pb.Account)

	for _, user := range []*[]byte{&alice, &bob, &charlie, &dexter} {
		_, _, pid, err := txtest.NewKeyPair()
		if err != nil {
			panic(err)
		}
		*user = pid
		accounts[string(*user)] = &pb.Account{Balance: 20, Nonce: 0}
	}
}

type provider struct {
	resources map[string]bool
}

func (p *provider) Provide(ctx context.Context, key cid.Cid, brdcst bool) error {
	if p.resources == nil {
		p.resources = map[string]bool{}
	}
	if brdcst {
		p.resources[key.String()] = true
	}
	return nil
}

func TestProcessor_Process(t *testing.T) {
	dht := &provider{}
	p := processor.NewProcessor(dht)
	s := testutil.NewSimpleState(t, state.OptPrefix([]byte("s")))
	c := &testutil.SimpleChain{}

	for user, account := range accounts {
		err := s.UpdateAccount([]byte(user), account)
		assert.NoError(t, err, fmt.Sprintf("s.UpdateAccount(%v)", user))
	}

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

	ctx := context.Background()

	assert.NoError(t, p.Process(ctx, block1, s, c))
	checkState(t, accounts, []*pb.Block{block1}, s)

	assert.NoError(t, p.Process(ctx, block2, s, c))
	checkState(t, accounts, []*pb.Block{block1, block2}, s)

	assert.NoError(t, p.Process(ctx, block2, s, c))
	checkState(t, accounts, []*pb.Block{block1, block2}, s)

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

	// Check provider
	cid1, err := cid.Cast(h1)
	assert.NoError(t, err, "cid.Cast()")
	assert.True(t, dht.resources[cid1.String()], "dht.Provide()")
	cid2, err := cid.Cast(h2)
	assert.NoError(t, err, "cid.Cast()")
	assert.True(t, dht.resources[cid2.String()], "dht.Provide()")
}

func TestProcessor_ProcessWithReorgs(t *testing.T) {
	/* Test a chain reorg
	In this case, we start with a chain containing block1->block21->block31 (same as above).
	We then add other blocks forking the chain from block1 until it is higher that the main chain:

	block1  --->  block21  --->  block31
				﹨
				 ﹨-->  block22  ---> block32 ---> block4

	The test checks that transactions from block21 and block31 are rollbacked in right order
	and that transactions from block22, block32 and block4 are applied to the state.
	*/

	dht := &provider{}
	p := processor.NewProcessor(dht)
	s := testutil.NewSimpleState(t, state.OptPrefix([]byte("s")))
	c := &testutil.SimpleChain{}

	for user, account := range accounts {
		err := s.UpdateAccount([]byte(user), account)
		assert.NoError(t, err, fmt.Sprintf("s.UpdateAccount(%v)", user))
	}

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
			Nonce: 3,
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
			From:  dexter,
			To:    bob,
			Value: 1,
			Nonce: 42,
		}},
	}
	h32, err := coinutil.HashHeader(block32.Header)
	assert.NoError(t, err, "HashHeader(block32)")

	block4 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  3,
			PreviousHash: h32,
			Nonce:        41,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    dexter,
			Value: 4,
			Nonce: 2,
		}},
	}
	h4, err := coinutil.HashHeader(block4.Header)
	assert.NoError(t, err, "HashHeader(block3)")

	block4invalid := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  3,
			PreviousHash: h32,
			Nonce:        42,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    dexter,
			Value: 100,
			Nonce: 4,
		}},
	}

	ctx := context.Background()
	assert.NoError(t, p.Process(ctx, block1, s, c))
	assert.NoError(t, p.Process(ctx, block21, s, c))
	assert.NoError(t, p.Process(ctx, block31, s, c))
	assert.NoError(t, p.Process(ctx, block22, s, c))
	assert.NoError(t, p.Process(ctx, block32, s, c))

	// Check state no reorg no new head.
	checkState(t, accounts, []*pb.Block{block1, block21, block31}, s)
	header, err := c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block31.Header, header)

	// Check state no reorg invalid block.
	assert.NoError(t, p.Process(ctx, block4invalid, s, c))

	checkState(t, accounts, []*pb.Block{block1, block21, block31}, s)
	header, err = c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block31.Header, header)

	// Adding block 4 will trigger reorg.
	assert.NoError(t, p.Process(ctx, block4, s, c))

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

	header, err = c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block4.Header, header, "CurrentHeader()")

	checkState(t, accounts, []*pb.Block{block1, block22, block32, block4}, s)
}

func TestProcessor_ProcessWithInvalidState(t *testing.T) {
	dht := &provider{}
	p := processor.NewProcessor(dht)
	s := testutil.NewSimpleState(t, state.OptPrefix([]byte("s")))
	c := &testutil.SimpleChain{}

	for user, account := range accounts {
		err := s.UpdateAccount([]byte(user), account)
		assert.NoError(t, err, fmt.Sprintf("s.UpdateAccount(%v)", user))
	}

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

	// block2 is invalid because of the transaction nonce.
	block2 := &pb.Block{
		Header: &pb.Header{
			BlockNumber:  1,
			PreviousHash: h1,
		},
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    bob,
			Value: 5,
			Nonce: 1,
		}},
	}
	h2, err := coinutil.HashHeader(block2.Header)
	assert.NoError(t, err, "coinutil.HashHeader()")

	ctx := context.Background()

	assert.NoError(t, p.Process(ctx, block1, s, c))
	checkState(t, accounts, []*pb.Block{block1}, s)

	assert.NoError(t, p.Process(ctx, block2, s, c))
	checkState(t, accounts, []*pb.Block{block1}, s)

	// Check chain
	b, err := c.GetBlock(h1, 0)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block1, b, "GetBlock()")

	b, err = c.GetBlock(h2, 1)
	assert.NoError(t, err, "GetBlock()")
	assert.Equal(t, block2, b, "GetBlock()")

	header, err := c.CurrentHeader()
	assert.NoError(t, err, "CurrentHeader()")
	assert.Equal(t, block1.Header, header, "CurrentHeader()")

}

func checkState(t *testing.T, accounts map[string]*pb.Account, blocks []*pb.Block, s state.Reader) {
	a := make(map[string]*pb.Account)

	for user, account := range accounts {
		a[user] = &pb.Account{Balance: account.Balance, Nonce: account.Nonce}
	}

	for _, b := range blocks {
		for _, tx := range b.GetTransactions() {
			a[string(tx.From)].Balance -= tx.Value + tx.Fee
			a[string(tx.To)].Balance += tx.Value
			a[string(tx.From)].Nonce = tx.Nonce
		}
	}

	for u, account := range a {
		aa, err := s.GetAccount([]byte(u))

		require.NoError(t, err)
		assert.Equal(t, account, aa)
	}
}
