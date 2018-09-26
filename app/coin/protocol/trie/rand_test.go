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

package trie

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stratumn/go-node/core/db"
	"github.com/stretchr/testify/require"
)

// Does some random integrity tests.
func TestTrie_Random(t *testing.T) {
	times := 10000

	trie := New()
	key := make([]byte, 16*times)
	value := make([]byte, 16*times)

	if _, err := rand.Read(key); err != nil {
		require.NoError(t, err, "rand.Read(key)")
	}

	if _, err := rand.Read(value); err != nil {
		require.NoError(t, err, "rand.Read(value)")
	}

	// Insert half the values.
	for i := 0; i < times/2; i++ {
		k := key[i*16 : (i+1)*16]
		v := value[i*16 : (i+1)*16]

		require.NoError(t, trie.Put(k, v), "trie.Put()")

		val, err := trie.Get(k)
		require.NoError(t, err, "trie.Get()")
		require.Equal(t, v, val, "trie.Get()")
	}

	require.NoError(t, trie.Check(context.Background()), "trie.Check()")
	require.NoError(t, trie.Commit(), "trie.Commit()")
	require.NoError(t, trie.Check(context.Background()), "trie.Check()")

	// Save the Merkle root and a proof.
	root1, err := trie.MerkleRoot()
	require.NoError(t, err, "trie.MerkleRoot()")
	proof1, err := trie.Proof(key[:16])
	require.NoError(t, err, "trie.Proof()")

	// Insert the other half.
	for i := times / 2; i < times; i++ {
		k := key[i*16 : (i+1)*16]
		v := value[i*16 : (i+1)*16]

		require.NoError(t, trie.Put(k, v), "trie.Put()")

		val, err := trie.Get(k)
		require.NoError(t, err, "trie.Get()")
		require.Equal(t, v, val, "trie.Get()")
	}

	require.NoError(t, trie.Check(context.Background()), "trie.Check()")
	require.NoError(t, trie.Commit(), "trie.Commit()")
	require.NoError(t, trie.Check(context.Background()), "trie.Check()")

	// Verify proofs.
	root, err := trie.MerkleRoot()
	require.NoError(t, err, "trie.MerkleRoot()")

	for i := 0; i < times; i++ {
		k := key[i*16 : (i+1)*16]
		v := value[i*16 : (i+1)*16]

		proof, err := trie.Proof(k)
		require.NoError(t, err, "trie.Proof()")

		require.NoError(t, proof.Verify(root, k, v))
	}

	// Delete half of the values.
	for i := times / 2; i < times; i++ {
		k := key[i*16 : (i+1)*16]

		require.NoError(t, trie.Delete(k), "trie.Delete()")

		_, err := trie.Get(k)
		require.EqualError(t, err, db.ErrNotFound.Error(), "trie.Get()")
	}

	require.NoError(t, trie.Check(context.Background()), "trie.Check()")
	require.NoError(t, trie.Commit(), "trie.Commit()")
	require.NoError(t, trie.Check(context.Background()), "trie.Check()")

	for i := times / 2; i < times; i++ {
		k := key[i*16 : (i+1)*16]

		_, err := trie.Get(k)
		require.EqualError(t, err, db.ErrNotFound.Error(), "trie.Get()")
	}

	// Check other half still exists.
	for i := 0; i < times/2; i++ {
		k := key[i*16 : (i+1)*16]
		v := value[i*16 : (i+1)*16]

		val, err := trie.Get(k)
		require.NoError(t, err, "trie.Get()")
		require.Equal(t, v, val, "trie.Get()")
	}

	// Make sure the Merkle root and proof are the same as before
	// hence deterministic.
	root2, err := trie.MerkleRoot()
	require.NoError(t, err, "trie.MerkleRoot()")
	require.Equal(t, root2, root1, "trie.MerkleRoot()")
	proof2, err := trie.Proof(key[:16])
	require.NoError(t, err, "trie.Proof()")
	require.Equal(t, proof1, proof2, "trie.Proof()")

	// Verify proofs.
	for i := 0; i < times/2; i++ {
		k := key[i*16 : (i+1)*16]
		v := value[i*16 : (i+1)*16]

		proof, err := trie.Proof(k)
		require.NoError(t, err, "trie.Proof()")

		require.NoError(t, proof.Verify(root2, k, v))
	}
}
