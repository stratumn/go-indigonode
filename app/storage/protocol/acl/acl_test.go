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
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"testing"

	"github.com/stratumn/go-indigonode/core/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestACL(t *testing.T) {

	db, err := db.NewMemDB(nil)
	assert.NoError(t, err, "NewMemDB")

	acl := NewACL(db)

	pid1, err := peer.IDB58Decode("QmQnYf23kQ7SvuPZ3mQcg3RuJMr9E39fBvm89Nz4bevJdt")
	assert.NoError(t, err, "IDB58Decode")
	pid2, err := peer.IDB58Decode("QmYxjwtGm1mKL61Cc6NhBRgMWFz39r3k5tRRqWWiQoux7V")
	assert.NoError(t, err, "IDB58Decode")
	pid3, err := peer.IDB58Decode("QmcqyMD1UrnqtVgoG5rHaXmQfqg8LK6Wk6NUpvdUmf78Rv")
	assert.NoError(t, err, "IDB58Decode")
	pid4, err := peer.IDB58Decode("QmNnP696qYWjwyiqoepYauwpdZp62NoKEjBV4Q1MJVhMUC")
	assert.NoError(t, err, "IDB58Decode")

	// fileHash must be exactly 34 (== hashSize) bytes.
	fileHash := []byte("this is the hash of the file !!!!!")
	require.Len(t, fileHash, hashSize, "len(fileHash)")

	// Authorize peers.
	err = acl.Authorize(context.Background(), []peer.ID{pid1, pid2}, fileHash)
	assert.NoError(t, err, "Authorize")

	// Check that peers are authorized.
	assertPeerAuthorized(t, acl, pid1, fileHash)
	assertPeerAuthorized(t, acl, pid2, fileHash)
	assertPeerNotAuthorized(t, acl, pid3, fileHash)
	assertPeerNotAuthorized(t, acl, pid4, fileHash)

	// Add a peer.
	err = acl.Authorize(context.Background(), []peer.ID{pid3}, fileHash)
	assert.NoError(t, err, "Authorize")

	// Check that all peers are still authorized.
	assertPeerAuthorized(t, acl, pid1, fileHash)
	assertPeerAuthorized(t, acl, pid2, fileHash)
	assertPeerAuthorized(t, acl, pid3, fileHash)
	assertPeerNotAuthorized(t, acl, pid4, fileHash)

}

func TestACL_BadHashSize(t *testing.T) {

	db, err := db.NewMemDB(nil)
	assert.NoError(t, err, "NewMemDB")

	acl := NewACL(db)

	pid1, err := peer.IDB58Decode("QmQnYf23kQ7SvuPZ3mQcg3RuJMr9E39fBvm89Nz4bevJdt")
	assert.NoError(t, err, "IDB58Decode")

	fileHash := []byte("The hash is too short !")

	// Authorize peers fails.
	err = acl.Authorize(context.Background(), []peer.ID{pid1}, fileHash)
	assert.EqualError(t, err, ErrIncorrectHashSize.Error(), "Authorize")

	// IsQuthorized fails too.
	b, err := acl.IsAuthorized(context.Background(), pid1, fileHash)
	assert.EqualError(t, err, ErrIncorrectHashSize.Error(), "Authorize")
	assert.False(t, b, "Authorize")

}

// ====================================================================================================================
// ====					 																					HELPERS																									=====
// ====================================================================================================================
func assertPeerAuthorized(t *testing.T, acl ACL, pid peer.ID, h []byte) {
	a, err := acl.IsAuthorized(context.Background(), pid, h)
	assert.NoError(t, err, "IsAuthorized")
	assert.True(t, a, "IsAuthorized")
}

func assertPeerNotAuthorized(t *testing.T, acl ACL, pid peer.ID, h []byte) {
	a, err := acl.IsAuthorized(context.Background(), pid, h)
	assert.NoError(t, err, "IsAuthorized")
	assert.False(t, a, "IsAuthorized")
}
