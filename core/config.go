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

package core

import (
	"github.com/stratumn/go-node/core/cfg"
	logging "github.com/stratumn/go-node/core/log"
	"github.com/stratumn/go-node/core/manager"
)

// ConfigVersionKey is the key of the configuration version number in the TOML
// file.
const ConfigVersionKey = "core.configuration_version"

// DefaultConfig is the default core configuration.
var DefaultConfig = Config{
	ConfigVersion: len(migrations),
	BootService:   "boot",
	ServiceGroups: []ServiceGroupConfig{{
		ID:       "boot",
		Name:     "Boot Services",
		Desc:     "Starts boot services.",
		Services: []string{"system", "bootstrap", "api", "util"},
	}, {
		ID:       "system",
		Name:     "System Services",
		Desc:     "Starts system services.",
		Services: []string{"signal", "pruner", "monitoring"},
	}, {
		ID:       "p2p",
		Name:     "P2P Services",
		Desc:     "Starts P2P services.",
		Services: []string{"identify", "relay", "kaddht", "ping", "clock", "pubsub"},
	}, {
		ID:       "network",
		Name:     "Network Services",
		Desc:     "Starts network services.",
		Services: []string{"host", "natmgr"},
	}, {
		ID:       "api",
		Name:     "API Services",
		Desc:     "Starts API services.",
		Services: []string{"grpcapi", "grpcweb"},
	}, {
		ID:       "util",
		Name:     "Utility Services",
		Desc:     "Starts utility services.",
		Services: []string{"contacts", "event"},
	}, {
		ID:       "indigo",
		Name:     "Stratumn Indigo Services",
		Desc:     "Starts Stratumn Indigo services for Proof-of-Process networks.",
		Services: []string{"indigostore", "indigofossilizer"},
	}},
	EnableBootScreen: true,
	BootScreenHost:   "host",
}

// NewConfigurableSet creates a new set of configurables bound to the given
// services.
// If no services are given, the builtin services are used.
func NewConfigurableSet(services []manager.Service) cfg.Set {
	if services == nil {
		services = BuiltinServices()
	}
	configurables := make([]cfg.Configurable, 0, len(services)+2)
	for _, service := range services {
		if configurable, ok := service.(cfg.Configurable); ok {
			configurables = append(configurables, configurable)
		}
	}
	configurables = append(configurables, &ConfigHandler{})
	configurables = append(configurables, &logging.ConfigHandler{})
	return cfg.NewSet(configurables)
}

// LoadConfig loads the configuration file and applies migrations.
func LoadConfig(set cfg.Set, filename string) error {
	return cfg.Migrate(set, filename, ConfigVersionKey, migrations, 0600)
}

// ServiceGroupConfig contains settings for a service group.
type ServiceGroupConfig struct {
	// ID is the unique identifier of the group.
	ID string `toml:"id" comment:"Unique identifier of the service group."`

	// Name is the name of the group.
	Name string `toml:"name" comment:"Name of the service group."`

	// Desc is the description of the group.
	Desc string `toml:"description" comment:"Description of the service group."`

	// Services are the services that the group starts.
	Services []string `toml:"services" comment:"Services started by the group."`
}

// Config contains the core settings.
type Config struct {
	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the configuration file."`

	// BootService is the service to launch when starting the node.
	BootService string `toml:"boot_service" comment:"Service to launch when starting the node."`

	// ServiceGroups are groups of services that can be started as a
	// service.
	ServiceGroups []ServiceGroupConfig `toml:"service_groups" comment:"Groups of services."`

	// EnableBootScreen is whether to show the boot screen when starting
	// the node.
	EnableBootScreen bool `toml:"enable_boot_screen" comment:"Whether to show the boot screen when starting the node."`

	// BootScreenHost is the name of the host service used by the
	// boot screen to display metrics and host addresses.
	BootScreenHost string `toml:"boot_screen_host" comment:"Name of the host service used by the boot screen to display metrics and host addresses."`
}

// ConfigHandler is a configurable for the core settings.
type ConfigHandler struct {
	config *Config
}

// ID returns the unique identifier of the service.
func (c *ConfigHandler) ID() string {
	return "core"
}

// Config returns the current service configuration or creates one with
// good default values.
func (c *ConfigHandler) Config() interface{} {
	if c.config != nil {
		return *c.config
	}

	return DefaultConfig
}

// SetConfig configures the service handler.
func (c *ConfigHandler) SetConfig(config interface{}) error {
	conf := config.(Config)
	c.config = &conf
	return nil
}
