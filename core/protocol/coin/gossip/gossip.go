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

package gossip

import (
	"context"

	"github.com/pkg/errors"

	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"

	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

const (
	// TxTopicName is the topic name for transactions.
	TxTopicName = "coin.tx"
)

// log is the logger for the coin gossip.
var log = logging.Logger("coin.gossip")

// Gossip handles the gossiping of blocks and transactions.
type Gossip struct {
	pubsub    floodsub.PubSub
	state     state.Reader
	validator validator.Validator
	coinCtx   context.Context

	txSubscription *floodsub.Subscription
}

// NewGossip returns gossip.
func NewGossip(
	coinCtx context.Context,
	p floodsub.PubSub,
	s state.Reader,
	v validator.Validator,
) *Gossip {
	return &Gossip{
		pubsub:    p,
		state:     s,
		validator: v,
		coinCtx:   coinCtx,
	}
}

// SubscribeTx subscribes to the transaction topic.
func (g *Gossip) SubscribeTx() error {
	sub, err := g.pubsub.Subscribe(TxTopicName)
	if err != nil {
		return err
	}

	g.txSubscription = sub

	err = g.pubsub.RegisterTopicValidator(TxTopicName, func(ctx context.Context, m *floodsub.Message) bool {
		tx := &pb.Transaction{}
		if err := tx.Unmarshal(m.GetData()); err != nil {
			log.Infof("invalid transaction format: %v", err.Error())
			return false
		}

		if err := g.validator.ValidateTx(tx, g.state); err != nil {
			log.Infof("invalid transaction: %v", err.Error())
			return false
		}

		return true
	})

	return errors.WithStack(err)
}

// ListenTx listens to incoming transactions.
func (g *Gossip) ListenTx(callback func(*pb.Transaction) error) error {
	if !g.isSubscribed(TxTopicName) {
		return errors.New("subscribe to tx topic first")
	}

	go func() {
		msg, errIncoming := g.txSubscription.Next(g.coinCtx)
		for errIncoming == nil {
			tx := &pb.Transaction{}
			if err := tx.Unmarshal(msg.GetData()); err != nil {
				log.Event(g.coinCtx, "InvalidTxFormat", logging.Metadata{"error": err})
			}
			if err := callback(tx); err != nil {
				log.Event(g.coinCtx, "ProcessIncomingTxFailed", logging.Metadata{"error": err})
			}

			msg, errIncoming = g.txSubscription.Next(g.coinCtx)
		}

		log.Errorf("stopped listening to transactions: %v", errIncoming.Error())
	}()

	return nil
}

// PublishTx publishes a transaction message.
func (g *Gossip) PublishTx(tx *pb.Transaction) error {
	txData, err := tx.Marshal()
	if err != nil {
		return err
	}

	return g.pubsub.Publish(TxTopicName, txData)
}

func (g *Gossip) isSubscribed(topic string) bool {
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
