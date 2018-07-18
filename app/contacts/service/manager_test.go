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
