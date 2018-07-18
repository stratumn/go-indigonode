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

package db

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemDB(t *testing.T) {
	testImplementation(t, func(*testing.T) DB {
		db, err := NewMemDB(nil)
		require.NoError(t, err, "NewMemDB()")

		return db
	})
}

func TestFileDB(t *testing.T) {
	var tmpDirs []string

	defer func() {
		for _, dir := range tmpDirs {
			os.RemoveAll(dir)
		}
	}()

	testImplementation(t, func(*testing.T) DB {
		filename, err := ioutil.TempDir("", "")
		require.NoError(t, err, "ioutil.TempDir()")
		tmpDirs = append(tmpDirs, filename)

		db, err := NewFileDB(filename, nil)
		require.NoError(t, err, "OpenLevelDBFile()")

		return db
	})
}
