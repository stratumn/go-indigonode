// Copyright © 2017-2018  Stratumn SAS
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

package service

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "contacts.toml")

	mgr, err := NewManager(filename)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	assert.Equal(t, map[string]Contact{}, mgr.List(), "mgr.List()")
}

func TestNewManager_existing(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "contacts.toml")

	mgr1, err := NewManager(filename)
	require.NoError(t, err, "mgr1")

	err = mgr1.Set("alice", &Contact{PeerID: testPID})
	require.NoError(t, err, "mgr.Set()")

	mgr2, err := NewManager(filename)
	require.NoError(t, err, "mgr2")

	contact1, err := mgr1.Get("alice")
	require.NoError(t, err, "mgr1.Get()")

	contact2, err := mgr2.Get("alice")
	require.NoError(t, err, "mgr2.Get()")

	assert.Equal(t, contact1, contact2, "contact2")
}

func TestManager_Delete(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "contacts.toml")

	mgr, err := NewManager(filename)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	err = mgr.Set("alice", &Contact{PeerID: testPID})
	require.NoError(t, err, "mgr.Set()")

	require.NoError(t, mgr.Delete("alice"))

	_, err = mgr.Get("alice")
	assert.Equal(t, ErrNotFound, errors.Cause(err), "mgr.Get()")
}

func TestManager_Delete_notFound(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "contacts.toml")

	mgr, err := NewManager(filename)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	assert.Equal(t, ErrNotFound, errors.Cause(mgr.Delete("alice")))
}
