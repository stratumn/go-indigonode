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

// Package coin is a service to interact with Alice coins.
// It runs a proof of work consensus engine between Alice nodes.
package coin

import (
	"context"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/db"
	protocol "github.com/stratumn/alice/core/protocol/coin"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/engine"
	"github.com/stratumn/alice/core/protocol/coin/gossip"
	"github.com/stratumn/alice/core/protocol/coin/p2p"
	"github.com/stratumn/alice/core/protocol/coin/processor"
	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/synchronizer"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	rpcpb "github.com/stratumn/alice/grpc/coin"
	pb "github.com/stratumn/alice/pb/coin"

	"google.golang.org/grpc"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	kaddht "gx/ipfs/QmT9TxakNKCHg3uBcLnNzBSBhhACvqH8tRzJvYZjUevrvE/go-libp2p-kad-dht"
	floodsub "gx/ipfs/QmVKrsEgixRtMWcMd6WQzuwqCUC3jfLf7Q7xcjnKoMMikS/go-libp2p-floodsub"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrNotPubSub is returned when the connected service is not a pubsub.
	ErrNotPubSub = errors.New("connected service is not a pubsub")

	// ErrNotKadDHT is returned when the connected service is not a pubsub.
	ErrNotKadDHT = errors.New("connected service is not a kaddht")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")

	// ErrMissingMinerID is returned when the miner's peer ID is missing
	// from the configuration file.
	ErrMissingMinerID = errors.New("the miner's peer ID should be configured")
)

var log = logging.Logger("coin")

// Host represents an Alice host.
type Host = ihost.Host

// Service is the Coin service.
type Service struct {
	config *Config
	host   Host
	coin   *protocol.Coin
	pubsub *floodsub.PubSub
	kaddht *kaddht.IpfsDHT
}

// Config contains configuration options for the Coin service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Version is the version of the coin service.
	Version int `toml:"version" comment:"The version of the coin service."`

	// MaxTxPerBlock is the maximum number of transactions in a block.
	MaxTxPerBlock int `toml:"max_tx_per_block" comment:"The maximum number of transactions in a block."`

	// MinerReward is the reward miners should get when producing blocks.
	MinerReward int `toml:"miner_reward" comment:"The reward miners should get when producing blocks."`

	// BlockDifficulty is the difficulty for block production.
	BlockDifficulty int `toml:"block_difficulty" comment:"The difficulty for block production."`

	// DbPath is the path to the database used for the state and the chain..
	DbPath string `toml:"db_path" comment:"The path to the database used for the state and the chain."`

	// PubSub is the name of the pubsub service.
	PubSub string `toml:"pubsub" comment:"The name of the pubsub service."`

	// MinerID is the peer ID of the miner.
	// Block rewards will be sent to this peer.
	// Note that a miner can generate as many peer IDs as it wants and use any of them.
	MinerID string `toml:"miner_id" comment:"The peer ID of the miner."`

	// KadDHT is the name of the kaddht service.
	KadDHT string `toml:"kaddht" comment:"The name of the kaddht service."`

	// GenesisBlock is the genesis block in hex. If none given, we use the default one in protocol/coin.
	GenesisBlock string `toml:"genesis_block" comment:"The genesis block in hex."`
}

// GetMinerID reads the miner's peer ID from the configuration.
func (c *Config) GetMinerID() (peer.ID, error) {
	if len(c.MinerID) == 0 {
		return "", ErrMissingMinerID
	}

	minerID, err := peer.IDB58Decode(c.MinerID)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return minerID, nil
}

