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

//go:generate mockgen -package mockgossip -destination mockgossip/mockgossip.go github.com/stratumn/alice/core/protocol/coin/gossip Gossip

package gossip

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"

	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	host "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

const (
	// TxTopicName is the topic name for transactions.
	TxTopicName = "coin.tx"

	// BlockTopicName is the topic name for blocks.
	BlockTopicName = "coin.block"
)

// log is the logger for the coin gossip.
var log = logging.Logger("coin.gossip")

// Gossip is an interface used to gossip transactions and blocks.
type Gossip interface {
	// ListenTx passes incoming transactions to a callback.
	// Only valid transactions will be passed to the callback.
	ListenTx(ctx context.Context, processTx func(*pb.Transaction) error) error
	// ListenBlock passes incoming blocks to a callback.
	// Only valid blocks will be passed to the callback.
	ListenBlock(ctx context.Context, processBlock func(*pb.Block) error) error
	// PublishTx sends a new transaction to the gossip.
	PublishTx(tx *pb.Transaction) error
	// PublishBlock sends a new block to the gossip.
	PublishBlock(block *pb.Block) error

	// Close closes the gossip layer.
	Close() error
}

// Gossip handles the gossiping of blocks and transactions.
type gossip struct {
	host      host.Host
	pubsub    *floodsub.PubSub
	state     state.Reader
	validator validator.Validator

	mu   sync.RWMutex
	subs map[string]*floodsub.Subscription
}

// NewGossip returns gossip.
func NewGossip(
	h host.Host,
	p *floodsub.PubSub,
	s state.Reader,
	v validator.Validator,
) Gossip {
	return &gossip{
		host:      h,
		pubsub:    p,
		state:     s,
		validator: v,
		subs:      make(map[string]*floodsub.Subscription),
	}
}

func (g *gossip) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, s := range g.subs {
		s.Cancel()
	}

	// TODO: Remove topic validators once it is possible.
	// https://github.com/libp2p/go-floodsub/issues/68

	return nil
}

// ListenTx subscribes to transaction topic and listens to incoming transactions.
func (g *gossip) ListenTx(ctx context.Context, processTx func(*pb.Transaction) error) error {
	if err := g.subscribeTx(); err != nil {
		return err
	}

	return g.listen(ctx, TxTopicName, func(msg *floodsub.Message) error {
		tx := &pb.Transaction{}
		if err := tx.Unmarshal(msg.GetData()); err != nil {
			return err
		}

		from, _ := peer.IDFromBytes(tx.From)
		to, _ := peer.IDFromBytes(tx.To)
		log.Event(context.Background(), "TransactionReceived", &logging.Metadata{
			"From":  peer.IDB58Encode(from),
			"To":    peer.IDB58Encode(to),
			"Value": tx.Value,
			"Fee":   tx.Fee,
			"Nonce": tx.Nonce,
		})

		return processTx(tx)
	})
}

// ListenBlock subscribes to block topic and listens to incoming blocks.
func (g *gossip) ListenBlock(ctx context.Context, processBlock func(*pb.Block) error) error {
	if err := g.subscribeBlock(); err != nil {
		return err
	}

	return g.listen(ctx, BlockTopicName, func(msg *floodsub.Message) error {
		block := &pb.Block{}
		if err := block.Unmarshal(msg.GetData()); err != nil {
			return err
		}

		log.Event(context.Background(), "BlockReceived", &logging.Metadata{
			"BlockNumber":  block.BlockNumber(),
			"PreviousHash": block.PreviousHash(),
			"Nonce":        block.Nonce(),
			"TxCount":      len(block.Transactions),
		})

		return processBlock(block)
	})
}

// PublishTx publishes a transaction message.
func (g *gossip) PublishTx(tx *pb.Transaction) error {
	txData, err := tx.Marshal()
	if err != nil {
		return err
	}

	return g.publish(TxTopicName, txData)
}

// PublishBlock publishes a block message.
func (g *gossip) PublishBlock(block *pb.Block) error {
	blockData, err := block.Marshal()
	if err != nil {
		return err
	}

	return g.publish(BlockTopicName, blockData)
}

// subscribeTx subscribes to the transaction topic.
func (g *gossip) subscribeTx() error {
	return g.subscribe(TxTopicName, func(ctx context.Context, m *floodsub.Message) bool {
		tx := &pb.Transaction{}
		if err := tx.Unmarshal(m.GetData()); err != nil {
			log.Event(ctx, "InvalidTransactionFormat", logging.Metadata{"error": err.Error()})
			return false
		}

		if err := g.validator.ValidateTx(tx, g.state); err != nil {
			log.Event(ctx, "InvalidTransaction", logging.Metadata{"error": err.Error()})
			return false
		}

		return true
	})
}

// subscribeBlock subscribes to the block topic.
func (g *gossip) subscribeBlock() error {
	return g.subscribe(BlockTopicName, func(ctx context.Context, m *floodsub.Message) bool {
		block := &pb.Block{}
		if err := block.Unmarshal(m.GetData()); err != nil {
			log.Event(ctx, "InvalidBlockFormat", logging.Metadata{"error": err.Error()})
			return false
		}

		if err := g.validator.ValidateBlock(block, g.state); err != nil {
			log.Event(ctx, "InvalidBlock", logging.Metadata{"error": err.Error()})
			return false
		}

		return true
	})
}

func (g *gossip) subscribe(topic string, validator floodsub.Validator) error {
	if g.isSubscribed(topic) {
		return nil
	}

	sub, err := g.pubsub.Subscribe(topic)
	if err != nil {
		return err
	}

	g.subs[topic] = sub

	// We can't remove a topic validator for now...
	// https://github.com/libp2p/go-floodsub/issues/68
	err = g.pubsub.RegisterTopicValidator(topic, validator)
	if err != nil && err.Error() == fmt.Sprintf("Duplicate validator for topic %s", topic) {
		return nil
	}
	return errors.WithStack(err)
}

func (g *gossip) listen(ctx context.Context, topic string, callback func(msg *floodsub.Message) error) error {
	if !g.isSubscribed(topic) {
		return errors.Errorf("not subscribed to %v topic", topic)
	}

	sub := g.subs[topic]

	go func() {
		msg, errIncoming := sub.Next(ctx)
		for errIncoming == nil {
			if g.host.ID() != msg.GetFrom() {
				if err := callback(msg); err != nil {
					seqno := binary.BigEndian.Uint64(msg.GetSeqno())
					log.Event(ctx, "UnableToProcessMessage", logging.Metadata{
						"topic": topic,
						"from":  msg.GetFrom().Pretty(),
						"seqno": seqno,
						"error": err.Error(),
					})
				}
			}
			msg, errIncoming = sub.Next(ctx)
		}
	}()

	return nil
}

func (g *gossip) publish(topic string, data []byte) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	err := g.pubsub.Publish(topic, data)
	if err != nil {
		log.Event(context.Background(), "UnableToPublish", logging.Metadata{"topic": topic, "error": err.Error()})
	}

	return err
}

func (g *gossip) isSubscribed(topic string) bool {
	isSubscribed := false
	topics := g.pubsub.GetTopics()

	for _, t := range topics {
		if t == topic {
			isSubscribed = true
			break
		}
	}

	return isSubscribed
}
