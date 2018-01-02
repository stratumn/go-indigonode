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

package cli

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestMigrations(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf(`ioutil.TempDir("", ""): error: %s`, err)
	}

	filename := filepath.Join(dir, "cfg.toml")

	// Save original configuration.
	if err := ioutil.WriteFile(filename, []byte(confZero), 0600); err != nil {
		t.Fatalf("ioutil.WriteFile(filename, []byte(confZero), 0600): error: %s", err)
	}

	set := NewConfigurableSet()

	// Migrate and load.
	if err := LoadConfig(set, filename); err != nil {
		t.Fatalf("LoadConfig(set, filename): error: %s", err)
	}

	migratedConf := set.Configs()

	// Create default config.
	defConf := NewConfigurableSet().Configs()

	// If migrations are consistent, both configs should be the same.
	for k, gotVal := range migratedConf {
		gotBuf := bytes.NewBuffer(nil)
		enc := toml.NewEncoder(gotBuf)
		enc.QuoteMapKeys(true)
		if err := enc.Encode(gotVal); err != nil {
			t.Errorf("%s: enc.Encode(gotVal): error: %s", k, err)
			continue
		}

		wantVal := defConf[k]
		wantBuf := bytes.NewBuffer(nil)
		enc = toml.NewEncoder(wantBuf)
		enc.QuoteMapKeys(true)
		if err := enc.Encode(wantVal); err != nil {
			t.Errorf("%s: enc.Encode(wantVal): error: %s", k, err)
			continue
		}

		if got, want := gotBuf.String(), wantBuf.String(); got != want {
			t.Errorf("%s: got:\n%s\nwant\n%s\n", k, got, want)
		}
	}
}

// Original configuration before migrations.
const confZero = `
# Alice configuration file. Keep private!!!

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

  # The version of Alice that generated this file.
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
