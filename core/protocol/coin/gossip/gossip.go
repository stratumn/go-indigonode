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

	txSubscription *floodsub.Subscription
}

// NewGossip returns gossip.
func NewGossip(
	p floodsub.PubSub,
	s state.Reader,
	v validator.Validator,
) *Gossip {
	return &Gossip{
		pubsub:    p,
		state:     s,
		validator: v,
	}
}

// SubscribeTx subscribes to the transaction topic.
func (g *Gossip) SubscribeTx() error {
	sub, err := g.pubsub.Subscribe(TxTopicName)
	if err != nil {
		return err
	}

	g.txSubscription = sub

	return g.pubsub.RegisterTopicValidator(TxTopicName, func(ctx context.Context, m *floodsub.Message) bool {
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
}

// ListenTx listens to incoming transactions.
func (g *Gossip) ListenTx(ctx context.Context, callback func(*pb.Transaction) error, errChan chan<- error) {
	go func() {
		msg, errIncoming := g.txSubscription.Next(ctx)
		for errIncoming == nil {
			tx := &pb.Transaction{}
			if err := tx.Unmarshal(msg.GetData()); err != nil {
				log.Event(ctx, "InvalidTxFormat", logging.Metadata{"error": err})
			}
			if err := callback(tx); err != nil {
				log.Event(ctx, "ProcessIncomingTxFailed", logging.Metadata{"error": err})
			}

			msg, errIncoming = g.txSubscription.Next(ctx)
		}

		errChan <- errIncoming
		log.Warningf("stopped listening to transactions: %v", errIncoming.Error())
	}()
}

// PublishTx publishes a transaction message.
func (g *Gossip) PublishTx(tx *pb.Transaction) error {
	txData, err := tx.Marshal()
	if err != nil {
		return err
	}

	return g.pubsub.Publish(TxTopicName, txData)
}
