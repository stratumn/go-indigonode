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

package acl

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/db"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// log is the logger for the service.
var log = logging.Logger("storage.acl")

var (
	prefixFilePermission = []byte("ap") // prefixFilePermission + filehash + peerID -> byte
)

var (
	// ErrUnauthorized is returned when a peer tries to access a file he
	// is not allowed to get.
	ErrUnauthorized = errors.New("peer not authorized for requested file")
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
	_, err := a.db.Get(append(append(prefixFilePermission, fileHash...), pid...))
	if errors.Cause(err) == db.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}
