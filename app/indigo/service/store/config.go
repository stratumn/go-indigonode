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
	"context"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/dummystore"
	"github.com/stratumn/go-indigocore/postgresstore"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/validator"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit/dummyauditstore"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit/postgresauditstore"
	swarmSvc "github.com/stratumn/go-indigonode/core/app/swarm/service"
	"github.com/stratumn/go-indigonode/core/protector"
	"github.com/stratumn/go-indigonode/core/streamutil"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var log = logging.Logger("indigo.store.config")

const (
	// InMemoryStorage is the StorageType value for storing in memory (no hard storage).
	InMemoryStorage = "in-memory"
	// PostgreSQLStorage is the StorageType value for storing in an external PostgreSQL database.
	PostgreSQLStorage = "postgreSQL"

	// duplicateTable is the error returned by pq when trying to create an existing table.
	duplicateTable = pq.ErrorCode("42P07")
)

var (
	// ErrStorageNotSupported is returned when the configured storage type and url are invalid.
	ErrStorageNotSupported = errors.New("invalid or unsupported storage")

	// ErrMissingConfig is returned when the provided configuration is missing required settings.
	ErrMissingConfig = errors.New("missing configuration settings")

	// ErrMissingPrivateKey is returned when no private key is set in the configuration.
	ErrMissingPrivateKey = errors.New("indigo store nodes must have a private key")
)

// PostgresConfig contains configuration for the postgres indigo store.
type PostgresConfig struct {
	// StorageDBURL is the URL of the storage database (if external storage is chosen).
	StorageDBURL string `toml:"storage_db_url" comment:"If external storage is used, the url of that storage."`
}

// ValidationConfig contains configuration for the link validation rules.
type ValidationConfig struct {
	// RulesPath is the path to the validation rules file.
	RulesPath string `toml:"rules_path" comment:"The path to the validation rules file."`

	// PluginsPath is the directory where the validator scripts are located.
	PluginsPath string `toml:"plugins_path" comment:"The directory where the validator scripts are located."`
}

// Config contains configuration options for the Indigo service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Swarm is the name of the swarm service.
	Swarm string `toml:"swarm" comment:"The name of the swarm service."`

	// Version is the version of the Indigo Store service.
	Version string `toml:"version" comment:"The version of the indigo service."`

	// NetworkID is the ID of your Indigo PoP network.
	NetworkID string `toml:"network_id" comment:"The ID of your Indigo PoP network."`

	// StoreType is the type of storage used.
	StorageType string `toml:"storage_type" comment:"The type of storage to use.\n Supported values: in-memory and postgreSQL."`

	// PostgresConfig contains configuration for the postgres indigo store.
	PostgresConfig *PostgresConfig `toml:"postgres" comment:"Configure settings for the Indigo PostgreSQL Store in the following section."`

	// ValidationConfig contains configuration for the links' validation rules.
	ValidationConfig *ValidationConfig `toml:"validation" comment:"Configure settings for the validation rules of your indigo network in the following section."`
}

// CreateAuditStore creates an audit store from the configuration.
func (c *Config) CreateAuditStore(ctx context.Context) (audit.Store, error) {
	switch c.StorageType {
	case InMemoryStorage:
		return dummyauditstore.New(), nil
	case PostgreSQLStorage:
		if c.PostgresConfig == nil {
			return nil, ErrMissingConfig
		}
		p, err := postgresauditstore.New(&postgresstore.Config{
			URL: c.PostgresConfig.StorageDBURL,
		})
		if err != nil {
			return nil, err
		}
		// We want to ignore the error in case we try to create an existing table.
		if err := p.Create(); err != nil {
			if e, ok := err.(*pq.Error); !ok || e.Code != duplicateTable {
				log.Event(ctx, "ErrPostgresCreate", logging.Metadata{"error": err.Error()})
				return nil, err
			}
		}
		return p, p.Prepare()
	default:
		log.Event(ctx, "ErrStorageNotSupported", logging.Metadata{"storageType": c.StorageType})
		return nil, ErrStorageNotSupported
	}
}

// CreateIndigoStore creates an indigo store from the configuration.
func (c *Config) CreateIndigoStore(ctx context.Context) (indigostore.Adapter, error) {
	switch c.StorageType {
	case InMemoryStorage:
		return dummystore.New(&dummystore.Config{}), nil
	case PostgreSQLStorage:
		if c.PostgresConfig == nil {
			return nil, ErrMissingConfig
		}
		p, err := postgresstore.New(&postgresstore.Config{
			URL: c.PostgresConfig.StorageDBURL,
		})
		if err != nil {
			return nil, err
		}
		// We want to ignore the error in case we try to create an existing table.
		if err := p.Create(); err != nil {
			if e, ok := err.(*pq.Error); !ok || e.Code != duplicateTable {
				log.Event(ctx, "ErrPostgresCreate", logging.Metadata{"error": err.Error()})
				return nil, err
			}
		}
		return p, p.Prepare()
	default:
		log.Event(ctx, "ErrStorageNotSupported", logging.Metadata{"storageType": c.StorageType})
		return nil, ErrStorageNotSupported
	}
}

// CreateValidator creates an indigo multivalidator.
func (c *Config) CreateValidator(ctx context.Context, store indigostore.Adapter) (validator.GovernanceManager, error) {
	if c.ValidationConfig == nil {
		return nil, errors.Wrap(ErrMissingConfig, "validation settings not found")
	}

	if store == nil {
		return nil, errors.New("an indigo store adapter is needed to initialize a validator")
	}

	return validator.NewLocalGovernor(ctx, store, &validator.Config{
		RulesPath:   c.ValidationConfig.RulesPath,
		PluginsPath: c.ValidationConfig.PluginsPath,
	})
}

// JoinIndigoNetwork joins the Indigo network and configures its manager.
// It will use different mechanisms if we are in a public or private network.
func (c *Config) JoinIndigoNetwork(ctx context.Context, host Host, swarm *swarmSvc.Swarm) (store.NetworkManager, error) {
	var networkMgr store.NetworkManager
	if swarm.NetworkMode != nil && swarm.NetworkMode.ProtectionMode == protector.PrivateWithCoordinatorMode {
		networkMgr = store.NewPrivateNetworkManager(
			streamutil.NewStreamProvider(),
			swarm.NetworkConfig,
		)
	} else {
		networkMgr = store.NewPubSubNetworkManager()
	}

	if err := networkMgr.Join(ctx, c.NetworkID, host); err != nil {
		return nil, err
	}

	return networkMgr, nil
}
