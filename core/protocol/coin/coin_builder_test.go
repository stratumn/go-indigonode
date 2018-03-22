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
	"testing"

	"github.com/stratumn/alice/core/db"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/gossip"
	"github.com/stratumn/alice/core/protocol/coin/p2p"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/synchronizer"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
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
