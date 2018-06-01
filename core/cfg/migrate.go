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
	"bytes"
	"context"
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
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

	// ErrOutdatedExec is returned when the config version is more recent
	// than the executable.
	ErrOutdatedExec = errors.New("exec is out of date with config version")
)

// MigrateHandler mutates a TOML tree to update it from one version to the
// next.
type MigrateHandler func(*Tree) error

// Migrate loads a TOML file and sets the configurations of a set of
// configurables, applying migrations if needed.
//
// The version key should point to the tree path that contains the int config
// file version.
//
// The migrations should be a slice of MigrateHandler that upgrade the
// configuration from the version corresponding to its index in the slice to
// the next version.
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

	t := &Tree{tree}

	version, ok := t.GetDefault(versionKey, int64(0)).(int64)
	if !ok {
		err := errors.WithStack(ErrInvalidVersion)
		event.SetError(err)
		return err
	}

	if version > int64(len(migrations)) {
		return errors.WithStack(ErrOutdatedExec)
	}

	migrations = migrations[version:]

	for _, m := range migrations {
		if err := m(t); err != nil {
			event.SetError(err)
			return errors.Wrap(err, fmt.Sprintf("migration %d", version))
		}

		version++

		if err := t.Set(versionKey, version); err != nil {
			return err
		}
	}

	if err := set.fromTree(ctx, tree); err != nil {
		event.SetError(err)
		return err
	}

	// Save config file if there were migrations.
	if len(migrations) > 0 {
		return set.Save(filename, perms, true)
	}

	return nil
}

// treeFromStruct creates a TOML tree from a struct.
func treeFromStruct(s interface{}) (*toml.Tree, error) {
	b := bytes.NewBuffer(nil)
	enc := toml.NewEncoder(b)
	enc.QuoteMapKeys(true)

	if err := enc.Encode(s); err != nil {
		return nil, errors.WithStack(err)
	}

	tree, err := toml.LoadBytes(b.Bytes())
	return tree, errors.WithStack(err)
}
