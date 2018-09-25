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
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stratumn/go-node/core/db"
)

func BenchmarkNode_MarshalBinary(b *testing.B) {
	tests := []struct {
		name string
		node node
	}{{
		"leaf",
		&leaf{Value: []byte("Alice")},
	}, {
		"edge",
		&edge{Hash: testHash1},
	}, {
		"parent",
		&branch{
			Value: []byte("Alice"),
			EmbeddedNodes: [...]node{
				null{},
				null{},
				null{},
				&edge{Path: []uint8{1, 2, 3}, Hash: testHash1},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				&edge{Path: []uint8{10, 11, 12}, Hash: testHash2},
				null{},
				null{},
				null{},
				null{},
			},
		},
	}}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tt.node.MarshalBinary()
			}
		})
	}
}

func BenchmarkUnmarshalNode(b *testing.B) {
	tests := []struct {
		name string
		node node
	}{{
		"leaf",
		&leaf{Value: []byte("Alice")},
	}, {
		"edge",
		&edge{Hash: testHash1},
	}, {
		"parent",
		&branch{
			Value: []byte("Alice"),
			EmbeddedNodes: [...]node{
				null{},
				null{},
				null{},
				&edge{Path: []uint8{1, 2, 3}, Hash: testHash1},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				null{},
				&edge{Path: []uint8{10, 11, 12}, Hash: testHash2},
				null{},
				null{},
				null{},
				null{},
			},
		},
	}}

	for _, tt := range tests {
		buf, err := tt.node.MarshalBinary()
		if err != nil {
			b.Error(err)
			continue
		}

		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _, err := unmarshalNode(buf)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}

func BenchmarkNibs_Append(b *testing.B) {
	n1 := newNibs([]byte{0x12, 0x33}, true)
	n2 := newNibs([]byte{0x45, 0x67}, false)
	n3 := newNibs([]byte{0x89, 0xAB}, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n1.Append(n2, n3)
	}
}

func BenchmarkTrie_Put_Map(b *testing.B) {
	trie := New()
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Fatal(err)
	}

	if _, err := rand.Read(value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}
	}

	if err := trie.Commit(); err != nil {
		b.Error(err)
	}
}

func BenchmarkTrie_Put_MemDB(b *testing.B) {
	database, err := db.NewMemDB(nil)
	if err != nil {
		b.Fatal(err)
	}

	trie := New(OptDB(database))
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Fatal(err)
	}

	if _, err := rand.Read(value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}
	}

	if err := trie.Commit(); err != nil {
		b.Error(err)
	}
}

func BenchmarkTrie_Put_MemDB_Diff(b *testing.B) {
	database, err := db.NewMemDB(nil)
	if err != nil {
		b.Fatal(err)
	}

	diff := db.NewDiff(database)
	trie := New(OptDB(diff))
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Fatal(err)
	}

	if _, err := rand.Read(value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}

	}

	if err := trie.Commit(); err != nil {
		b.Error(err)
	}

	if err := diff.Apply(); err != nil {
		b.Error(err)
	}
}

func BenchmarkTrie_Put_FileDB(b *testing.B) {
	filename, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(filename)

	database, err := db.NewFileDB(filename, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer database.Close()

	trie := New(OptDB(database))
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Fatal(err)
	}

	if _, err := rand.Read(value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}
	}

	if err := trie.Commit(); err != nil {
		b.Error(err)
	}
}

func BenchmarkTrie_Put_FileDB_Diff(b *testing.B) {
	filename, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(filename)

	database, err := db.NewFileDB(filename, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer database.Close()

	diff := db.NewDiff(database)
	trie := New(OptDB(diff))
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Fatal(err)
	}

	if _, err := rand.Read(value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}

	}

	if err := trie.Commit(); err != nil {
		b.Error(err)
	}

	if err := diff.Apply(); err != nil {
		b.Error(err)
	}
}

func BenchmarkTrie_Range_FileDB(b *testing.B) {
	filename, err := ioutil.TempDir("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(filename)

	database, err := db.NewFileDB(filename, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer database.Close()

	trie := New(OptDB(database))
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Fatal(err)
	}

	if _, err := rand.Read(value); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}
	}

	if err := trie.Commit(); err != nil {
		b.Error(err)
	}

	iter := trie.IterateRange(nil, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := iter.Next(); err != nil {
			b.Error(err)
		}
	}
}
