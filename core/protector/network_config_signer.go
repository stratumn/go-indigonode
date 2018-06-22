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

	"github.com/stratumn/alice/core/protector/pb"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// ConfigSigner wraps a NetworkConfig implementation and signs it
// whenever it changes.
type ConfigSigner struct {
	NetworkConfig
	privKey crypto.PrivKey
}

// WrapWithSignature wraps a NetworkConfig implementation and signs it
// whenever it changes.
func WrapWithSignature(networkConfig NetworkConfig, privKey crypto.PrivKey) NetworkConfig {
	return &ConfigSigner{
		NetworkConfig: networkConfig,
		privKey:       privKey,
	}
}

// AddPeer adds a peer to the network configuration
// and updates the signature.
func (c *ConfigSigner) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	defer log.EventBegin(ctx, "ConfigSigner.AddPeer").Done()

	err := c.NetworkConfig.AddPeer(ctx, peerID, addrs)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}

// RemovePeer removes a peer from the network configuration
// and updates the signature.
func (c *ConfigSigner) RemovePeer(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "ConfigSigner.RemovePeer").Done()

	err := c.NetworkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}

// SetNetworkState sets the current state of the network protection
// and updates the signature.
func (c *ConfigSigner) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	defer log.EventBegin(ctx, "ConfigSigner.SetNetworkState").Done()

	err := c.NetworkConfig.SetNetworkState(ctx, networkState)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
// It updates the local signature.
func (c *ConfigSigner) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	defer log.EventBegin(ctx, "ConfigSigner.Reset").Done()

	err := c.NetworkConfig.Reset(ctx, networkConfig)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}
