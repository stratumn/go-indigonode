// Copyright © 2017  Stratumn SAS
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

package cfg

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

type Config struct {
	Name    string
	Version string
}

type Handler struct {
	config  *Config
	name    string
	version string
}

func (h *Handler) ID() string {
	return h.name
}

func (h *Handler) Config() interface{} {
	if h.config != nil {
		return *h.config
	}

	return Config{h.name, "v0.1.0"}
}

func (h *Handler) SetConfig(config interface{}) error {
	c := config.(Config)
	h.config = &c
	h.name = c.Name
	h.version = c.Version
	return nil
}

func TestCfg(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf(`ioutil.TempDir("", ""): error: %+v`, err)
	}

	filename := filepath.Join(dir, "alice.core.toml")

	zipSave := Handler{name: "zip extractor"}
	tarSave := Handler{name: "tar extractor"}
	setSave := Set{
		"zip": &zipSave,
		"tar": &tarSave,
	}

	if err := Save(filename, setSave, 0644, false); err != nil {
		t.Fatalf("Save(filename): error: %+v", err)
	}

	zipLoad := Handler{name: "default"}
	tarLoad := Handler{name: "default"}
	setLoad := Set{
		"zip": &zipLoad,
		"tar": &tarLoad,
	}

	if err := Load(filename, setLoad); err != nil {
		t.Fatalf("Load(filename): error: %+v", err)
	}

	if got, want := zipLoad.name, zipSave.name; got != want {
		t.Errorf("Load(filename): zipLoad.name = %q want %q", got, want)
	}

	if got, want := tarLoad.name, tarSave.name; got != want {
		t.Errorf("Load(filename): tarLoad.name = %q want %q", got, want)
	}
}
