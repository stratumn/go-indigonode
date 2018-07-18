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
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "cfg.toml")

	// Save original configuration.
	err = ioutil.WriteFile(filename, []byte(confZero), 0600)
	require.NoError(t, err, "ioutil.WriteFile(filename, []byte(confZero), 0600)")

	set := NewConfigurableSet()

	// Migrate and load.
	err = LoadConfig(set, filename)
	require.NoError(t, err, "LoadConfig(set, filename)")

	migratedConf := set.Configs()

	// Create default config.
	defConf := NewConfigurableSet().Configs()

	// If migrations are consistent, both configs should be the same.
	for k, gotVal := range migratedConf {
		gotBuf := bytes.NewBuffer(nil)
		enc := toml.NewEncoder(gotBuf)
		enc.QuoteMapKeys(true)
		err := enc.Encode(gotVal)
		if !assert.NoErrorf(t, err, "%s: enc.Encode(gotVal)", k) {
			continue
		}

		wantVal := defConf[k]
		wantBuf := bytes.NewBuffer(nil)
		enc = toml.NewEncoder(wantBuf)
		enc.QuoteMapKeys(true)
		err = enc.Encode(wantVal)
		if !assert.NoErrorf(t, err, "%s: enc.Encode(wantVal)", k) {
			continue
		}

		assert.Equalf(t, wantBuf.String(), gotBuf.String(), "%s", k)
	}
}

// Original configuration before migrations.
const confZero = `
# Indigo Node configuration file. Keep private!!!

# Settings for the cli module.
[cli]

  # The address of the gRPC API.
  api_address = "/ip4/127.0.0.1/tcp/8904"

  # The maximum allowed duration to dial the API.
  dial_timeout = "30s"

  # Whether to display color output.
  enable_color_output = true

  # Whether to display debug output.
  enable_debug_output = false

  # The version of Indigo Node that generated this file.
  generated_by_version = "v0.0.1"

  # Which prompt backend to use (vt100, readline). VT100 is not available on Windows.
  prompt_backend = "vt100"

  # The maximum allowed duration for requests to the API.
  request_timeout = "30s"

  # Path to a TLS certificate.
  tls_certificate_file = ""

  # Override the server name of the TLS authority (for testing only).
  tls_server_name_override = ""
`
