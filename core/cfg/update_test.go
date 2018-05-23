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
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createConfigFile(t *testing.T, filename string, config map[string]interface{}) {
	tree, err := toml.TreeFromMap(config)
	require.NoError(t, err, "toml.TreeFromMap()")

	mode := os.O_WRONLY | os.O_CREATE
	f, err := os.OpenFile(filename, mode, 0600)
	require.NoError(t, err, "os.OpenFile()")

	defer f.Close()

	_, err = tree.WriteTo(f)
	require.NoError(t, err, "tree.WriteTo(f)")
}

func TestUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "config")
	require.NoError(t, err, "ioutil.TempDir()")

	filename := filepath.Join(dir, "update.toml")
	createConfigFile(t, filename, map[string]interface{}{
		"zip": map[string]string{"name": "zip extractor"},
	})

	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err, "ioutil.ReadDir()")
	require.Len(t, files, 1)

	updates := []UpdateHandler{
		func(tree *Tree) error {
			return tree.Set("zip.version", "2.0.0")
		},
		func(tree *Tree) error {
			return tree.Set("tar", testConfig{Name: "tar"})
		},
	}

	zip := testHandler{}
	tar := testHandler{}

	set := Set{
		"zip": &zip,
		"tar": &tar,
	}

	err = Update(set, filename, updates, 0600)
	require.NoError(t, err, "Update()")

	// No backup should have been made.
	files, err = ioutil.ReadDir(dir)
	require.NoError(t, err, "ioutil.ReadDir()")
	require.Len(t, files, 1)

	assert.Equal(t, "2.0.0", zip.version, "zip.version")
	assert.Equal(t, "tar", tar.name, "tar.name")
}
