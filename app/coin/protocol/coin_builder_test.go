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

package protocol

import (
	"testing"

	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/chain"
	"github.com/stratumn/go-indigonode/app/coin/protocol/engine"
	"github.com/stratumn/go-indigonode/app/coin/protocol/gossip"
	"github.com/stratumn/go-indigonode/app/coin/protocol/p2p"
	"github.com/stratumn/go-indigonode/app/coin/protocol/processor"
	"github.com/stratumn/go-indigonode/app/coin/protocol/state"
	"github.com/stratumn/go-indigonode/app/coin/protocol/synchronizer"
	"github.com/stratumn/go-indigonode/app/coin/protocol/testutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/validator"
	"github.com/stratumn/go-indigonode/core/db"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

var CoinBuilderDefaultGen = &pb.Block{Header: &pb.Header{Nonce: 42, Version: 32}}

// CoinBuilder is a utility to create a coin protocol with custom mocks
// to simulate a wide variety of configurations.
type CoinBuilder struct {
	genesisBlock *pb.Block
	chain        chain.Chain
	engine       engine.PoW
	gossip       gossip.Gossip
	p2p          p2p.P2P
	processor    processor.Processor
	state        state.State
	synchronizer synchronizer.Synchronizer
	txpool       state.TxPool
	validator    validator.Validator

	db db.DB

	minerID peer.ID

	difficulty    uint64
	maxTxPerBlock uint32
	reward        uint64
}

// NewCoinBuilder creates a CoinBuilder with good default values to build
// a test coin protocol. By default it uses an in-memory DB for the state
// and the chain, and real implementations of the underlying components.
func NewCoinBuilder(t *testing.T) *CoinBuilder {
	db, err := db.NewMemDB(nil)
	require.NoError(t, err, "db.NewMemDB()")

	return &CoinBuilder{
		db:            db,
		minerID:       "alice",
		difficulty:    1,
		maxTxPerBlock: 2,
		reward:        5,
	}
}

// WithMinerID configures the builder to use the given miner peer ID.
func (c *CoinBuilder) WithMinerID(id peer.ID) *CoinBuilder {
	c.minerID = id
	return c
}

// MinerID returns the peer ID that will be used for the miner.
func (c *CoinBuilder) MinerID() peer.ID {
	return c.minerID
}

// WithDifficulty configures the builder to use the given block difficulty.
func (c *CoinBuilder) WithDifficulty(difficulty uint64) *CoinBuilder {
	c.difficulty = difficulty
	return c
}

// WithMaxTxPerBlock configures the builder to use the given maximum
// number of transactions per block.
func (c *CoinBuilder) WithMaxTxPerBlock(maxTxPerBlock uint32) *CoinBuilder {
	c.maxTxPerBlock = maxTxPerBlock
	return c
}

// WithReward configures the builder to use the given block reward.
func (c *CoinBuilder) WithReward(reward uint64) *CoinBuilder {
	c.reward = reward
	return c
}

// WithChain configures the builder to use the given chain.
func (c *CoinBuilder) WithChain(chain chain.Chain) *CoinBuilder {
	c.chain = chain
	return c
}

// WithEngine configures the builder to use the given engine.
func (c *CoinBuilder) WithEngine(engine engine.PoW) *CoinBuilder {
	c.engine = engine
	return c
}

// WithGossip configures the builder to use the given gossip.
func (c *CoinBuilder) WithGossip(gossip gossip.Gossip) *CoinBuilder {
	c.gossip = gossip
	return c
}

// WithProcessor configures the builder to use the given processor.
func (c *CoinBuilder) WithProcessor(processor processor.Processor) *CoinBuilder {
	c.processor = processor
	return c
}

// WithState configures the builder to use the given state.
func (c *CoinBuilder) WithState(state state.State) *CoinBuilder {
	c.state = state
	return c
}

// WithTxPool configures the builder to use the given txpool.
func (c *CoinBuilder) WithTxPool(txpool state.TxPool) *CoinBuilder {
	c.txpool = txpool
	return c
}

// WithValidator configures the builder to use the given validator.
func (c *CoinBuilder) WithValidator(validator validator.Validator) *CoinBuilder {
	c.validator = validator
	return c
}

// WithP2P configures the builder to use the given p2p apis.
func (c *CoinBuilder) WithP2P(p2p p2p.P2P) *CoinBuilder {
	c.p2p = p2p
	return c
}

// WithSynchronizer configures the builder to use the given synchronizer.
func (c *CoinBuilder) WithSynchronizer(synchronizer synchronizer.Synchronizer) *CoinBuilder {
	c.synchronizer = synchronizer
	return c
}

// WithGenesisBlock configures the builder to use the given genesis block.
func (c *CoinBuilder) WithGenesisBlock(b *pb.Block) *CoinBuilder {
	c.genesisBlock = b
	return c
}

// Build builds the underlying coin protocol and returns it.
// It's now ready to use in your tests.
func (c *CoinBuilder) Build(t *testing.T) *Coin {
	// Set meaningful default for components that weren't configured
	// by the caller.
	if c.txpool == nil {
		c.txpool = &state.GreedyInMemoryTxPool{}
	}
	if c.state == nil {
		c.state = state.NewState(c.db, state.OptPrefix([]byte("s")))
	}
	if c.chain == nil {
		c.chain = chain.NewChainDB(c.db, chain.OptPrefix([]byte("c")))
	}
	if c.engine == nil {
		c.engine = engine.NewHashEngine(c.minerID, c.difficulty, c.reward)
	}
	if c.gossip == nil {
		c.gossip = testutil.NewDummyGossip(t)
	}
	if c.processor == nil {
		c.processor = processor.NewProcessor(nil)
	}
	if c.validator == nil {
		c.validator = validator.NewBalanceValidator(c.maxTxPerBlock, c.engine)
	}
	if c.p2p == nil {
		c.p2p = testutil.NewDummyP2P(t)
	}
	if c.synchronizer == nil {
		// TODO: something smart, a mock or a dummy sync.
		c.synchronizer = nil
	}

	if c.genesisBlock == nil {
		c.genesisBlock = CoinBuilderDefaultGen
	}

	return NewCoin(
		c.genesisBlock,
		c.txpool,
		c.engine,
		c.state,
		c.chain,
		c.gossip,
		c.validator,
		c.processor,
		c.p2p,
		c.synchronizer,
	)
}