// GetGenesisBlock gets the genesis block from the config if present.
func (c *Config) GetGenesisBlock() (*pb.Block, error) {
	// If no genesis block passed in the config, we will use the default one.
	if len(c.GenesisBlock) == 0 {
		return GetGenesisBlock()
	}

	b, err := hex.DecodeString(c.GenesisBlock)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	block := &pb.Block{}
	if err = block.Unmarshal(b); err != nil {
		return nil, err
	}

	return block, nil
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "coin"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Coin"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "A service to use Alice coins."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.WithStack(err))
	}

	coinDBPath, err := filepath.Abs(filepath.Join(cwd, "data", "coin", "db"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	// Set the default configuration settings of your service here.
	return Config{
		Host:            "host",
		Version:         1,
		MaxTxPerBlock:   100,
		MinerReward:     10,
		BlockDifficulty: 29,
		DbPath:          coinDBPath,
		PubSub:          "pubsub",
		KadDHT:          "kaddht",
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Host] = struct{}{}
	needs[s.config.PubSub] = struct{}{}
	needs[s.config.KadDHT] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.host, ok = exposed[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	if s.pubsub, ok = exposed[s.config.PubSub].(*floodsub.PubSub); !ok {
		return errors.Wrap(ErrNotPubSub, s.config.PubSub)
	}

	if s.kaddht, ok = exposed[s.config.KadDHT].(*kaddht.IpfsDHT); !ok {
		return errors.Wrap(ErrNotKadDHT, s.config.KadDHT)
	}

	return nil
}

// Expose exposes the coin service to other services.
func (s *Service) Expose() interface{} {
	return s.coin
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	coinCtx, cancel := context.WithCancel(ctx)
	errChan := make(chan error)

	close, err := s.createCoin(coinCtx)
	if err != nil {
		cancel()
		return err
	}
	defer close()

	if err := s.coin.Run(coinCtx); err != nil {
		cancel()
		return err
	}

	go func() {
		errChan <- s.coin.StartMining(coinCtx)
	}()

	// Wrap the stream handler with the context.
	handler := func(stream inet.Stream) {
		s.coin.StreamHandler(ctx, stream)
	}

	s.host.SetStreamHandler(protocol.ProtocolID, handler)

	running()
	<-ctx.Done()
	stopping()

	// Stop accepting streams.
	s.host.RemoveStreamHandler(protocol.ProtocolID)

	cancel()

	err = <-errChan
	s.coin = nil

	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(ctx.Err())
}

func (s *Service) createCoin(ctx context.Context) (func(), error) {
	db, err := db.NewFileDB(s.config.DbPath, nil)
	if err != nil {
		return nil, err
	}

	stateDBPrefix := []byte("s")
	chainDBPrefix := []byte("c")

	txpool := &state.GreedyInMemoryTxPool{}
	state := state.NewState(db, state.OptPrefix(stateDBPrefix))
	chain := chain.NewChainDB(db, chain.OptPrefix(chainDBPrefix))

	minerID, err := s.config.GetMinerID()
	if err != nil {
		return nil, err
	}
	genesisBlock, err := s.config.GetGenesisBlock()
	if err != nil {
		return nil, err
	}

	engine := engine.NewHashEngine(minerID, uint64(s.config.BlockDifficulty), uint64(s.config.MinerReward))

	processor := processor.NewProcessor(s.kaddht)
	balanceValidator := validator.NewBalanceValidator(uint32(s.config.MaxTxPerBlock), engine)
	gossipValidator := validator.NewGossipValidator(uint32(s.config.MaxTxPerBlock), engine, chain)
	p2p := p2p.NewP2P(s.host, protocol.ProtocolID)
	sync := synchronizer.NewSynchronizer(p2p, s.kaddht)
	gossip := gossip.NewGossip(s.host, s.pubsub, state, chain, gossipValidator)

	s.coin = protocol.NewCoin(genesisBlock, txpool, engine, state, chain, gossip, balanceValidator, processor, p2p, sync)

	close := func() {
		if err := gossip.Close(); err != nil {
			log.Event(ctx, "gossip.Close()", logging.Metadata{"error": err.Error()})
		}
		if err := db.Close(); err != nil {
			log.Event(ctx, "db.Close()", logging.Metadata{"error": err.Error()})
		}
	}

	return close, nil
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	rpcpb.RegisterCoinServer(gs, grpcServer{
		DoGetAccount: func(peerID []byte) (*pb.Account, error) {
			if s.coin == nil {
				return nil, ErrUnavailable
			}

			return s.coin.GetAccount(peerID)
		},
		AddTransaction: func(tx *pb.Transaction) error {
			if s.coin == nil {
				return ErrUnavailable
			}
			if err := s.coin.AddTransaction(tx); err != nil {
				return err
			}

			return s.coin.PublishTransaction(tx)
		},
		GetAccountTransactions: func(peerID []byte) ([]*pb.Transaction, error) {
			if s.coin == nil {
				return nil, ErrUnavailable
			}

			return s.coin.GetAccountTransactions(peerID)
		},
		GetBlockchain: func(blockNumber uint64, hash []byte, count uint32) ([]*pb.Block, error) {
			if s.coin == nil {
				return nil, ErrUnavailable
			}

			return s.coin.GetBlockchain(blockNumber, hash, count)
		},
		GetTransactionPool: func(count uint32) (uint64, []*pb.Transaction, error) {
			if s.coin == nil {
				return 0, nil, ErrUnavailable
			}

			return s.coin.GetTransactionPool(count)
		},
	})
}
