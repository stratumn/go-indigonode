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

package protector

import (
	"context"
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/stratumn/alice/core/protector/pb"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
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
	event := log.EventBegin(ctx, "NewInMemoryConfig")
	defer event.Done()

	err := networkConfig.ValidateContent(ctx)
	if err != nil {
		event.SetError(err)
		return nil, err
	}

	return &InMemoryConfig{
		data: deepcopy.Copy(networkConfig).(*pb.NetworkConfig),
	}, nil
}

// AddPeer adds a peer to the network configuration.
func (c *InMemoryConfig) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	defer log.EventBegin(ctx, "InMemoryConfig.AddPeer", logging.Metadata{
		"peer": peerID.Pretty(),
	}).Done()

	var marshalledAddrs []string
	for _, addr := range addrs {
		marshalledAddrs = append(marshalledAddrs, addr.String())
	}

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.data.Participants[peerID.Pretty()] = &pb.PeerAddrs{Addresses: marshalledAddrs}
	return nil
}

// RemovePeer removes a peer from the network configuration.
func (c *InMemoryConfig) RemovePeer(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "InMemoryConfig.RemovePeer", logging.Metadata{
		"peer": peerID.Pretty(),
	}).Done()

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	delete(c.data.Participants, peerID.Pretty())
	return nil
}

// IsAllowed returns true if the given peer is allowed in the network.
func (c *InMemoryConfig) IsAllowed(ctx context.Context, peerID peer.ID) bool {
	defer log.EventBegin(ctx, "InMemoryConfig.IsAllowed", logging.Metadata{
		"peer_id": peerID.Pretty(),
	}).Done()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	_, ok := c.data.Participants[peerID.Pretty()]
	return ok
}

// AllowedPeers returns the IDs of the peers in the network.
func (c *InMemoryConfig) AllowedPeers(ctx context.Context) []peer.ID {
	defer log.EventBegin(ctx, "InMemoryConfig.AllowedPeers").Done()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	var allowed []peer.ID
	for peerStr := range c.data.Participants {
		peerID, _ := peer.IDB58Decode(peerStr)
		allowed = append(allowed, peerID)
	}

	return allowed
}

// NetworkState returns the current state of the network protection.
func (c *InMemoryConfig) NetworkState(ctx context.Context) pb.NetworkState {
	event := log.EventBegin(ctx, "InMemoryConfig.NetworkState")
	defer event.Done()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	event.Append(logging.Metadata{"state": c.data.NetworkState})
	return pb.NetworkState(c.data.NetworkState)
}

// SetNetworkState sets the current state of the network protection.
func (c *InMemoryConfig) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	event := log.EventBegin(ctx, "InMemoryConfig.SetNetworkState", logging.Metadata{
		"state": networkState,
	})
	defer event.Done()

	switch networkState {
	case pb.NetworkState_BOOTSTRAP, pb.NetworkState_PROTECTED:
		c.dataLock.Lock()
		defer c.dataLock.Unlock()

		c.data.NetworkState = networkState
		return nil
	default:
		event.SetError(pb.ErrInvalidNetworkState)
		return pb.ErrInvalidNetworkState
	}
}

// Sign signs the underlying configuration.
func (c *InMemoryConfig) Sign(ctx context.Context, privKey crypto.PrivKey) error {
	event := log.EventBegin(ctx, "InMemoryConfig.Sign")
	defer event.Done()

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	err := c.data.Sign(ctx, privKey)
	if err != nil {
		event.SetError(err)
	}

	return err
}

// Copy returns a copy of the underlying configuration.
func (c *InMemoryConfig) Copy(ctx context.Context) pb.NetworkConfig {
	defer log.EventBegin(ctx, "InMemoryConfig.Copy").Done()

	c.dataLock.RLock()
	defer c.dataLock.RUnlock()

	return *deepcopy.Copy(c.data).(*pb.NetworkConfig)
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
func (c *InMemoryConfig) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	event := log.EventBegin(ctx, "InMemoryConfig.Reset")
	defer event.Done()

	err := networkConfig.ValidateContent(ctx)
	if err != nil {
		event.SetError(err)
		return err
	}

	c.dataLock.Lock()
	defer c.dataLock.Unlock()

	c.data = deepcopy.Copy(networkConfig).(*pb.NetworkConfig)
	return nil
}
