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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testVersionConfig struct {
	Version int `toml:"version"`
}

type testVersionHandler struct {
	config *testVersionConfig
}

func (h *testVersionHandler) ID() string {
	return "version"
}

func (h *testVersionHandler) Config() interface{} {
	if h.config != nil {
		return *h.config
	}

	return testVersionConfig{0}
}

func (h *testVersionHandler) SetConfig(config interface{}) error {
	c := config.(testVersionConfig)
	h.config = &c
	return nil
}

func createOld(t *testing.T, filename string, version interface{}) {
	old := map[string]interface{}{
		"version": map[string]interface{}{"version": version},
		"zip":     map[string]string{"name": "zip extractor"},
	}

	oldTree, err := toml.TreeFromMap(old)
	require.NoError(t, err, "toml.TreeFromMap(old)")

	mode := os.O_WRONLY | os.O_CREATE
	f, err := os.OpenFile(filename, mode, 0600)
	require.NoError(t, err, "os.OpenFile(filename, mode, 0600)")

	defer f.Close()

	_, err = oldTree.WriteTo(f)
	require.NoError(t, err, "old.Tree.WriteTo(f)")
}

func TestMigrate(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dir, err := ioutil.TempDir("", "")
	require.NoError(err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "migrate.toml")
	createOld(t, filename, 1)

	migrations := []MigrateHandler{
		func(tree *Tree) error {
			assert.Fail("migration 0 should not be called")
			return nil
		},
		func(tree *Tree) error {
			return tree.Set("zip.version", "2.0.0")
		},
		func(tree *Tree) error {
			return tree.Set("tar", testConfig{Name: "tar"})
		},
	}

	version := testVersionHandler{}
	zip := testHandler{}
	tar := testHandler{}

	set := Set{
		"version": &version,
		"zip":     &zip,
		"tar":     &tar,
	}

	err = Migrate(set, filename, "version.version", migrations, 0600)
	require.NoError(err, "Migrate()")

	assert.Equal(3, version.config.Version, "version.config.Version")
	assert.Equal("2.0.0", zip.version, "zip.version")
	assert.Equal("tar", tar.name, "tar.name")
}

func TestMigrate_migrationError(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "migrate.toml")
	createOld(t, filename, 0)

	migrations := []MigrateHandler{
		func(tree *Tree) error {
			return errors.New("fail")
		},
	}

	set := Set{"version": &testVersionHandler{}}

	err = Migrate(set, filename, "version.version", migrations, 0600)
	assert.EqualError(t, err, "migration 0: fail", "Migrate()")
}

func TestMigrate_invalidVersion(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "migrate.toml")
	createOld(t, filename, "not an int")

	set := Set{"version": &testVersionHandler{}}

	err = Migrate(set, filename, "version.version", nil, 0600)
	assert.EqualError(t, err, ErrInvalidVersion.Error(), "Migrate()")
}

func TestMigrate_outdatedVersion(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "migrate.toml")
	createOld(t, filename, 1)

	set := Set{"version": &testVersionHandler{}}

	err = Migrate(set, filename, "version.version", nil, 0600)
	assert.EqualError(t, err, ErrOutdatedExec.Error(), "Migrate()")
}
