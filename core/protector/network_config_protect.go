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

	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector/pb"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

// ProtectUpdater wraps a NetworkConfig implementation and updates a
// protector when the configuration changes.
type ProtectUpdater struct {
	NetworkConfig
	peerStore   peerstore.Peerstore
	protect     Protector
	protectChan chan NetworkUpdate
}

// WrapWithProtectUpdater wraps a NetworkConfig implementation and updates
// the given protector when the configuration changes.
func WrapWithProtectUpdater(
	networkConfig NetworkConfig,
	protect Protector,
	peerStore peerstore.Peerstore,
) NetworkConfig {
	conf := ProtectUpdater{
		NetworkConfig: networkConfig,
		peerStore:     peerStore,
		protect:       protect,
		protectChan:   make(chan NetworkUpdate),
	}

	// This go routine has the same lifetime as the ProtectUpdater object,
	// so it makes sense to launch it here. When the ProtectUpdater object
	// is collected by the GC, the channel is closed which stops
	// this go routine.
	go protect.ListenForUpdates(conf.protectChan)

	return &conf
}

// AddPeer adds a peer to the network configuration
// and updates the protector and peer store.
func (c *ProtectUpdater) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_protect", "AddPeer")
	defer span.End()

	err := c.NetworkConfig.AddPeer(ctx, peerID, addrs)
	if err != nil {
		return err
	}

	c.peerStore.AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)
	c.protectChan <- CreateAddNetworkUpdate(peerID)

	return nil
}

// RemovePeer removes a peer from the network configuration
// and updates the protector.
func (c *ProtectUpdater) RemovePeer(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_protect", "RemovePeer")
	defer span.End()

	err := c.NetworkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	c.protectChan <- CreateRemoveNetworkUpdate(peerID)

	return nil
}

// SetNetworkState sets the current state of the network protection
// and updates the protector if it's interested in state changes.
func (c *ProtectUpdater) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_protect", "SetNetworkState")
	defer span.End()

	err := c.NetworkConfig.SetNetworkState(ctx, networkState)
	if err != nil {
		return err
	}

	stateAwareProtector, ok := c.protect.(NetworkStateWriter)
	if ok {
		err := stateAwareProtector.SetNetworkState(ctx, networkState)
		if err != nil {
			span.SetUnknownError(err)
			return err
		}
	}

	return nil
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
// It updates the protector accordingly.
func (c *ProtectUpdater) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_protect", "Reset")
	defer span.End()

	err := c.NetworkConfig.Reset(ctx, networkConfig)
	if err != nil {
		return err
	}

	stateAwareProtector, ok := c.protect.(NetworkStateWriter)
	if ok {
		err := stateAwareProtector.SetNetworkState(ctx, networkConfig.NetworkState)
		if err != nil {
			span.SetUnknownError(err)
			return err
		}
	}

	for _, peerID := range c.protect.AllowedPeers(ctx) {
		c.protectChan <- CreateRemoveNetworkUpdate(peerID)
	}

	for _, peerID := range c.AllowedPeers(ctx) {
		c.protectChan <- CreateAddNetworkUpdate(peerID)
	}

	return nil
}
