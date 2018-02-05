// Copyright © 2017-2018  Stratumn SAS
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

package trie

import (
	"crypto/rand"
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/db"
)

func BenchmarkNode_MarshalBinary(b *testing.B) {
	tests := []struct {
		name string
		node Node
	}{{
		"leaf",
		&Leaf{Value: []byte("Alice")},
	}, {
		"hash",
		&Hash{Hash: testHash1},
	}, {
		"parent",
		&Branch{
			Value: []byte("Alice"),
			EmbeddedNodes: [...]Node{
				Null{},
				Null{},
				Null{},
				&Hash{Hash: testHash1},
				Null{},
				Null{},
				Null{},
				Null{},
				Null{},
				Null{},
				Null{},
				&Hash{Hash: testHash2},
				Null{},
				Null{},
				Null{},
				Null{},
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
		node Node
	}{{
		"leaf",
		&Leaf{Value: []byte("Alice")},
	}, {
		"hash",
		&Hash{Hash: testHash1},
	}, {
		"parent",
		&Branch{
			Value: []byte("Alice"),
			EmbeddedNodes: [...]Node{
				Null{},
				Null{},
				Null{},
				&Hash{Hash: testHash1},
				Null{},
				Null{},
				Null{},
				Null{},
				Null{},
				Null{},
				Null{},
				&Hash{Hash: testHash2},
				Null{},
				Null{},
				Null{},
				Null{},
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
				_, _, err := UnmarshalNode(buf)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}

func BenchmarkNibs_Append(b *testing.B) {
	n1 := NewNibs([]byte{0x12, 0x33}, true)
	n2 := NewNibs([]byte{0x45, 0x67}, false)
	n3 := NewNibs([]byte{0x89, 0xAB}, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n1.Append(n2, n3)
	}
}

func BenchmarkTrie_Put_MemDB(b *testing.B) {
	database, err := db.NewMemDB(nil)
	if err != nil {
		b.Error(err)
		return
	}

	trie := New(database)
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Error(err)
		return
	}

	if _, err := rand.Read(value); err != nil {
		b.Error(err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkTrie_Put_MemDB_Diff(b *testing.B) {
	database, err := db.NewMemDB(nil)
	if err != nil {
		b.Error(err)
		return
	}

	diff := db.NewDiff(database)
	trie := New(diff)
	key := make([]byte, 64*b.N)
	value := make([]byte, 128*b.N)

	if _, err := rand.Read(key); err != nil {
		b.Error(err)
		return
	}

	if _, err := rand.Read(value); err != nil {
		b.Error(err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		diff.Reset()

		if err := trie.Put(key[i*64:(i+1)*64], value[i*128:(i+1)*128]); err != nil {
			b.Error(err)
		}

		if err := diff.Apply(); err != nil {
			b.Error(err)
		}
	}
}
