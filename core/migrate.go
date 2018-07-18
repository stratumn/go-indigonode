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

package core

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/cfg"
)

var migrations = []cfg.MigrateHandler{
	func(tree *cfg.Tree) error {
		return tree.Set(ConfigVersionKey, 1)
	},
	// Set contacts default DB
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		filename, err := filepath.Abs(filepath.Join(cwd, "data", "contacts.toml"))
		if err != nil {
			return errors.WithStack(err)
		}

		return tree.Set("contacts.filename", filename)
	},
	// Set coin default DB
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		folder, err := filepath.Abs(filepath.Join(cwd, "data", "coin", "db"))
		if err != nil {
			return errors.WithStack(err)
		}

		return tree.Set("coin.db_path", folder)
	},
	// Set chat history default DB
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		filename, err := filepath.Abs(filepath.Join(cwd, "data", "chat-history.db"))
		if err != nil {
			return err
		}

		return tree.Set("chat.history_db_path", filename)
	},
	// Set storage default paths
	func(tree *cfg.Tree) error {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.WithStack(err)
		}

		localStoragePath, err := filepath.Abs(filepath.Join(cwd, "data", "storage", "files"))
		if err != nil {
			return errors.WithStack(err)
		}

		storageDB, err := filepath.Abs(filepath.Join(cwd, "data", "storage", "db"))
		if err != nil {
			return errors.WithStack(err)
		}

		if err := tree.Set("storage.local_storage", localStoragePath); err != nil {
			return err
		}

		return tree.Set("storage.db_path", storageDB)
	},
	// Set public bootstrap seeds
	func(tree *cfg.Tree) error {
		seeds := append(
			tree.Get("bootstrap.addresses").([]interface{}),
			"/ip4/138.197.77.223/tcp/8903/ipfs/Qmc1QbSba7RtPgxEw4NqXNeDpB5CpCTwv9dvdZRdTkche1",
			"/ip4/165.227.14.175/tcp/8903/ipfs/QmQVdocY8ZbYxrKRSrff2Vxmm27Mhu6DgXyWXQwmuz1b6P",
			"/ip6/2a03:b0c0:1:d0::c9:9001/tcp/8903/ipfs/QmQJib6mnEMgdCe3bGH1YP7JswHbQQejyNucvW9BjFqmWr",
			"/ip6/2400:6180:0:d0::b1:b001/tcp/8903/ipfs/Qmc1rLFp5stHrjtq4duFg6KakBcDCpB3bTjjMZVSAdnHLj",
		)
		return tree.Set("bootstrap.addresses", seeds)
	},
	// New migrations should be added here for backwards-compatibility.
	// Existing migrations should not be edited as they won't be re-run.
}
