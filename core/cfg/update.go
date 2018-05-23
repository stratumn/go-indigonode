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
	"context"
	"os"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// UpdateHandler mutates a TOML tree to update it.
// It doesn't affect the configuration version, so users won't be able
// to track the updates.
//
// This should only be used when you want to apply silent updates or
// initialize a configuration with non-default values (for example
// when running alice init with custom flags).
type UpdateHandler func(*Tree) error

// Update loads a TOML file and sets the configurations of a set of
// configurables, applying updates.
//
// The updates should be "silent" updates that the user won't be able to track.
// This is mostly useful to tweak the default service configuration to take
// into account user command-line flags.
func Update(
	set Set,
	filename string,
	updates []UpdateHandler,
	perms os.FileMode,
) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename": filename,
	})
	event := log.EventBegin(ctx, "update")
	defer event.Done()

	tree, err := toml.LoadFile(filename)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	t := &Tree{tree}

	for _, u := range updates {
		if err := u(t); err != nil {
			event.SetError(err)
			return errors.WithStack(err)
		}
	}

	if err := setValuesFromTree(ctx, set, tree); err != nil {
		event.SetError(err)
		return err
	}

	return Save(set, filename, perms, SaveParams{Overwrite: true, Backup: false})
}
