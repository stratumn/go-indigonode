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
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"testing"

	"github.com/stratumn/alice/core/db"
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
