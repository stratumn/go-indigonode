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
	"math/rand"
	"time"

	"github.com/pkg/errors"
	protocol "github.com/stratumn/alice/core/protocol/coin"
	"github.com/stratumn/alice/core/protocol/coin/chain"
	"github.com/stratumn/alice/core/protocol/coin/db"
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

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	kaddht "gx/ipfs/QmZcitfGoJ1JCwF9zdQvczfmwzLKugqtEGvTMXYcX23zbU/go-libp2p-kad-dht"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
	floodsub "gx/ipfs/QmctbcXMMhxTjm5ybWpjMwDmabB39ANuhB5QNn8jpD4JTv/go-libp2p-floodsub"
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

	// Set the default configuration settings of your service here.
	return Config{
		Host:            "host",
		Version:         1,
		MaxTxPerBlock:   100,
		MinerReward:     10,
		BlockDifficulty: 42,
		DbPath:          "data/coin/db",
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

	go s.genTxs()

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
		GetAccount: func(peerID []byte) (*pb.Account, error) {
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

func (s *Service) genTxs() {
	miners := make([][]byte, 4)
	miner0, _ := peer.IDB58Decode("QmQnYf23kQ7SvuPZ3mQcg3RuJMr9E39fBvm89Nz4bevJdt")
	miners[0] = []byte(miner0)
	miner1, _ := peer.IDB58Decode("QmYxjwtGm1mKL61Cc6NhBRgMWFz39r3k5tRRqWWiQoux7V")
	miners[1] = []byte(miner1)
	miner2, _ := peer.IDB58Decode("QmcqyMD1UrnqtVgoG5rHaXmQfqg8LK6Wk6NUpvdUmf78Rv")
	miners[2] = []byte(miner2)
	miner3, _ := peer.IDB58Decode("QmNnP696qYWjwyiqoepYauwpdZp62NoKEjBV4Q1MJVhMUC")
	miners[3] = []byte(miner3)

	pubKeys := make([][]byte, 4)
	privKeys := make([]ic.PrivKey, 4)

	miner0KeyBytes, _ := ic.ConfigDecodeKey("CAESYE53i0HQ1GuPegCRNgLOObZYe4zPV14nYTTNZ60+wM8q/jnmvcWtU8wnGcpwhfVqBs0aFPpIn0z3Tunw+tmkpyX+Oea9xa1TzCcZynCF9WoGzRoU+kifTPdO6fD62aSnJQ==")
	miner0Key, _ := ic.UnmarshalPrivateKey(miner0KeyBytes)
	miner0PubKey, _ := miner0Key.GetPublic().Bytes()
	pubKeys[0] = miner0PubKey
	privKeys[0] = miner0Key

	miner1KeyBytes, _ := ic.ConfigDecodeKey("CAESYKUMPQcudYkJcdhZnnidF8vjesrPyIFO7hw18QaDfq0v1DVxc8WTp+XQIbaMTooGqthRWNK89DmpJ7R/W8MVqNLUNXFzxZOn5dAhtoxOigaq2FFY0rz0OakntH9bwxWo0g==")
	miner1Key, _ := ic.UnmarshalPrivateKey(miner1KeyBytes)
	miner1PubKey, _ := miner1Key.GetPublic().Bytes()
	pubKeys[1] = miner1PubKey
	privKeys[1] = miner1Key

	miner2KeyBytes, _ := ic.ConfigDecodeKey("CAESYFKsxpr3ZqpsW5GEPngFnHPOhehbz3ZXWITZ1uXcjt+6MC26htfs4o7NwRLTh6o5ddRvNhU7LxY4074ctTHHENIwLbqG1+zijs3BEtOHqjl11G82FTsvFjjTvhy1MccQ0g==")
	miner2Key, _ := ic.UnmarshalPrivateKey(miner2KeyBytes)
	miner2PubKey, _ := miner2Key.GetPublic().Bytes()
	pubKeys[2] = miner2PubKey
	privKeys[2] = miner2Key

	miner3KeyBytes, _ := ic.ConfigDecodeKey("CAESYF/pw0CjTfxADsQSPTz0JPb8EtK76sR13dXIgv0J3/DjNFjaBhPYAb6V/jw2oQbL89VKYz0qcIeoCYVh3/ywkog0WNoGE9gBvpX+PDahBsvz1UpjPSpwh6gJhWHf/LCSiA==")
	miner3Key, _ := ic.UnmarshalPrivateKey(miner3KeyBytes)
	miner3PubKey, _ := miner3Key.GetPublic().Bytes()
	pubKeys[3] = miner3PubKey
	privKeys[3] = miner3Key

	// Every 30s gen a random tx.
	for _ = range time.NewTicker(time.Second * 30).C {
		from := rand.Intn(4)
		to := rand.Intn(4)
		for from == to {
			to = rand.Intn(4)
		}
		val := rand.Intn(12)
		for val == 0 {
			val = rand.Intn(4)
		}

		tx := &pb.Transaction{
			From:  []byte(miners[from]),
			To:    []byte(miners[to]),
			Nonce: uint64(time.Now().UnixNano()),
			Value: uint64(val),
		}

		txb, _ := tx.Marshal()

		sig, _ := privKeys[from].Sign(txb)
		tx.Signature = &pb.Signature{
			KeyType:   pb.KeyType_Ed25519,
			PublicKey: pubKeys[from],
			Signature: sig,
		}

		s.coin.AddTransaction(tx)

		s.coin.PublishTransaction(tx)

	}

}
