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

package cfg

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Author  string `toml:"author"`
}

type testHandler struct {
	config  *testConfig
	name    string
	version string
	author  string
}

func (h *testHandler) ID() string {
	return h.name
}

func (h *testHandler) Config() interface{} {
	if h.config != nil {
		return *h.config
	}

	return testConfig{h.name, "v0.1.0", h.author}
}

func (h *testHandler) SetConfig(config interface{}) error {
	c := config.(testConfig)
	h.config = &c
	h.name = c.Name
	h.version = c.Version
	h.author = c.Author
	return nil
}

func TestCfg(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "cfg.toml")

	zipSave := testHandler{name: "zip extractor"}
	tarSave := testHandler{name: "tar extractor"}
	setSave := Set{
		"zip": &zipSave,
		"tar": &tarSave,
	}

	err = Save(setSave, filename, 0644, false)
	require.NoError(t, err, "Save(filename)")

	zipLoad := testHandler{name: "default"}
	tarLoad := testHandler{name: "default"}
	setLoad := Set{
		"zip": &zipLoad,
		"tar": &tarLoad,
	}

	err = Load(setLoad, filename)
	require.NoError(t, err, "Load(filename)")

	assert.Equal(t, zipSave.name, zipLoad.name, "zipLoad")
	assert.Equal(t, tarSave.name, tarLoad.name, "tarLoad")
}
