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

	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector/pb"

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
	ctx, span := monitoring.StartSpan(ctx, "protector.config_signer", "AddPeer")
	defer span.End()

	err := c.NetworkConfig.AddPeer(ctx, peerID, addrs)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}

// RemovePeer removes a peer from the network configuration
// and updates the signature.
func (c *ConfigSigner) RemovePeer(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_signer", "RemovePeer")
	defer span.End()

	err := c.NetworkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}

// SetNetworkState sets the current state of the network protection
// and updates the signature.
func (c *ConfigSigner) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_signer", "SetNetworkState")
	defer span.End()

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
	ctx, span := monitoring.StartSpan(ctx, "protector.config_signer", "Reset")
	defer span.End()

	err := c.NetworkConfig.Reset(ctx, networkConfig)
	if err != nil {
		return err
	}

	return c.NetworkConfig.Sign(ctx, c.privKey)
}
