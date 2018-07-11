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
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector/pb"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// ConfigSaver wraps a NetworkConfig implementation and saves it
// to disk whenever it changes.
type ConfigSaver struct {
	NetworkConfig
	configPath string
}

// WrapWithSaver wraps a NetworkConfig implementation and saves it
// to disk whenever it changes.
func WrapWithSaver(networkConfig NetworkConfig, configPath string) NetworkConfig {
	return &ConfigSaver{
		NetworkConfig: networkConfig,
		configPath:    configPath,
	}
}

// Save saves the network configuration to disk.
func (c *ConfigSaver) Save(ctx context.Context) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_saver", "Save")
	span.AddStringAttribute("path", c.configPath)
	defer span.End()

	networkConfig := c.NetworkConfig.Copy(ctx)
	configBytes, err := json.Marshal(networkConfig)
	if err != nil {
		span.SetUnknownError(err)
		return errors.WithStack(err)
	}

	err = ioutil.WriteFile(c.configPath, configBytes, 0644)
	if err != nil {
		span.SetUnknownError(err)
		return errors.WithStack(err)
	}

	return nil
}

// AddPeer adds a peer to the network configuration
// and saves it to disk.
func (c *ConfigSaver) AddPeer(ctx context.Context, peerID peer.ID, addrs []multiaddr.Multiaddr) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_saver", "AddPeer")
	defer span.End()

	err := c.NetworkConfig.AddPeer(ctx, peerID, addrs)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}

// RemovePeer removes a peer from the network configuration
// and saves it to disk.
func (c *ConfigSaver) RemovePeer(ctx context.Context, peerID peer.ID) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_saver", "RemovePeer")
	defer span.End()

	err := c.NetworkConfig.RemovePeer(ctx, peerID)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}

// SetNetworkState sets the current state of the network protection
// and saves it to disk.
func (c *ConfigSaver) SetNetworkState(ctx context.Context, networkState pb.NetworkState) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_saver", "SetNetworkState")
	defer span.End()

	err := c.NetworkConfig.SetNetworkState(ctx, networkState)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}

// Reset clears the current configuration and applies the given one.
// It assumes that the incoming configuration signature has been validated.
// It saves it to disk.
func (c *ConfigSaver) Reset(ctx context.Context, networkConfig *pb.NetworkConfig) error {
	ctx, span := monitoring.StartSpan(ctx, "protector.config_saver", "Reset")
	defer span.End()

	err := c.NetworkConfig.Reset(ctx, networkConfig)
	if err != nil {
		return err
	}

	return c.Save(ctx)
}
