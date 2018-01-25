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

package core

import (
	"github.com/stratumn/alice/core/cfg"
	logging "github.com/stratumn/alice/core/log"
	"github.com/stratumn/alice/core/manager"
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
		Services: []string{"signal", "pruner"},
	}, {
		ID:       "p2p",
		Name:     "P2P Services",
		Desc:     "Starts P2P services.",
		Services: []string{"identify", "relay", "kaddht", "ping", "clock", "pubsub", "raft"},
	}, {
		ID:       "network",
		Name:     "Network Services",
		Desc:     "Starts network services.",
		Services: []string{"host", "natmgr"},
	}, {
		ID:       "api",
		Name:     "API Services",
		Desc:     "Starts API services.",
		Services: []string{"grpcapi"},
	}, {
		ID:       "util",
		Name:     "Utility Services",
		Desc:     "Starts utility services.",
		Services: []string{"contacts", "event"},
	}},
	EnableBootScreen:  true,
	BootScreenHost:    "host",
	BootScreenMetrics: "metrics",
}

// ConfigurableSet represents a set of configurables.
//
// This avoids packages depending on the core package to have to depend on the
// cfg package.
type ConfigurableSet = cfg.Set

// NewConfigurableSet creates a new set of configurables bound to the given
// services.
//
// If no services are given, the builtin services are used.
func NewConfigurableSet(services []manager.Service) ConfigurableSet {
	if services == nil {
		services = BuiltinServices()
	}

	set := ConfigurableSet{
		"core": &ConfigHandler{},
		"log":  &logging.ConfigHandler{},
	}

	for _, serv := range services {
		// Register configurable if it is one.
		if configurable, ok := serv.(cfg.Configurable); ok {
			set[configurable.ID()] = configurable
		}
	}

	return set
}

// InitConfig creates a configuration file.
//
// It fails if the file already exists.
func InitConfig(set ConfigurableSet, filename string) error {
	return cfg.Save(set, filename, 0600, false)
}

// LoadConfig loads the configuration file and applies migrations.
func LoadConfig(set ConfigurableSet, filename string) error {
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

	// BootScreenMetrics is the name of the metrics service used by the
	// boot screen to display metrics.
	BootScreenMetrics string `toml:"boot_screen_metrics" comment:"Name of the metrics service used by the boot screen to display metrics."`
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
