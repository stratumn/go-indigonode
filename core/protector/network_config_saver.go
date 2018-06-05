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
	"io/ioutil"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/pb/protector"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// ConfigSaver wraps a NetworkConfig implementation and saves it
// to disk whenever it changes.
type ConfigSaver struct {
	networkConfig NetworkConfig
	configPath    string
}

// WrapWithSaver wraps a NetworkConfig implementation and saves it
// to disk whenever it changes.
func WrapWithSaver(networkConfig NetworkConfig, configPath string) NetworkConfig {
	return &ConfigSaver{
		networkConfig: networkConfig,
		configPath:    configPath,
	}
}

// Save saves the network configuration to disk.
func (c *ConfigSaver) Save(ctx context.Context) error {
	event := log.EventBegin(ctx, "ConfigSaver.Save", logging.Metadata{"path": c.configPath})
	defer event.Done()

	networkConfig := c.networkConfig.Copy(ctx)
	configBytes, err := json.Marshal(networkConfig)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	err = ioutil.WriteFile(c.configPath, configBytes, 0644)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	return nil
}

// AddPeer adds a peer to the network configuration
// and saves it to disk.
func (c *ConfigSaver) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	defer log.EventBegin(ctx, "ConfigSaver.AddPeer").Done()

	err := c.networkConfig.AddPeer(ctx, peerID, addrs)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}

// RemovePeer removes a peer from the network configuration
// and saves it to disk.
func (c *ConfigSaver) RemovePeer(ctx context.Context, peerID peer.ID) error {
	defer log.EventBegin(ctx, "ConfigSaver.RemovePeer").Done()

	err := c.networkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}

// IsAllowed returns true if the given peer is allowed in the network.
func (c *ConfigSaver) IsAllowed(ctx context.Context, peerID peer.ID) bool {
	return c.networkConfig.IsAllowed(ctx, peerID)
}

// AllowedPeers returns the IDs of the peers in the network.
func (c *ConfigSaver) AllowedPeers(ctx context.Context) []peer.ID {
	return c.networkConfig.AllowedPeers(ctx)
}

// NetworkState returns the current state of the network protection.
func (c *ConfigSaver) NetworkState(ctx context.Context) pb.NetworkState {
	return c.networkConfig.NetworkState(ctx)
}

// SetNetworkState sets the current state of the network protection
// and saves it to disk.
func (c *ConfigSaver) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	defer log.EventBegin(ctx, "ConfigSaver.SetNetworkState").Done()

	err := c.networkConfig.SetNetworkState(ctx, networkState)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}

// Sign signs the underlying configuration.
func (c *ConfigSaver) Sign(ctx context.Context, privKey crypto.PrivKey) error {
	return c.networkConfig.Sign(ctx, privKey)
}

// Copy returns a copy of the underlying configuration.
func (c *ConfigSaver) Copy(ctx context.Context) pb.NetworkConfig {
	return c.networkConfig.Copy(ctx)
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
// It saves it to disk.
func (c *ConfigSaver) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	defer log.EventBegin(ctx, "ConfigSaver.Reset").Done()

	err := c.networkConfig.Reset(ctx, networkConfig)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}
