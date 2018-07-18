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

package cli

import (
	"github.com/stratumn/go-indigonode/core/cfg"
)

// ConfigVersionKey is the key of the configuration version number in the TOML
// file.
const ConfigVersionKey = "cli.configuration_version"

// DefaultConfig is the default CLI configuration.
var DefaultConfig = Config{
	ConfigVersion:     len(migrations),
	PromptBackend:     "vt100",
	EnableColorOutput: true,
	APIAddress:        "/ip4/127.0.0.1/tcp/8904",
	DialTimeout:       "30s",
	RequestTimeout:    "30s",
	EnableDebugOutput: false,
}

// NewConfigurableSet creates a new set of configurables.
func NewConfigurableSet() cfg.Set {
	return cfg.NewSet([]cfg.Configurable{&ConfigHandler{}})
}

// LoadConfig loads the configuration file and applies migrations.
func LoadConfig(set cfg.Set, filename string) error {
	return cfg.Migrate(set, filename, ConfigVersionKey, migrations, 0600)
}

// Config contains configuration options for the CLI.
type Config struct {
	// ConfigVersion is the version of the configuration file.
	ConfigVersion int `toml:"configuration_version" comment:"The version of the configuration file."`

	// PromptBackend is the name of the prompt to use.
	PromptBackend string `toml:"prompt_backend" comment:"Which prompt backend to use (vt100, readline). VT100 is not available on Windows."`

	// EnableColorColor is whether to enable color output.
	EnableColorOutput bool `toml:"enable_color_output" comment:"Whether to display color output."`

	// APIAddress is the address of the gRPC API.
	APIAddress string `toml:"api_address" comment:"The address of the gRPC API."`

	// TLSCertificateFile is a path to a TLS certificate.
	TLSCertificateFile string `toml:"tls_certificate_file" comment:"Path to a TLS certificate."`

	// TLSServerNameOverride overrides the server name of the TLS authority.
	TLSServerNameOverride string `toml:"tls_server_name_override" comment:"Override the server name of the TLS authority (for testing only)."`

	// DialTimeout is the maximum duration allowed to dial the API.
	DialTimeout string `toml:"dial_timeout" comment:"The maximum allowed duration to dial the API."`

	// RequestTimeout is the maximum duration allowed for requests the API.
	RequestTimeout string `toml:"request_timeout" comment:"The maximum allowed duration for requests to the API."`

	// EnableDebugOutput if whether to display debug output.
	EnableDebugOutput bool `toml:"enable_debug_output" comment:"Whether to display debug output."`

	// InitScripts are the filenames of scripts that should be executed
	// after the CLI has started.
	InitScripts []string `toml:"init_scripts" comment:"Filenames of scripts that should be executed after the CLI has started."`
}

// ConfigHandler implements cfg.Configurable for easy configuration
// management.
type ConfigHandler struct {
	conf *Config
}

// ID returns the unique identifier of the configuration.
func (h *ConfigHandler) ID() string {
	return "cli"
}

// Config returns the current configuration or creates one with good
// defaults.
func (h *ConfigHandler) Config() interface{} {
	if h.conf != nil {
		return *h.conf
	}

	return DefaultConfig
}

// SetConfig sets the configuration.
func (h *ConfigHandler) SetConfig(config interface{}) error {
	conf := config.(Config)
	h.conf = &conf

	return nil
}
