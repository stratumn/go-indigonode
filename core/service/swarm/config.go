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

import "github.com/pkg/errors"

// Supported network protection modes.
const (
	// PrivateCoordinatorMode uses a coordinator node
	// for network participants updates.
	PrivateCoordinatorMode = "private-coordinator"
)

// Configuration errors.
var (
	ErrInvalidProtectionMode = errors.New("invalid network protection mode")
)

// Config contains configuration options for the Swarm service.
type Config struct {
	// PeerID is peer ID of the node.
	PeerID string `toml:"peer_id" comment:"The peer ID of the host."`

	// PrivateKey is the private key of the node.
	PrivateKey string `toml:"private_key" comment:"The private key of the host."`

	// Addresses are the list of addresses to bind to.
	Addresses []string `toml:"addresses" comment:"List of addresses to bind to."`

	// ProtectionMode describes the network protection mode.
	ProtectionMode string `toml:"protection_mode" comment:"Protection mode for private network (blank = disabled)."`

	// StreamMuxer is the name of the stream muxer service.
	StreamMuxer string `toml:"stream_muxer" comment:"The name of the stream muxer service."`

	// Metrics is the name of the metrics service.
	Metrics string `toml:"metrics" comment:"The name of the metrics service (blank = disabled)."`
}

// ValidateProtectionMode checks that the protection mode is supported.
func (c *Config) ValidateProtectionMode() error {
	if c.ProtectionMode != "" && c.ProtectionMode != PrivateCoordinatorMode {
		return ErrInvalidProtectionMode
	}

	return nil
}
