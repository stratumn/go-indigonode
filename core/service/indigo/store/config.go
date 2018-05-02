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
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit"
	"github.com/stratumn/alice/core/protocol/indigo/store/audit/dummyauditstore"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/postgresstore"
	indigostore "github.com/stratumn/go-indigocore/store"

	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

const (
	// InMemoryStorage is the StorageType value for storing in memory (no hard storage).
	InMemoryStorage = "in-memory"
	// PostgreSQLStorage is the StorageType value for storing in an external PostgreSQL database.
	PostgreSQLStorage = "postgresql"

	// duplicateTable is the error returned by pq when trying to create an existing table.
	duplicateTable = pq.ErrorCode("42P07")
)

var (
	// ErrStorageNotSupported is returned when the configured storage type and url are invalid.
	ErrStorageNotSupported = errors.New("invalid or unsupported storage")

	// ErrMissingConfig is returned when the provided configuration is missing required settings.
	ErrMissingConfig = errors.New("missing configuration settings")
)

// PostgresConfig contains configuration for the postgres indigo store.
type PostgresConfig struct {
	// StorageDbURL is the URL of the storage database (if external storage is chosen).
	StorageDbURL string `toml:"storage_db_url" comment:"If external storage is used, the url of that storage."`
}

// Config contains configuration options for the Indigo service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Version is the version of the Indigo Store service.
	Version string `toml:"version" comment:"The version of the indigo service."`

	// NetworkId is the id of your Indigo PoP network.
	NetworkID string `toml:"network_id" comment:"The id of your Indigo PoP network."`

	// PrivateKey is the private key of the node.
	PrivateKey string `toml:"private_key" comment:"The private key of the host."`

	// StoreType is the type of storage used.
	StorageType string `toml:"storage_type" comment:"The type of storage to us.\n Supported values: in-memory and postgresql."`

	// PostgresConfig contains configuration for the postgres indigo store.
	PostgresConfig *PostgresConfig `toml:"postgres" comment:"Configure settings for the Indigo PostgreSQL Store in the following section."`
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

// CreateAuditStore creates an audit store from the configuration.
func (c *Config) CreateAuditStore() (audit.Store, error) {
	switch c.StorageType {
	case InMemoryStorage:
		return dummyauditstore.New(), nil
	default:
		return nil, ErrStorageNotSupported
	}
}

// CreateIndigoStore creates an indigo store from the configuration.
func (c *Config) CreateIndigoStore() (indigostore.Adapter, error) {
	switch c.StorageType {
	case InMemoryStorage:
		return dummystore.New(&dummystore.Config{}), nil
	case PostgreSQLStorage:
		if c.PostgresConfig == nil {
			return nil, ErrMissingConfig
		}
		p, err := postgresstore.New(&postgresstore.Config{
			URL: c.PostgresConfig.StorageDbURL,
		})
		if err != nil {
			return nil, err
		}
		// We want to ignore the error in case we try to create an existing table.
		if err := p.Create(); err != nil {
			if e, ok := err.(*pq.Error); ok && e.Code != duplicateTable {
				return nil, err
			}
		}
		return p, p.Prepare()
	default:
		return nil, ErrStorageNotSupported
	}
}
