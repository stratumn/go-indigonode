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

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// Supported network protection modes.
const (
	// PrivateWithCoordinatorMode uses a coordinator node
	// for network participants updates.
	PrivateWithCoordinatorMode = "private-with-coordinator"
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
	// CoordinatorID is the peer ID of the coordinator node.
	CoordinatorID string `toml:"coordinator_id" comment:"The peer ID of the coordinator."`

	// CoordinatorAddresses is the list of addresses of the coordinator node.
	CoordinatorAddresses []string `toml:"coordinator_addresses" comment:"Coordinator addresses."`
}

// ValidateProtectionMode checks that the protection mode is supported.
func (c Config) ValidateProtectionMode() error {
	if c.ProtectionMode == "" {
		return nil
	}

	protectionModes := map[string]func() error{
		PrivateWithCoordinatorMode: c.validateCoordinatorConfig,
	}

	validator, ok := protectionModes[c.ProtectionMode]
	if !ok {
		return ErrInvalidProtectionMode
	}

	return validator()
}

func (c Config) validateCoordinatorConfig() error {
	if c.CoordinatorConfig == nil {
		return ErrInvalidCoordinatorConfig
	}

	_, err := peer.IDB58Decode(c.CoordinatorConfig.CoordinatorID)
	if err != nil {
		return ErrInvalidCoordinatorConfig
	}

	if len(c.CoordinatorConfig.CoordinatorAddresses) == 0 {
		return ErrInvalidCoordinatorConfig
	}

	for _, addr := range c.CoordinatorConfig.CoordinatorAddresses {
		if _, err := multiaddr.NewMultiaddr(addr); err != nil {
			return ErrInvalidCoordinatorConfig
		}
	}

	return nil
}
