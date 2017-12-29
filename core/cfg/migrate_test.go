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
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
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

func createOld(t *testing.T, filename string, version int) {
	old := map[string]interface{}{
		"version": map[string]int{"version": version},
		"zip":     map[string]string{"name": "zip extractor"},
	}

	oldTree, err := toml.TreeFromMap(old)
	if err != nil {
		t.Fatalf("toml.TreeFromMap(old): error: %s", err)
	}

	mode := os.O_WRONLY | os.O_CREATE
	f, err := os.OpenFile(filename, mode, 0600)
	if err != nil {
		t.Fatalf("os.OpenFile(filename, mode, 0600): error: %s", err)
	}

	defer f.Close()

	if _, err := oldTree.WriteTo(f); err != nil {
		t.Fatalf("old.Tree.WriteTo(f): error: %s", err)
	}
}

func TestMigrate(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf(`ioutil.TempDir("", ""): error: %s`, err)
	}

	filename := filepath.Join(dir, "migrate.toml")
	createOld(t, filename, 1)

	migrations := []MigrateHandler{
		func(tree *Tree) error {
			t.Error("migration 0 should not be called")
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
	if err != nil {
		t.Fatalf(`LoadWithMigrations(set, filename, "version", migrations): error: %s`, err)
	}

	if got, want := version.config.Version, 3; got != want {
		t.Errorf("Load(filename): version.Value = %d want %d", got, want)
	}

	if got, want := zip.version, "2.0.0"; got != want {
		t.Errorf("Load(filename): zip.version = %q want %q", got, want)
	}

	if got, want := tar.name, "tar"; got != want {
		t.Errorf("Load(filename): tar.name = %q want %q", got, want)
	}
}

func TestMigrate_migrationError(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf(`ioutil.TempDir("", ""): error: %s`, err)
	}

	filename := filepath.Join(dir, "migrate.toml")
	createOld(t, filename, 0)

	migrations := []MigrateHandler{
		func(tree *Tree) error {
			return errors.New("fail")
		},
	}

	set := Set{"version": &testVersionHandler{}}

	err = Migrate(set, filename, "version.version", migrations, 0600)
	if err == nil {
		t.Error(`LoadWithMigrations(set, filename, "version", migrations): expected error`)
	}
}
