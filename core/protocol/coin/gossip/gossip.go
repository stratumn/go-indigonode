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
	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"

	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/validator"
	pb "github.com/stratumn/alice/pb/coin"
)

const (
	// TXTopicName is the topic name for transactions.
	TXTopicName = "cointx"
)

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

// Subscribe subsscribes to the relevant topics.
func (g *Gossip) Subscribe() error {
	return g.SubscribeTXTopic()
}

// SubscribeTXTopic subscribes to the transaction topic.
func (g *Gossip) SubscribeTXTopic() error {
	sub, err := g.pubsub.Subscribe(TXTopicName)
	if err != nil {
		return err
	}

	g.txSubscription = sub

	return g.pubsub.RegisterTopicValidator(TXTopicName, func(ctx context.Context, m *floodsub.Message) bool {
		tx := &pb.Transaction{}
		if err := tx.Unmarshal(m.GetData()); err != nil {
			return false
		}

		if err := g.validator.ValidateTx(tx, g.state); err != nil {
			return false
		}

		return true
	})
}

// PublishTX publishes a transaction message.
func (g *Gossip) PublishTX(tx *pb.Transaction) error {
	txData, err := tx.Marshal()
	if err != nil {
		return err
	}

	return g.pubsub.Publish(TXTopicName, txData)
}
