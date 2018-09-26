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

package protector

import (
	"context"
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/stratumn/go-node/core/monitoring"
	"github.com/stratumn/go-node/core/protector/pb"

	"gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

// InMemoryConfig implements the NetworkConfig interface.
// It only keeps the configuration in RAM.
// It should be wrapped to add more features (e.g. signing the config
// when changes happen, configuring the underlying protector,
// saving the configuration to a file or database, etc).
type InMemoryConfig struct {
	dataLock sync.RWMutex
	data     *pb.NetworkConfig
}

// NewInMemoryConfig creates a thread-safe NetworkConfig.
// It's the source of truth for the network configuration and
// should be the only object mutating the underlying data.
func NewInMemoryConfig(ctx context.Context, networkConfig *pb.NetworkConfig) (NetworkConfig, error) {
	ctx, span := monitoring.StartSpan(ctx, "protector.memory_config", "NewInMemoryConfig")
	defer span.End()

	err := networkConfig.ValidateContent(ctx)
	if err != nil {
		span.SetUnknownError(err)
		return nil, err
	}

	return &InMemoryConfig{
		data: deepcopy.Copy(networkConfig).(*pb.NetworkConfig),
	}, nil
}

// AddPeer adds a peer to the network configuration.
func (c *InMemoryConfig) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "AddPeer", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	// We don't want to put localhost addresses in the network configuration.
	// It wouldn't make any sense for external nodes.
	localAddrs := map[string]struct{}{}
	localAddrs["127.0.0.1"] = struct{}{}
	localAddrs["0.0.0.0"] = struct{}{}
	localAddrs["::1"] = struct{}{}
	localAddrs["::"] = struct{}{}

	var marshalledAddrs []string
	for _, addr := range addrs {
		ip4, err := addr.ValueForProtocol(multiaddr.P_IP4)
		if err == nil {
			_, ok := localAddrs[ip4]
			if ok {
				continue
			}
		}

		ip6, err := addr.ValueForProtocol(multiaddr.P_IP6)
		if err == nil {
			_, ok := localAddrs[ip6]
			if ok {
				continue
			}
		}

		marshalledAddrs = append(marshalledAddrs, addr.String())
	}

	if len(marshalledAddrs) == 0 {
		return ErrMissingNonLocalAddr
	}

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.data.Participants[peerID.Pretty()] = &pb.PeerAddrs{Addresses: marshalledAddrs}
	return nil
}

// RemovePeer removes a peer from the network configuration.
func (c *InMemoryConfig) RemovePeer(ctx context.Context, peerID peer.ID) error {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "RemovePeer", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	delete(c.data.Participants, peerID.Pretty())
	return nil
}

// IsAllowed returns true if the given peer is allowed in the network.
func (c *InMemoryConfig) IsAllowed(ctx context.Context, peerID peer.ID) bool {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "IsAllowed", monitoring.SpanOptionPeerID(peerID))
	defer span.End()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	_, ok := c.data.Participants[peerID.Pretty()]
	return ok
}

// AllowedPeers returns the IDs of the peers in the network.
func (c *InMemoryConfig) AllowedPeers(ctx context.Context) []peer.ID {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "AllowedPeers")
	defer span.End()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	var allowed []peer.ID
	for peerStr := range c.data.Participants {
		peerID, _ := peer.IDB58Decode(peerStr)
		allowed = append(allowed, peerID)
	}

	return allowed
}

// AllowedAddrs returns the whitelisted addresses of the given peer.
func (c *InMemoryConfig) AllowedAddrs(ctx context.Context, peerID peer.ID) []multiaddr.Multiaddr {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "AllowedAddrs")
	defer span.End()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	var allowed []multiaddr.Multiaddr
	for peerStr, peerAddrs := range c.data.Participants {
		allowedPeerID, _ := peer.IDB58Decode(peerStr)
		if allowedPeerID == peerID {
			for _, addrStr := range peerAddrs.Addresses {
				addr, _ := multiaddr.NewMultiaddr(addrStr)
				allowed = append(allowed, addr)
			}

			break
		}
	}

	return allowed
}

// NetworkState returns the current state of the network protection.
func (c *InMemoryConfig) NetworkState(ctx context.Context) pb.NetworkState {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "NetworkState")
	defer span.End()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	span.AddStringAttribute("state", c.data.NetworkState.String())
	return pb.NetworkState(c.data.NetworkState)
}

// SetNetworkState sets the current state of the network protection.
func (c *InMemoryConfig) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "SetNetworkState")
	defer span.End()

	span.AddStringAttribute("state", networkState.String())

	switch networkState {
	case pb.NetworkState_BOOTSTRAP, pb.NetworkState_PROTECTED:
		c.dataLock.Lock()
		defer c.dataLock.Unlock()

		c.data.NetworkState = networkState
		return nil
	default:
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeInvalidArgument, pb.ErrInvalidNetworkState.Error()))
		return pb.ErrInvalidNetworkState
	}
}

// Sign signs the underlying configuration.
func (c *InMemoryConfig) Sign(ctx context.Context, privKey crypto.PrivKey) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.memory_config", "Sign")
	defer span.End()

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	err := c.data.Sign(ctx, privKey)
	if err != nil {
		span.SetUnknownError(err)
	}

	return err
}

// Copy returns a copy of the underlying configuration.
func (c *InMemoryConfig) Copy(ctx context.Context) pb.NetworkConfig {
	_, span := monitoring.StartSpan(ctx, "protector.memory_config", "Copy")
	defer span.End()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	return *deepcopy.Copy(c.data).(*pb.NetworkConfig)
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
func (c *InMemoryConfig) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.memory_config", "Reset")
	defer span.End()

	err := networkConfig.ValidateContent(ctx)
	if err != nil {
		span.SetUnknownError(err)
		return err
	}

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	if err := c.validateLastUpdated(networkConfig); err != nil {
		span.SetStatus(monitoring.NewStatus(monitoring.StatusCodeFailedPrecondition, pb.ErrInvalidLastUpdated.Error()))
		return pb.ErrInvalidLastUpdated
	}

	c.data = deepcopy.Copy(networkConfig).(*pb.NetworkConfig)
	return nil
}

// validateLastUpdated validates the timestamp of an incoming signed network
// configuration.
// It expects that the data lock is held by the caller.
func (c *InMemoryConfig) validateLastUpdated(networkConfig *pb.NetworkConfig) error {
	if networkConfig.LastUpdated == nil {
		return pb.ErrMissingLastUpdated
	}

	// If we don't have a timespamped configuration, we accept the incoming
	// configuration.
	if c.data.LastUpdated == nil {
		return nil
	}

	// If we don't have network participants yet, we accept the incoming
	// configuration.
	if len(c.data.Participants) < 2 {
		return nil
	}

	// Otherwise we check that we don't have a more recent one.
	if networkConfig.LastUpdated.Seconds < c.data.LastUpdated.Seconds {
		return pb.ErrInvalidLastUpdated
	}

	return nil
}
