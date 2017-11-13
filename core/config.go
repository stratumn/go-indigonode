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
	"github.com/stratumn/alice/core/manager"
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
		Services: []string{"identify", "relay", "kaddht", "ping"},
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

// Register all the configurable services.
func init() {
	for _, serv := range services {
		// Register configurable if it is one.
		if configurable, ok := serv.(cfg.Configurable); ok {
			configSet[configurable.ID()] = configurable
		}
	}
}

// configSet is the global configuration set.
var configSet = cfg.Set{
	"core": configHandler,
	"log":  &logging.ConfigHandler{},
}

// configHandler is the core configuration handler.
var configHandler = &ConfigHandler{}

// GlobalConfigSet returns the global configuration set.
func GlobalConfigSet() cfg.Set {
	return configSet
}

// InitConfig creates or recreates the configuration file.
//
// It fails if the file already exists, unless recreate is true, in which
// case it will load the configuration file then save it. This is useful to
// add new or missing settings to a configuration file.
func InitConfig(filename string, recreate bool) error {
	if recreate {
		if err := LoadConfig(filename); err != nil {
			return err
		}
	}

	return cfg.Save(filename, configSet, 0600, recreate)
}

// LoadConfig loads the configuration file.
//
// This avoids packages depending on the core package to have to depend on the
// cfg package.
func LoadConfig(filename string) error {
	return cfg.Load(filename, configSet)
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
	// GeneratedByVersion is the version of Alice that last save the
	// configurations file. It is ignored now but could be useful for
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

// BootService returns the ID of the boot service.
func (c *ConfigHandler) BootService() string {
	return c.config.BootService
}

// GroupServices returns services that start the service groups.
func (c *ConfigHandler) GroupServices() []manager.Service {
	var services []manager.Service

	for _, config := range c.config.ServiceGroups {
		group := manager.ServiceGroup{
			GroupID:   config.ID,
			GroupName: config.Name,
			GroupDesc: config.Desc,
			Services:  map[string]struct{}{},
		}

		for _, dep := range config.Services {
			group.Services[dep] = struct{}{}
		}

		services = append(services, &group)
	}

	return services
}
