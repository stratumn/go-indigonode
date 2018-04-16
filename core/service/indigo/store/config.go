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

package store

import (
	"github.com/pkg/errors"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

// Config contains configuration options for the Indigo service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Version is the version of the Indigo service.
	Version string `toml:"version" comment:"The version of the indigo service."`

	// NetworkId is the id of your Indigo PoP network.
	NetworkID string `toml:"network_id" comment:"The id of your Indigo PoP network."`

	// PrivateKey is the private key of the node.
	PrivateKey string `toml:"private_key" comment:"The private key of the host."`
}

// UnmarshalPrivateKey unmarshals the configured private key.
func (c *Config) UnmarshalPrivateKey() (ic.PrivKey, error) {
	keyBytes, err := ic.ConfigDecodeKey(c.PrivateKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	key, err := ic.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return key, nil
}
