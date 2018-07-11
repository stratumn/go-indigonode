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

package store

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector"
	protectorpb "github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/core/streamutil"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

var (
	// IndigoLinkProtocolID is the protocol used to share a new link between
	// network participants.
	IndigoLinkProtocolID = protocol.ID("/indigo/node/indigo/store/newlink/v1.0.0")
)

// PrivateNetworkManager implements the NetworkManager interface.
// It directly connect to all participants of a private network.
type PrivateNetworkManager struct {
	host           Host
	networkCfg     protector.NetworkConfigReader
	streamProvider streamutil.Provider

	listenersMutex *sync.RWMutex
	listeners      []chan *cs.Segment
}

// NewPrivateNetworkManager creates a new NetworkManager in a private network.
// It assumes that all network participants are interested in Indigo links
// and will connect to all participants to exchange links.
func NewPrivateNetworkManager(streamProvider streamutil.Provider, networkCfg protector.NetworkConfigReader) NetworkManager {
	return &PrivateNetworkManager{
		networkCfg:     networkCfg,
		streamProvider: streamProvider,
		listenersMutex: &sync.RWMutex{},
	}
}

// NodeID returns the ID of the node in the network.
func (m PrivateNetworkManager) NodeID() peer.ID {
	return m.host.ID()
}

// Join assumes that the whole private network is part of the PoP network.
// It configures the host to accept Indigo message protocols.
func (m *PrivateNetworkManager) Join(ctx context.Context, _ string, host Host) error {
	defer log.EventBegin(ctx, "Join").Done()

	m.host = host
	return nil
}

// Leave assumes that the whole private network is part of the PoP network.
// It removes Indigo-specific configuration from the host.
func (m *PrivateNetworkManager) Leave(ctx context.Context, _ string) error {
	defer log.EventBegin(ctx, "Leave").Done()

	m.host = nil
	return nil
}

// Publish sends a message to all the network.
func (m *PrivateNetworkManager) Publish(ctx context.Context, link *cs.Link) error {
	event := log.EventBegin(ctx, "Publish")
	defer event.Done()

	if m.networkCfg.NetworkState(ctx) != protectorpb.NetworkState_PROTECTED {
		event.SetError(ErrNetworkNotReady)
		return ErrNetworkNotReady
	}

	key := m.host.Peerstore().PrivKey(m.host.ID())
	signedSegment, err := audit.SignLink(ctx, key, link)
	if err != nil {
		event.SetError(err)
		return err
	}

	encodedSegment, _ := pb.FromSegment(signedSegment)

	wg := &sync.WaitGroup{}
	for _, peerID := range m.networkCfg.AllowedPeers(ctx) {
		if peerID == m.host.ID() {
			continue
		}

		wg.Add(1)
		go func(peerID peer.ID) {
			event := log.EventBegin(ctx, "PublishToPeer", peerID)
			defer event.Done()
			defer wg.Done()

			stream, err := m.streamProvider.NewStream(ctx,
				m.host,
				streamutil.OptPeerID(peerID),
				streamutil.OptProtocolIDs(IndigoLinkProtocolID))
			if err != nil {
				return
			}

			defer stream.Close()

			err = stream.Codec().Encode(encodedSegment)
			if err != nil {
				event.SetError(err)
			}
		}(peerID)
	}

	wg.Wait()
	return nil
}

// Listen to messages from the network.
// Cancel the context to stop listening.
func (m *PrivateNetworkManager) Listen(ctx context.Context) error {
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

	m.host.SetStreamHandler(
		IndigoLinkProtocolID,
		streamutil.WithAutoClose("indigo.store", "Listen", m.HandleNewLink),
	)

	<-ctx.Done()

	m.host.RemoveStreamHandler(IndigoLinkProtocolID)
	return ctx.Err()
}

// HandleNewLink handles an incoming link.
// If the message can be decoded correctly, it's forwarded to listeners.
func (m *PrivateNetworkManager) HandleNewLink(
	ctx context.Context,
	span *monitoring.Span,
	stream inet.Stream,
	codec streamutil.Codec,
) error {
	var encodedSegment pb.Segment
	err := codec.Decode(&encodedSegment)
	if err != nil {
		return errors.WithStack(err)
	}

	signedSegment, err := encodedSegment.ToSegment()
	if err != nil {
		return err
	}

	m.listenersMutex.RLock()
	defer m.listenersMutex.RUnlock()

	for _, listener := range m.listeners {
		go func(listener chan *cs.Segment) {
			event := log.EventBegin(ctx, "ForwardingToListener")
			listener <- signedSegment
			event.Done()
		}(listener)
	}

	return nil
}

// AddListener adds a listeners for incoming segments.
func (m *PrivateNetworkManager) AddListener() <-chan *cs.Segment {
	m.listenersMutex.Lock()
	defer m.listenersMutex.Unlock()

	listenChan := make(chan *cs.Segment)
	m.listeners = append(m.listeners, listenChan)

	return listenChan
}

// RemoveListener removes a listener.
func (m *PrivateNetworkManager) RemoveListener(c <-chan *cs.Segment) {
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
