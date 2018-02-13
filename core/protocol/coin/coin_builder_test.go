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

	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	"github.com/stratumn/alice/core/protocol/coin/db"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/gossip"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/require"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// CoinBuilder is a utility to create a coin protocol with custom mocks
// to simulate a wide variety of configurations.
type CoinBuilder struct {
	chain     chain.Chain
	engine    engine.PoW
	gossip    gossip.Gossip
	processor processor.Processor
	state     state.State
	txpool    state.TxPool
	validator validator.Validator

	db db.DB

	pk *coinutil.PublicKey

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

	_, pubKey, err := ic.GenerateKeyPair(ic.Ed25519, 0)
	require.NoError(t, err, "ic.GenerateKeyPair()")

	return &CoinBuilder{
		db:            db,
		pk:            coinutil.NewPublicKey(pubKey, pb.KeyType_Ed25519),
		difficulty:    1,
		maxTxPerBlock: 2,
		reward:        5,
	}
}

// WithPublicKey configures the builder to use the given miner public key.
func (c *CoinBuilder) WithPublicKey(pk *coinutil.PublicKey) *CoinBuilder {
	c.pk = pk
	return c
}

// PublicKey returns the public key that will be used for the miner.
func (c *CoinBuilder) PublicKey() *coinutil.PublicKey {
	return c.pk
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

// Build builds the underlying coin protocol and returns it.
// It's now ready to use in your tests.
func (c *CoinBuilder) Build() *Coin {
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
		c.engine = engine.NewHashEngine(c.pk, c.difficulty, c.reward)
	}
	if c.processor == nil {
		c.processor = processor.NewProcessor()
	}
	if c.validator == nil {
		c.validator = validator.NewBalanceValidator(c.maxTxPerBlock, c.engine)
	}

	return NewCoin(
		c.txpool,
		c.engine,
		c.state,
		c.chain,
		c.gossip,
		c.validator,
		c.processor,
	)
}
