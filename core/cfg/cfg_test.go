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
	"sort"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name        string   `toml:"name"`
	Version     int      `toml:"version"`
	Started     bool     `toml:"started"`
	Author      string   `toml:"author"`
	StringStuff []string `toml:"string_stuff"`
	IntStuff    []int    `toml:"int_stuff"`
}

type testHandler struct {
	config      *testConfig
	name        string
	version     int
	started     bool
	author      string
	stringStuff []string
	intStuff    []int
}

func (h *testHandler) ID() string {
	return h.name
}

func (h *testHandler) Config() interface{} {
	if h.config != nil {
		return *h.config
	}

	return testConfig{h.name, h.version, h.started, h.author, h.stringStuff, h.intStuff}
}

func (h *testHandler) SetConfig(config interface{}) error {
	c := config.(testConfig)
	h.config = &c
	h.name = c.Name
	h.version = c.Version
	h.started = c.Started
	h.stringStuff = c.StringStuff
	h.intStuff = c.IntStuff
	h.author = c.Author
	return nil
}

func TestCfg(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "cfg.toml")

	zipHandlerName := "zip"
	tarHandlerName := "tar"

	zipSave := testHandler{name: zipHandlerName, intStuff: []int{1}}
	tarSave := testHandler{name: tarHandlerName}
	setSave := NewSet([]Configurable{&zipSave, &tarSave})
	assert.EqualValues(t, Set{
		zipHandlerName: &zipSave,
		tarHandlerName: &tarSave,
	}, setSave)

	err = setSave.Save(filename, 0644, &ConfigSaveOpts{
		Overwrite: false,
		Backup:    true,
	})
	require.NoError(t, err, "Save(filename)")

	zipLoad := testHandler{name: zipHandlerName, version: 1}
	tarLoad := testHandler{name: tarHandlerName, version: 1}
	setLoad := NewSet([]Configurable{&zipLoad, &tarLoad})

	err = setLoad.Load(filename)
	require.NoError(t, err, "Load(filename)")

	assert.Equal(t, zipSave.version, zipLoad.version, "zipLoad")
	assert.Equal(t, zipSave.intStuff, zipLoad.intStuff, "zipLoad")
	assert.Equal(t, tarSave.version, tarLoad.version, "tarLoad")

	t.Run("Save", func(t *testing.T) {
		files, err := ioutil.ReadDir(dir)
		require.NoError(t, err)

		// Should create a backup.
		err = setSave.Save(filename, 0644, &ConfigSaveOpts{
			Overwrite: true,
			Backup:    true,
		})
		require.NoError(t, err, "Save(filename)")
		filesPlusBackup, _ := ioutil.ReadDir(dir)
		assert.Equal(t, len(files)+1, len(filesPlusBackup))

		// Should not create a backup.
		err = setSave.Save(filename, 0644, &ConfigSaveOpts{
			Overwrite: true,
			Backup:    false,
		})
		require.NoError(t, err, "Save(filename)")
		filesNoBackup, _ := ioutil.ReadDir(dir)
		assert.Equal(t, len(filesPlusBackup), len(filesNoBackup))

		// Should return an error because file already exists.
		err = setSave.Save(filename, 0644, &ConfigSaveOpts{
			Overwrite: false,
			Backup:    false,
		})
		require.Error(t, err, "Save(filename)")
	})

	t.Run("Tree", func(t *testing.T) {
		tree, err := setSave.Tree()
		assert.NoError(t, err, "Tree")
		keys := tree.tree.Keys()
		sort.Strings(keys)
		assert.EqualValues(t, []string{"tar", "zip"}, keys)
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("Key exists - string value", func(t *testing.T) {
			val, err := setSave.Get("zip.name")
			assert.NoError(t, err, "Get")
			assert.Equal(t, zipHandlerName, val)
		})
		t.Run("Key exists - int value", func(t *testing.T) {
			val, err := setSave.Get("zip.version")
			assert.NoError(t, err, "Get")
			assert.Equal(t, int64(0), val)
		})
		t.Run("Key exists - bool value", func(t *testing.T) {
			val, err := setSave.Get("zip.started")
			assert.NoError(t, err, "Get")
			assert.Equal(t, false, val)
		})

		t.Run("Key exists - slice string value", func(t *testing.T) {
			val, err := setSave.Get("zip.string_stuff")
			assert.NoError(t, err, "Get")
			assert.EqualValues(t, []interface{}(nil), val)
		})
		t.Run("Key exists - slice int value", func(t *testing.T) {
			val, err := setSave.Get("zip.int_stuff")
			assert.NoError(t, err, "Get")
			assert.EqualValues(t, []interface{}{int64(1)}, val)
		})

		t.Run("Returns a node of the tree", func(t *testing.T) {
			val, err := setSave.Get("zip")
			assert.NoError(t, err, "Get")
			assert.IsType(t, &toml.Tree{}, val)
		})

		t.Run("Key does not exist", func(t *testing.T) {
			_, err := setSave.Get("none")
			assert.EqualError(t, err, "could not get \"none\": setting not found")
		})
	})

	t.Run("Set", func(t *testing.T) {
		t.Run("Value is set - string", func(t *testing.T) {
			err := setSave.Set("zip.name", "test")
			assert.NoError(t, err, "Set")
			val, _ := setSave.Get("zip.name")
			assert.Equal(t, "test", val)
		})
		t.Run("Value is set - bool", func(t *testing.T) {
			err := setSave.Set("zip.started", "true")
			assert.NoError(t, err, "Set")
			val, _ := setSave.Get("zip.started")
			assert.Equal(t, true, val)
		})
		t.Run("Value is set - int", func(t *testing.T) {
			err := setSave.Set("zip.version", "2")
			assert.NoError(t, err, "Set")
			val, _ := setSave.Get("zip.version")
			assert.Equal(t, int64(2), val)
		})
		t.Run("Value is set - slice of strings", func(t *testing.T) {
			err := setSave.Set("zip.string_stuff", "one,two,three")
			assert.NoError(t, err, "Set")
			val, _ := setSave.Get("zip.string_stuff")
			assert.EqualValues(t, []interface{}{"one", "two", "three"}, val)
		})

		t.Run("Value is set - slice of ints", func(t *testing.T) {
			err := setSave.Set("zip.int_stuff", "1,2,3")
			assert.NoError(t, err, "Set")
			val, _ := setSave.Get("zip.int_stuff")
			assert.EqualValues(t, []interface{}{int64(1), int64(2), int64(3)}, val)
		})

		t.Run("Key does not exist", func(t *testing.T) {
			err := setSave.Set("none", "something")
			assert.EqualError(t, err, "could not set \"none\": setting not found")
		})

		t.Run("Group of settings", func(t *testing.T) {
			err := setSave.Set("zip", "something")
			assert.EqualError(t, err, "could not set \"zip\": cannot edit a group of attribute")
		})

		t.Run("Wrong value type", func(t *testing.T) {
			err := setSave.Set("zip.version", "wrongtype")
			assert.EqualError(t, err, "could not set \"zip.version\": wrong type for value \"wrongtype\" (expected int)")
		})
	})
}
