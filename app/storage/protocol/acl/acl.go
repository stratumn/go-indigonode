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

package acl

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/db"

	peer "github.com/libp2p/go-libp2p-peer"
	logging "github.com/ipfs/go-log"
)

// log is the logger for the service.
var log = logging.Logger("storage.acl")

var (
	prefixFilePermission = []byte("ap") // prefixFilePermission + filehash + peerID -> byte
)

const hashSize = 2 + 32 // 2 bytes for multihash and 32 bytes for the SHA256.

var (
	// ErrUnauthorized is returned when a peer tries to access a file he
	// is not allowed to get.
	ErrUnauthorized = errors.New("peer not authorized for requested file")

	// ErrIncorrectHashSize is returned when using a hash that is not hashSize.
	ErrIncorrectHashSize = errors.Errorf("file hash should be %d bytes", hashSize)
)

// ACL handles permissions on the files in the storage.
type ACL interface {

	// Authorize gives a list of peers access to a file hash.
	Authorize(context.Context, []peer.ID, []byte) error

	// IsAuthorized checks if a peer is authorized for a file hash.
	IsAuthorized(context.Context, peer.ID, []byte) (bool, error)
}

type acl struct {
	db db.DB
}

// NewACL returns a new ACL handler.
func NewACL(db db.DB) ACL {
	return &acl{
		db: db,
	}
}

// Authorize gives a list of peers access to a file hash.
func (a *acl) Authorize(ctx context.Context, peerIds []peer.ID, fileHash []byte) error {

	if len(fileHash) != hashSize {
		return ErrIncorrectHashSize
	}

	pids := make([]string, len(peerIds))
	for i, pid := range peerIds {
		pids[i] = pid.Pretty()
	}

	event := log.EventBegin(ctx, "AuthorizePeer", &logging.Metadata{
		"peerIDs": strings.Join(pids, ", "),
	})

	defer event.Done()

	for _, pid := range peerIds {
		err := a.db.Put(append(append(prefixFilePermission, fileHash...), pid...), []byte{uint8(1)})
		if err != nil {
			return err
		}
	}

	return nil
}

// IsAuthorized checks if a peer is authorized for a file hash.
func (a *acl) IsAuthorized(ctx context.Context, pid peer.ID, fileHash []byte) (bool, error) {

	if len(fileHash) != hashSize {
		return false, ErrIncorrectHashSize
	}

	_, err := a.db.Get(append(append(prefixFilePermission, fileHash...), pid...))
	if errors.Cause(err) == db.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}
