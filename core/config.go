// Copyright © 2017 Stratumn SAS
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
	"github.com/stratumn/alice/release"
)

// DefaultConfig is the default core configuration.
var DefaultConfig = Config{
	GeneratedByVersion: release.Version,
	BootService:        "boot",
	ServiceGroups: []ServiceGroupConfig{{
		ID:       "boot",
		Name:     "Boot Services",
		Desc:     "Starts boot services.",
		Services: []string{"system", "bootstrap", "api"},
	}, {
		ID:       "system",
		Name:     "System Services",
		Desc:     "Starts system services.",
		Services: []string{"signal", "pruner"},
	}, {
		ID:       "p2p",
		Name:     "P2P Services",
		Desc:     "Starts P2P services.",
		Services: []string{"identify", "relay", "kaddht", "ping", "clock"},
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
	}},
	EnableBootScreen: true,
}

// ConfigSet represents a set of configurables.
//
// This avoids packages depending on the core package to have to depend on the
// cfg package.
type ConfigSet = cfg.Set

// NewConfigSet creates a new set of configurables.
func NewConfigSet() ConfigSet {
	set := cfg.Set{
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

// InitConfig creates or recreates the configuration file.
//
// It fails if the file already exists, unless recreate is true, in which
// case it will load the configuration file then save it. This is useful to
// add new or missing settings to a configuration file.
func InitConfig(set ConfigSet, filename string, recreate bool) error {
	if recreate {
		if err := LoadConfig(set, filename); err != nil {
			return err
		}
	}

	return cfg.Save(set, filename, 0600, recreate)
}

// LoadConfig loads the configuration file.
//
// This avoids packages depending on the core package to have to depend on the
// cfg package.
func LoadConfig(set ConfigSet, filename string) error {
	return cfg.Load(set, filename)
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
	// GeneratedByVersion is the version of Alice that last saved the
	// configuration file. It is ignored now but could be useful for
	// migrations.
	GeneratedByVersion string `toml:"generated_by_version" comment:"The version of Alice that generated this file."`

	// BootService is the service to launch when starting the node.
	BootService string `toml:"boot_service" comment:"Service to launch when starting the node."`

	// ServiceGroups are groups of services that can be started as a
	// service.
	ServiceGroups []ServiceGroupConfig `toml:"service_groups" comment:"Groups of services."`

	// EnableBootScreen is whether to show the boot screen when starting
	// the node.
	EnableBootScreen bool `toml:"enable_boot_screen" comment:"Whether to show the boot screen when starting the node."`
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
	c.config.GeneratedByVersion = release.Version

	return nil
}
