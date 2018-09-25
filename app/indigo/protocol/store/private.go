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
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector"
	protectorpb "github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/core/streamutil"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
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
	_, span := monitoring.StartSpan(ctx, "indigo.store", "Join")
	defer span.End()

	m.host = host
	return nil
}

// Leave assumes that the whole private network is part of the PoP network.
// It removes Indigo-specific configuration from the host.
func (m *PrivateNetworkManager) Leave(ctx context.Context, _ string) error {
	_, span := monitoring.StartSpan(ctx, "indigo.store", "Leave")
	defer span.End()

	m.host = nil
	return nil
}

// Publish sends a message to all the network.
func (m *PrivateNetworkManager) Publish(ctx context.Context, link *cs.Link) error {
	ctx, span := monitoring.StartSpan(ctx, "indigo.store", "Publish")
	defer span.End()

	if m.networkCfg.NetworkState(ctx) != protectorpb.NetworkState_PROTECTED {
		span.SetUnknownError(ErrNetworkNotReady)
		return ErrNetworkNotReady
	}

	key := m.host.Peerstore().PrivKey(m.host.ID())
	signedSegment, err := audit.SignLink(ctx, key, link)
	if err != nil {
		span.SetUnknownError(err)
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
			ctx, span := monitoring.StartSpan(ctx, "indigo.store", "PublishToPeer", monitoring.SpanOptionPeerID(peerID))
			defer span.End()
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
				span.SetUnknownError(err)
			}
		}(peerID)
	}

	wg.Wait()
	return nil
}

// Listen to messages from the network.
// Cancel the context to stop listening.
func (m *PrivateNetworkManager) Listen(ctx context.Context) error {
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
			listener <- signedSegment
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
