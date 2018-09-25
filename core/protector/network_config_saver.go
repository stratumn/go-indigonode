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
	"io/ioutil"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/protector/pb"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
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
