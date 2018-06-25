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

package trie

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrie_IterateRange(t *testing.T) {
	tests := []struct {
		name  string
		puts  []string
		start string
		stop  string
		want  []string
	}{{
		"root-empty",
		[]string{
			"", "",
		},
		"00",
		"ff",
		[]string{},
	}, {
		"tree-one",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"01",
		"0120",
		[]string{
			"01", "01",
			"0100", "02",
			"0110", "03",
		},
	}, {
		"tree-two",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"0121",
		"11",
		[]string{
			"0121", "05",
			"0200", "06",
		},
	}, {
		"tree-three",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"",
		"0110",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
		},
	}, {
		"tree-four",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"01FF",
		"FFFF",
		[]string{
			"0200", "06",
			"11", "07",
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := New()

			for i := 0; i < len(tt.puts); i += 2 {
				key, err := hex.DecodeString(tt.puts[i])
				require.NoError(t, err, "hex.DecodeString(key)")

				val, err := hex.DecodeString(tt.puts[i+1])
				require.NoError(t, err, "hex.DecodeString(val)")

				require.NoErrorf(t, trie.Put(key, val), "trie.Put(%x, %x)", key, val)
			}

			require.NoError(t, trie.Commit(), "trie.Commit()")

			start, err := hex.DecodeString(tt.start)
			require.NoError(t, err, "hex.DecodeString(start)")

			stop, err := hex.DecodeString(tt.stop)
			require.NoError(t, err, "hex.DecodeString(stop)")

			iter := trie.IterateRange(start, stop)
			defer iter.Release()

			i := 0

			for ; i < len(tt.want); i += 2 {
				k, err := hex.DecodeString(tt.want[i])
				require.NoErrorf(t, err, "hex.DecodeString(key[%d])", i/2)

				v, err := hex.DecodeString(tt.want[i+1])
				require.NoErrorf(t, err, "hex.DecodeString(val[%d])", i/2)

				next, err := iter.Next()
				require.NoErrorf(t, err, "iter.Next()#%d", i/2)
				assert.Equalf(t, true, next, "iter.Next()#%d", i/2)

				if !next {
					break
				}

				assert.Equalf(t, k, iter.Key(), "iter.Key()#%d", i/2)
				assert.Equalf(t, v, iter.Value(), "iter.Value()#%d", i/2)
			}

			next, err := iter.Next()
			require.NoErrorf(t, err, "iter.Next()#%d", i/2)
			assert.Equalf(t, false, next, "iter.Next()#%d", i/2)

			iter.Release()
			_, err = iter.Next()
			assert.EqualError(t, err, ErrIteratorReleased.Error(), "iter.Next()")
		})
	}
}

func TestTrie_IteratePrefix(t *testing.T) {
	tests := []struct {
		name   string
		puts   []string
		prefix string
		want   []string
	}{{
		"root-empty",
		[]string{
			"", "",
		},
		"00",
		[]string{},
	}, {
		"tree-one",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
	}, {
		"tree-two",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"02",
		[]string{
			"0200", "06",
		},
	}, {
		"tree-three",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"01",
		[]string{
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
		},
	}, {
		"tree-four",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0110", "03",
			"0120", "04",
			"0121", "05",
			"0200", "06",
			"11", "07",
		},
		"12",
		[]string{},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := New()

			for i := 0; i < len(tt.puts); i += 2 {
				key, err := hex.DecodeString(tt.puts[i])
				require.NoError(t, err, "hex.DecodeString(key)")

				val, err := hex.DecodeString(tt.puts[i+1])
				require.NoError(t, err, "hex.DecodeString(val)")

				require.NoErrorf(t, trie.Put(key, val), "trie.Put(%x, %x)", key, val)
			}

			require.NoError(t, trie.Commit(), "trie.Commit()")

			prefix, err := hex.DecodeString(tt.prefix)
			require.NoError(t, err, "hex.DecodeString(prefix)")

			iter := trie.IteratePrefix(prefix)
			defer iter.Release()

			i := 0

			for ; i < len(tt.want); i += 2 {
				k, err := hex.DecodeString(tt.want[i])
				require.NoErrorf(t, err, "hex.DecodeString(key[%d])", i/2)

				v, err := hex.DecodeString(tt.want[i+1])
				require.NoErrorf(t, err, "hex.DecodeString(val[%d])", i/2)

				next, err := iter.Next()
				require.NoErrorf(t, err, "iter.Next()#%d", i/2)
				assert.Equalf(t, true, next, "iter.Next()#%d", i/2)

				if !next {
					break
				}

				assert.Equalf(t, k, iter.Key(), "iter.Key()#%d", i/2)
				assert.Equalf(t, v, iter.Value(), "iter.Value()#%d", i/2)
			}

			next, err := iter.Next()
			require.NoErrorf(t, err, "iter.Next()#%d", i/2)
			assert.Equalf(t, false, next, "iter.Next()#%d", i/2)

			iter.Release()
			_, err = iter.Next()
			assert.EqualError(t, err, ErrIteratorReleased.Error(), "iter.Next()")
		})
	}
}
