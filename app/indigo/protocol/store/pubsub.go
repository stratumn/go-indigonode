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

package store

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/core/monitoring"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	floodsub "gx/ipfs/QmY1L5krVk8dv8d74uESmJTXGpoigVYqBVxXXz1aS8aFSb/go-libp2p-floodsub"
)

var (
	// ErrInvalidNetworkID is returned at startup when the network is misconfigured.
	ErrInvalidNetworkID = errors.New("invalid or missing network ID")
)

// PubSubNetworkManager implements the NetworkManager interface.
type PubSubNetworkManager struct {
	networkMutex sync.Mutex
	networkID    string
	host         Host
	pubsub       *floodsub.PubSub
	sub          *floodsub.Subscription

	listenersMutex sync.RWMutex
	listeners      []chan *cs.Segment
}

// NewPubSubNetworkManager creates a new NetworkManager that uses a PubSub
// mechanism to exchange messages with the network.
// This is particularly suited for public networks.
func NewPubSubNetworkManager() NetworkManager {
	return &PubSubNetworkManager{}
}

// NodeID returns the base58 peerID.
func (m *PubSubNetworkManager) NodeID() peer.ID {
	return m.host.ID()
}

// Join joins a PoP network that uses floodsub to share links.
// TODO: implement proper private networks / topic protection.
// It looks like floodsub is planning to implement authenticated topics
// and encryption modes (see in TopicDescriptor).
// It isn't implemented yet but once it is, this is what we should use.
func (m *PubSubNetworkManager) Join(ctx context.Context, networkID string, host Host) (err error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "Join")
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
		}

		span.End()
	}()

	span.AddStringAttribute("network_id", networkID)

	m.networkMutex.Lock()
	defer m.networkMutex.Unlock()

	// A network can only be joined once.
	if m.pubsub != nil {
		return nil
	}

	if networkID == "" {
		return ErrInvalidNetworkID
	}

	m.networkID = networkID

	pubsub, err := floodsub.NewFloodSub(ctx, host)
	if err != nil {
		return errors.WithStack(err)
	}

	sub, err := pubsub.Subscribe(networkID)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := pubsub.RegisterTopicValidator(
		networkID,
		// We always want to propagate messages:
		//	- if the message is valid, everyone should receive it
		//	- if it's not, everyone should have a chance to store it for auditing/non-repudiation
		func(context.Context, *floodsub.Message) bool { return true },
	); err != nil {
		return errors.WithStack(err)
	}

	m.pubsub = pubsub
	m.sub = sub
	m.host = host

	return nil
}

// Leave unregisters from the underlying pubsub.
func (m *PubSubNetworkManager) Leave(ctx context.Context, networkID string) (err error) {
	_, span := monitoring.StartSpan(ctx, "indigo.store", "Leave")
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
		}

		span.End()
	}()

	span.AddStringAttribute("network_id", networkID)

	m.networkMutex.Lock()
	defer m.networkMutex.Unlock()

	if networkID != m.networkID {
		return ErrInvalidNetworkID
	}

	if err := m.pubsub.UnregisterTopicValidator(m.networkID); err != nil {
		return errors.WithStack(err)
	}

	m.sub.Cancel()

	m.sub = nil
	m.pubsub = nil

	m.host.RemoveStreamHandler(floodsub.FloodSubID)

	return nil
}

// Publish shares a message with the network.
func (m *PubSubNetworkManager) Publish(ctx context.Context, link *cs.Link) (err error) {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "Publish")
	defer func() {
		if err != nil {
			span.SetUnknownError(err)
		}

		span.End()
	}()

	var pubsubPeers []string
	for _, peer := range m.pubsub.ListPeers(m.networkID) {
		pubsubPeers = append(pubsubPeers, peer.Pretty())
	}

	span.AddStringAttribute("peers", strings.Join(pubsubPeers, ","))

	key := m.host.Peerstore().PrivKey(m.host.ID())
	signedSegment, err := audit.SignLink(ctx, key, link)
	if err != nil {
		return err
	}

	signedSegmentBytes, err := json.Marshal(signedSegment)
	if err != nil {
		return errors.WithStack(err)
	}

	err = m.pubsub.Publish(m.networkID, signedSegmentBytes)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Listen to network messages and forwards them to listeners.
func (m *PubSubNetworkManager) Listen(ctx context.Context) error {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "Listen")
	defer span.End()

	defer func() {
		m.listenersMutex.Lock()
		defer m.listenersMutex.Unlock()

		for _, listener := range m.listeners {
			close(listener)
		}

		m.listeners = nil
	}()

	for {
		message, err := m.sub.Next(ctx)
		if err != nil {
			span.SetUnknownError(err)
			return errors.WithStack(err)
		}
		if m.host.ID() == message.GetFrom() {
			continue
		}

		signedSegment := cs.Segment{}
		err = json.Unmarshal(message.GetData(), &signedSegment)
		if err != nil {
			span.Annotate(ctx, "listen_message_err", err.Error())
			continue
		}

		m.forwardToListeners(&signedSegment)
	}
}

func (m *PubSubNetworkManager) forwardToListeners(segment *cs.Segment) {
	m.listenersMutex.RLock()
	defer m.listenersMutex.RUnlock()

	for _, listener := range m.listeners {
		go func(listener chan *cs.Segment) {
			listener <- segment
		}(listener)
	}
}

// AddListener adds a listeners for incoming links.
func (m *PubSubNetworkManager) AddListener() <-chan *cs.Segment {
	m.listenersMutex.Lock()
	defer m.listenersMutex.Unlock()

	listenChan := make(chan *cs.Segment)
	m.listeners = append(m.listeners, listenChan)

	return listenChan
}

// RemoveListener removes a listener.
func (m *PubSubNetworkManager) RemoveListener(c <-chan *cs.Segment) {
	m.listenersMutex.Lock()
	defer m.listenersMutex.Unlock()

	index := -1
	for i, l := range m.listeners {
		if l == c {
			index = i
			break
		}
	}

	if index >= 0 {
		close(m.listeners[index])
		m.listeners[index] = m.listeners[len(m.listeners)-1]
		m.listeners = m.listeners[:len(m.listeners)-1]
	}
}
