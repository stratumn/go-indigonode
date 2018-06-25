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

package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
)

// Configuration errors.
var (
	ErrNotHost          = errors.New("connected service is not a host")
	ErrNotSwarm         = errors.New("connected service is not a swarm")
	ErrInvalidStoreType = errors.New("invalid store type")
)

// Default config values.
const (
	DefaultStorePath = "data/network/proposals.json"
)

// Config contains configuration options for the Bootstrap service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Swarm is the name of the swarm service.
	Swarm string `toml:"swarm" comment:"The name of the swarm service."`

	// Needs are services that should be started in addition to the host
	// before bootstrapping.
	Needs []string `toml:"needs" comment:"Services that should be started in addition to the host before bootstrapping."`

	// Addresses is a list of known peer addresses.
	Addresses []string `toml:"addresses" comment:"A list of known peer addresses."`

	// MinPeerThreshold is the number of peers under which to bootstrap
	// connections.
	MinPeerThreshold int `toml:"min_peer_threshold" comment:"The number of peers under which to bootstrap connections."`

	// Interval is the duration of the interval between bootstrap jobs.
	Interval string `toml:"interval" comment:"Interval between bootstrap jobs."`

	// ConnectionTimeout is the connection timeout. It should be less than
	// the interval.
	ConnectionTimeout string `toml:"connection_timeout" comment:"The connection timeout. It should be less than the interval."`

	// StoreConfig configures the store used for network update proposals.
	StoreConfig *StoreConfig `toml:"store_config" comment:"Configure the store used for network update proposals."`
}

// Types of store supported.
const (
	InMemoryStore = "in-memory"
	FileStore     = "file"
)

// StoreConfig configures the store used for network update proposals.
type StoreConfig struct {
	// Type is the type of store used.
	Type string `toml:"type" comment:"Type of store to use.\n Supported values: in-memory and file."`

	// Path to the store files (when applicable).
	Path string `toml:"path" comment:"Path to the store files (when applicable)."`
}

// NewStore configures a store for network update proposals.
func (c *Config) NewStore(ctx context.Context) (proposal.Store, error) {
	if c.StoreConfig == nil {
		return proposal.NewInMemoryStore(), nil
	}

	switch c.StoreConfig.Type {
	case InMemoryStore:
		return proposal.NewInMemoryStore(), nil
	case FileStore:
		path := c.StoreConfig.Path
		if len(path) == 0 {
			path = DefaultStorePath
		}

		return proposal.WrapWithSaver(ctx, proposal.NewInMemoryStore(), path)
	default:
		return nil, ErrInvalidStoreType
	}
}
