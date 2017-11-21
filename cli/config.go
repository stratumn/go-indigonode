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

package cli

import (
	"github.com/stratumn/alice/core/cfg"
	"github.com/stratumn/alice/release"
)

// DefaultConfig is the default CLI configuration.
var DefaultConfig = Config{
	GeneratedByVersion: release.Version,
	PromptBackend:      "vt100",
	EnableColorOutput:  true,
	APIAddress:         "/ip4/127.0.0.1/tcp/8904",
	DialTimeout:        "30s",
	RequestTimeout:     "30s",
	EnableDebugOutput:  false,
}

// configSet is the global configuration set.
var configSet = cfg.Set{
	"cli": configHandler,
}

// configHandler is the core configuration handler.
var configHandler = &ConfigHandler{}

// Config contains configuration options for the CLI.
type Config struct {
	// GeneratedByVersion is the version of Alice that last save the
	// configurations file. It is ignored now but could be useful for
	// migrations.
	GeneratedByVersion string `toml:"generated_by_version" comment:"The version of Alice that generated this file."`

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

	return cfg.Save(configSet, filename, 0600, recreate)
}

// LoadConfig loads the configuration file.
//
// This avoids packages depending on the core package to have to depend on the
// cfg package.
func LoadConfig(filename string) error {
	return cfg.Load(configSet, filename)
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
