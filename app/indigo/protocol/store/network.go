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

//go:generate mockgen -package mocknetworkmanager -destination mocknetwork/mocknetwork.go github.com/stratumn/alice/app/indigo/protocol/store NetworkManager

package store

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigocore/cs"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	floodsub "gx/ipfs/QmVKrsEgixRtMWcMd6WQzuwqCUC3jfLf7Q7xcjnKoMMikS/go-libp2p-floodsub"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// ErrInvalidNetworkID is returned at startup when the network is misconfigured.
	ErrInvalidNetworkID = errors.New("invalid or missing network ID")
)

// Host represents an Alice host.
type Host = ihost.Host

// NetworkManager provides methods to manage and join PoP networks.
type NetworkManager interface {
	// NodeID returns the ID of the node in the network.
	NodeID() peer.ID

	// Join joins a PoP network.
	Join(ctx context.Context, networkID string, host Host) error
	// Leave leaves a PoP network.
	Leave(ctx context.Context, networkID string) error

	// Publish sends a message to all the network.
	Publish(ctx context.Context, link *cs.Link) error
	// Listen to messages from the network. Cancel the context
	// to stop listening.
	Listen(ctx context.Context) error

	// AddListener adds a listeners for incoming segments.
	AddListener() <-chan *cs.Segment
	// RemoveListener removes a listener.
	RemoveListener(<-chan *cs.Segment)
}

// PubSubNetworkManager implements the NetworkManager interface.
type PubSubNetworkManager struct {
	peerKey ic.PrivKey

	networkMutex sync.Mutex
	networkID    string
	host         Host
	pubsub       *floodsub.PubSub
	sub          *floodsub.Subscription

	listenersMutex sync.RWMutex
	listeners      []chan *cs.Segment
}

// NewNetworkManager creates a new NetworkManager.
func NewNetworkManager(peerKey ic.PrivKey) NetworkManager {
	return &PubSubNetworkManager{peerKey: peerKey}
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
	event := log.EventBegin(ctx, "Join", logging.Metadata{"network_id": networkID})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

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
	event := log.EventBegin(ctx, "Leave", logging.Metadata{"network_id": networkID})
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

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
	event := log.EventBegin(ctx, "Publish")
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Done()
	}()

	var pubsubPeers []string
	for _, peer := range m.pubsub.ListPeers(m.networkID) {
		pubsubPeers = append(pubsubPeers, peer.Pretty())
	}
	event.Append(logging.Metadata{"peers": strings.Join(pubsubPeers, ",")})

	signedSegment, err := audit.SignLink(ctx, m.peerKey, link)
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
	event := log.EventBegin(ctx, "Listen")
	defer event.Done()

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
			event.SetError(err)
			return errors.WithStack(err)
		}
		if m.host.ID() == message.GetFrom() {
			continue
		}

		signedSegment := cs.Segment{}
		err = json.Unmarshal(message.GetData(), &signedSegment)
		if err != nil {
			log.Event(ctx, "ListenMessageError", logging.Metadata{"err": err.Error()})
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
			ctx := context.Background()
			event := log.EventBegin(ctx, "ForwardingToListener")
			listener <- segment
			event.Done()
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
