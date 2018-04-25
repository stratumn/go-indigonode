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

// Package fossilizer contains the Indigo Fossilizer service.
package fossilizer

import (
	"context"
	"time"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/batchfossilizer"
	"github.com/stratumn/go-indigocore/bcbatchfossilizer"
	"github.com/stratumn/go-indigocore/blockchain/btc"
	"github.com/stratumn/go-indigocore/blockchain/btc/blockcypher"
	"github.com/stratumn/go-indigocore/blockchain/btc/btctimestamper"
	"github.com/stratumn/go-indigocore/blockchain/dummytimestamper"
	"github.com/stratumn/go-indigocore/dummyfossilizer"
	"github.com/stratumn/go-indigocore/fossilizer"
)

const (
	// Dummy designates the dummyfossilizer type.
	Dummy = "dummy"

	// DummyBatch designates the dummybatchfossilizer type.
	DummyBatch = "dummybatch"

	// BlockchainBatch designates the bcbatchfossilizer type.
	BlockchainBatch = "bcbatch"

	// BitcoinTimestamper designates the bitcoin timestamper.
	BitcoinTimestamper = "bitcoin"

	// DummyTimestamper designates the dummy timestamper.
	DummyTimestamper = "dummy"
)

var (
	// ErrNotImplemented is returned when trying to instantiate an unknown type of fossilizer.
	ErrNotImplemented = errors.New("fossilizer type is not implemented")

	// log is the logger for the configuration package.
	log = logging.Logger("indigo.fossilizer.config")
)

// Config contains configuration options for the Fossilizer service.
type Config struct {
	// Version is the version of the Indigo Fossilizer service.
	Version string `toml:"version" comment:"The version of the indigo fossilizer service."`

	// FossilizerType is the fossilizer implementation.
	FossilizerType string `toml:"fossilizer_type" comment:"The type of fossilizer (eg: dummy, dummybatch, bitcoin...)."`

	// Timestamper is the backend of the timestamping mechanism.
	Timestamper string `toml:"timestamper" comment:"The backend for timestamping (only applicable to blockchain fossilizers)."`

	// Interval between batches (if any).
	Interval int64 `toml:"interval" comment:"The time interval between batches expressed in seconds (only applicable to fossilizers using batches)."`

	// Maximum number of leaves of a Merkle tree.
	MaxLeaves int `toml:"max_leaves" comment:"The maximum number of leaves of a merkle tree in a batch (only applicable to fossilizers using batches)."`

	// BtcWIF is the Wallet Import Format encoded secret key.
	BtcWIF string `toml:"bitcoin_WIF" comment:"Wallet Import Format encoded secret key used to send transactions to the bitcoin blockchain (only applicable to the bitcoin fossilizer)."`

	// BtcFee is the fee to use when sending transactions to the bitcoin blockchain.
	BtcFee int64 `toml:"bitcoin_fee" comment:"amount of the fee to use when sending transactions to the bitcoin blockchain (only applicable to the bitcoin fossilizer)."`
}

// CreateIndigoFossilizer creates an indigo fossilizer from the configuration.
func (c *Config) CreateIndigoFossilizer(ctx context.Context) (fossilizer.Adapter, error) {
	switch c.FossilizerType {
	case Dummy:
		return dummyfossilizer.New(&dummyfossilizer.Config{}), nil
	case DummyBatch:
		return batchfossilizer.New(&batchfossilizer.Config{
			Interval:  time.Duration(c.Interval),
			MaxLeaves: c.MaxLeaves,
		})
	case BlockchainBatch:
		return c.createBlockchainFossilizer(ctx)
	default:
		return nil, ErrNotImplemented
	}
}

func (c *Config) createBlockchainFossilizer(ctx context.Context) (fossilizer.Adapter, error) {
	log.Event(ctx, "createBlockchainFossilizer", &logging.Metadata{
		"fossilizerConfig": c,
	})

	switch c.Timestamper {
	case BitcoinTimestamper:
		btcNetwork, err := btc.GetNetworkFromWIF(c.BtcWIF)
		if err != nil {
			return nil, err
		}
		bcy := blockcypher.New(&blockcypher.Config{
			Network:         btcNetwork,
			LimiterInterval: blockcypher.DefaultLimiterInterval,
			LimiterSize:     blockcypher.DefaultLimiterSize,
		})
		timestamper, err := btctimestamper.New(&btctimestamper.Config{
			UnspentFinder: bcy,
			Broadcaster:   bcy,
			WIF:           c.BtcWIF,
			Fee:           c.BtcFee,
		})
		if err != nil {
			return nil, err
		}
		return bcbatchfossilizer.New(&bcbatchfossilizer.Config{
			HashTimestamper: timestamper,
		}, &batchfossilizer.Config{
			Interval:  time.Duration(c.Interval),
			MaxLeaves: c.MaxLeaves,
		})
	case DummyTimestamper:
		return bcbatchfossilizer.New(&bcbatchfossilizer.Config{
			HashTimestamper: &dummytimestamper.Timestamper{},
		}, &batchfossilizer.Config{
			Interval:  time.Duration(c.Interval),
			MaxLeaves: c.MaxLeaves,
		})
	default:
		return nil, ErrNotImplemented
	}
}
