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
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/db"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// log is the logger for the service.
var log = logging.Logger("storage.acl")

var (
	prefixAuthorizedPeers = []byte("ap") // prefixAuthorizedPeers + filehash -> [peerIDs...]
)

var (
	// ErrUnauthorized is returned when a peer tries to access a file he
	// is not allowed to get
	ErrUnauthorized = errors.New("peer not authorized for requested file")
)

// ACL handles permissions on the files in the storage.
type ACL interface {

	// Authorize gives a list of peers access to a file hash.
	Authorize(context.Context, []peer.ID, []byte) error

	// IsAuthorized checks if a peer is authorized for a file hash.
	IsAuthorized(context.Context, peer.ID, []byte) (bool, error)
}

// Keep track of the authorized peers in a map
// to have efficient insert.
type authorizedPeersMap map[string]struct{}

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

	event := log.EventBegin(ctx, "AuthorizePeer", &logging.Metadata{
		"peerIDs": peerIds,
	})
	defer event.Done()

	// Get authorized peers from DB and update them.
	authorizedPeers := authorizedPeersMap{}
	ap, err := a.db.Get(append(prefixAuthorizedPeers, fileHash...))
	if err != nil && errors.Cause(err) != db.ErrNotFound {
		event.SetError(err)
		return err
	} else if errors.Cause(err) != db.ErrNotFound {
		if err = json.Unmarshal(ap, &authorizedPeers); err != nil {
			event.SetError(err)
			return err
		}
	}

	for _, p := range peerIds {
		authorizedPeers[p.String()] = struct{}{}
	}

	newAp, err := json.Marshal(authorizedPeers)
	if err != nil {
		event.SetError(err)
		return err
	}

	if err = a.db.Put(append(prefixAuthorizedPeers, fileHash...), newAp); err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// IsAuthorized checks if a peer is authorized for a file hash.
func (a *acl) IsAuthorized(ctx context.Context, pid peer.ID, fileHash []byte) (bool, error) {

	authorizedPeers := authorizedPeersMap{}
	ap, err := a.db.Get(append(prefixAuthorizedPeers, fileHash...))
	if errors.Cause(err) == db.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, errors.WithStack(err)
	}

	if err = json.Unmarshal(ap, &authorizedPeers); err != nil {
		return false, err
	}

	_, ok := authorizedPeers[pid.String()]

	return ok, nil

}
