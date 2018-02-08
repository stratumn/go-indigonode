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
	"testing"
)

func BenchmarkNode_Marshal(b *testing.B) {
	tests := []struct {
		name string
		node Node
	}{{
		"leaf",
		LeafNode{Value: []byte("Alice")},
	}, {
		"hash",
		HashNode{Hash: testHash1},
	}, {
		"parent",
		ParentNode{
			Value: []byte("Alice"),
			ChildHashes: []Node{
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash1},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash2},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
			},
		},
	}}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tt.node.Marshal()
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
		LeafNode{Value: []byte("Alice")},
	}, {
		"hash",
		HashNode{Hash: testHash1},
	}, {
		"parent",
		ParentNode{
			Value: []byte("Alice"),
			ChildHashes: []Node{
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash1},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
				HashNode{Hash: testHash2},
				NullNode{},
				NullNode{},
				NullNode{},
				NullNode{},
			},
		},
	}}

	for _, tt := range tests {
		buf, err := tt.node.Marshal()
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
