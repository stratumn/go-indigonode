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

package swarm

import (
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protector"
)

// Configuration errors.
var (
	ErrInvalidProtectionMode    = errors.New("invalid network protection mode")
	ErrInvalidCoordinatorConfig = errors.New("invalid network coordinator config")
)

// Config contains configuration options for the Swarm service.
type Config struct {
	// PeerID is the peer ID of the node.
	PeerID string `toml:"peer_id" comment:"The peer ID of the host."`

	// PrivateKey is the private key of the node.
	PrivateKey string `toml:"private_key" comment:"The private key of the host."`

	// Addresses are the list of addresses to bind to.
	Addresses []string `toml:"addresses" comment:"List of addresses to bind to."`

	// StreamMuxer is the name of the stream muxer service.
	StreamMuxer string `toml:"stream_muxer" comment:"The name of the stream muxer service."`

	// Metrics is the name of the metrics service.
	Metrics string `toml:"metrics" comment:"The name of the metrics service (blank = disabled)."`

	// ProtectionMode describes the network protection mode.
	ProtectionMode string `toml:"protection_mode" comment:"Protection mode for private network (blank = disabled)."`

	// CoordinatorConfig contains configuration options for the network coordinator (if enabled).
	CoordinatorConfig *CoordinatorConfig `toml:"coordinator" comment:"Configure settings for the network coordinator in the following section."`
}

// CoordinatorConfig contains configuration options
// for the network coordinator (if enabled).
type CoordinatorConfig struct {
	// IsCoordinator is true when the node is the coordinator of a private network.
	IsCoordinator bool `toml:"is_coordinator" comment:"True if we are the coordinator node."`

	// CoordinatorID is the peer ID of the coordinator node.
	CoordinatorID string `toml:"coordinator_id" comment:"The peer ID of the coordinator."`

	// CoordinatorAddresses is the list of addresses of the coordinator node.
	CoordinatorAddresses []string `toml:"coordinator_addresses" comment:"Coordinator addresses."`

	// Path to the signed network configuration file.
	ConfigPath string `toml:"config_path" comment:"Path to the signed network configuration file."`
}

// ParseNetworkMode parses network mode details from the configuration.
func (c Config) ParseNetworkMode() (*protector.NetworkMode, error) {
	switch c.ProtectionMode {
	case "":
		return nil, nil
	case protector.PrivateWithCoordinatorMode:
		if c.CoordinatorConfig == nil {
			return nil, ErrInvalidCoordinatorConfig
		}

		if c.CoordinatorConfig.ConfigPath == "" {
			return nil, ErrInvalidCoordinatorConfig
		}

		if c.CoordinatorConfig.IsCoordinator {
			return protector.NewCoordinatorNetworkMode(), nil
		}

		return protector.NewCoordinatedNetworkMode(
			c.CoordinatorConfig.CoordinatorID,
			c.CoordinatorConfig.CoordinatorAddresses,
		)
	default:
		return nil, ErrInvalidProtectionMode
	}
}
