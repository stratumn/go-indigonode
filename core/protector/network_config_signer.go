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

	pb "github.com/stratumn/alice/pb/protector"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// ConfigSigner wraps a NetworkConfig implementation and signs it
// whenever it changes.
type ConfigSigner struct {
	networkConfig NetworkConfig
	privKey       crypto.PrivKey
}

// WrapWithSignature wraps a NetworkConfig implementation and signs it
// whenever it changes.
func WrapWithSignature(networkConfig NetworkConfig, privKey crypto.PrivKey) NetworkConfig {
	return &ConfigSigner{
		networkConfig: networkConfig,
		privKey:       privKey,
	}
}

// AddPeer adds a peer to the network configuration
// and updates the signature.
func (c *ConfigSigner) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	defer log.EventBegin(ctx, "ConfigSigner.AddPeer").Done()

	err := c.networkConfig.AddPeer(ctx, peerID, addrs)
	if err != nil {
		return err
	}

	return c.networkConfig.Sign(ctx, c.privKey)
}

// RemovePeer removes a peer from the network configuration
// and updates the signature.
func (c *ConfigSigner) RemovePeer(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "ConfigSigner.RemovePeer").Done()

	err := c.networkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	return c.networkConfig.Sign(ctx, c.privKey)
}

// IsAllowed returns true if the given peer is allowed in the network.
func (c *ConfigSigner) IsAllowed(ctx context.Context, peerID peer.ID) bool {
	return c.networkConfig.IsAllowed(ctx, peerID)
}

// AllowedPeers returns the IDs of the peers in the network.
func (c *ConfigSigner) AllowedPeers(ctx context.Context) []peer.ID {
	return c.networkConfig.AllowedPeers(ctx)
}

// NetworkState returns the current state of the network protection.
func (c *ConfigSigner) NetworkState(ctx context.Context) pb.NetworkState {
	return c.networkConfig.NetworkState(ctx)
}

// SetNetworkState sets the current state of the network protection
// and updates the signature.
func (c *ConfigSigner) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	defer log.EventBegin(ctx, "ConfigSigner.SetNetworkState").Done()

	err := c.networkConfig.SetNetworkState(ctx, networkState)
	if err != nil {
		return err
	}

	return c.networkConfig.Sign(ctx, c.privKey)
}

// Sign signs the underlying configuration.
func (c *ConfigSigner) Sign(ctx context.Context, privKey crypto.PrivKey) error {
	return c.networkConfig.Sign(ctx, privKey)
}

// Copy returns a copy of the underlying configuration.
func (c *ConfigSigner) Copy(ctx context.Context) pb.NetworkConfig {
	return c.networkConfig.Copy(ctx)
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
// It updates the local signature.
func (c *ConfigSigner) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	defer log.EventBegin(ctx, "ConfigSigner.Reset").Done()

	err := c.networkConfig.Reset(ctx, networkConfig)
	if err != nil {
		return err
	}

	return c.networkConfig.Sign(ctx, c.privKey)
}
