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

package cfg

import (
	"context"
	"os"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	logging "github.com/ipfs/go-log"
)

const (
	// ConfZeroPID is the default PeerID set before any migration happens.
	// The real peerID will be generated to replace this one.
	ConfZeroPID = "QmRK6DGwgkzpHRG5rCgWUxTSKvQtmhZeMa5jdeHYTtxnNq"

	// ConfZeroPK is the default private key set before any migration happens.
	// The real private key will be generated to replace this one.
	ConfZeroPK = "CAASpwkwggSjAgEAAoIBAQCss3uW+aVGR5JLd3uw2Wubwm6nGV3/R2DFG8meu4NKB3PvmmmsngjCok++h0J6/jkqKE6o4uDubbP891GqjSOB9U/NrL8WYJV0xOKzuLgaJ3+alv3Fltg/hFH38GsMycI/3elBQpGWz56xYk/Jd+kUFxUdC9IaqC6oVdDDQHM5ItmCvDygTWTazPh1OFeqFfYjjH0GHFFHOhzRQvAA/CWFll6it7ZM+Oca/MLWwGxiiMvNKkwhZnL7oYJh4NIqAnJBSwHSAusKIBNLpo28gDbvMKuaCDLp97qqLB0tzD0wFVBsBOYbdRwyuKnR30TTIYmcunIIz6xjt2W76IzHiXIHAgMBAAECggEAHoihKkE7ImDXEbavTMY0C1bl/28xukehaVgPXpwiiz0kF1HCqz1JqTfPR41ciKhX7QcxWRS71gvZTblgW/oWNAzuLhwDsO4knn+M4V+gVSd0nR1jAsM3uosnfuGn25v0Vxxh+CLP4M0WbqBGIQWtVNr75aXIYOQpU6PQhCTp+kjPs0ibIqpc+KEHSwQvHNbE7N0oeoikQGhkX8lKGvPw0Q7jvuEr/C/S8vzciv6GyFkw2LncpvXOIBGNlXT5ZrPzLP1QrGMp3DxxTAWK0IzIwkjYahsURRQFbWxJkqcbKw1sLT/hwMAvMMuT4O5z79RwoJR/ril3r6vptXWZV5lH2QKBgQDXFGQU8LEVsmWhOpYEaqQCGMjjSWWzSGm7Yd60eJydEUpvf4l3VlZs+8JjFKJGyMyp/eRo+7J3A0JcsOgrJLaR/faaTSIIh3FDcOVez6Yl4bqQUlTVwBH4AhkwTkqK1QQ2hNThbvXbv093AhxsY3f3SGg8fkuN2ziFeKzq/Wh3FQKBgQDNjwOjI8S0Xo4u6DEdNMEezS/f00S058zZtZziHTIMEQ7fFJQZ3DFVAjkgBWRBVLb1NEq9U6XdTpD/lFNUBY3DH2qjWv4LBYPk5kqFgiuKnFmoD/gnzchGs2+4EgUdODZB/QPPBCqXA7XN/JT7VoiTWGIzAf7lv84YgUmT2BQLqwKBgCdSlRG3B8ldunMF0ROxo5a2jVPwwWVL4fjeZec8/fVBighkmu90m4yFYv7WcOzcHX8e6jm/etuDfwiPV4M7zR1X/1QqsgQ5Lx4Tb/wrnsbiREfKpbQGz8I2MADC76H+XCzTkFA/BzhL++1YN3YhoXdWh6g3tvySjfzpGURFXGoZAoGAGJS6jZ6wXhVUkV1oyiJN2b4VtIFSHQP/JiWmng95taGwkpKmZzVCnPTIGgErDPjxa/8V1PAUzJMhmb6F/G0xl5zBJsmxyWWecRfs32xCgq/RtNw8A56DDZlVicB15hmbu2ZjNzU7VpW1/uzub+PYLy6Jh6n8bkLyhVGol8pmE0MCgYEArcC+amAbHBU5DExuoXCP6pF/5Q6o13tFuG3PwWoJMCFgIGDcckBl4Hs102tkdhkH+KVmwhUFPIUp8Dg9GiEkhZi8PkkhVhAJwJEIZaNSNyOja1b5K8i7wm9xJWVumLD7wPz2I/dHv/jfHFZ1KsCb9yZg4aqGWVIRt8+vgIb97mI="
)

var (
	// ErrInvalidVersion is returned when the config version is invalid.
	ErrInvalidVersion = errors.New("config version is invalid")

	// ErrInvalidSubtree is returned when a config subtree is invalid.
	ErrInvalidSubtree = errors.New("subtree is invalid")

	// ErrOutdatedExec is returned when the config version is more recent
	// than the executable.
	ErrOutdatedExec = errors.New("exec is out of date with config version")
)

// MigrateHandler mutates a TOML tree to update it from one version to the
// next.
type MigrateHandler func(*Tree) error

// Migrator declares migrations specific to a configurable.
type Migrator interface {
	// Version key returns the key containing the current int version relative
	// to the subtree of the configurable.
	VersionKey() string

	// Migrations returns all the migrations from first to last. Keys are
	// relative to the subtree of the configurable.
	Migrations() []MigrateHandler
}

// Migrate loads a TOML file and sets the configurations of a set of
// configurables, applying migrations if needed.
//
// The version key should point to the tree path that contains the int config
// file version for global migrations.
//
// The migrations should be a slice of MigrateHandler that upgrade the
// configuration from the version corresponding to its index in the slice to
// the next version.
//
// Each configurable can have its own migrations if it implements the
// Migrator interface.
func Migrate(
	set Set,
	filename string,
	versionKey string,
	migrations []MigrateHandler,
	perms os.FileMode,
) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"filename":   filename,
		"versionKey": versionKey,
	})
	event := log.EventBegin(ctx, "migrate")
	defer event.Done()

	tree, err := toml.LoadFile(filename)
	if err != nil {
		event.SetError(err)
		return errors.WithStack(err)
	}

	t := &Tree{tree: tree}

	// Apply global migrations.
	updated, err := t.migrate(migrations, versionKey)
	if err != nil {
		event.SetError(err)
		return err
	}

	// Apply per configurable migrations.
	for name, configurable := range set {
		migrator, ok := configurable.(Migrator)
		if !ok {
			continue
		}

		// Migrations are namespaced to the configurable.
		subtree, ok := t.Get(name).(*Tree)
		if !ok {
			err = errors.WithStack(ErrInvalidSubtree)
			event.SetError(err)
			return err
		}

		u, err := subtree.migrate(migrator.Migrations(), migrator.VersionKey())
		if err != nil {
			event.SetError(err)
			return err
		}

		updated = updated || u
	}

	// Remember to set the values even if there are no migrations.
	if err := set.fromTree(ctx, tree); err != nil {
		event.SetError(err)
		return err
	}

	if !updated {
		return nil
	}

	return set.Save(filename, perms, ConfigSaveOpts{
		Overwrite: true,
		Backup:    true,
	})
}
