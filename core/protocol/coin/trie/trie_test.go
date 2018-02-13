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
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Watch out for indentation within strings, tabs must be two spaces.
// Technically these are whitebox tests, they make sure the trees are properly
// structured.

func TestTrie_Put(t *testing.T) {
	tests := []struct {
		name string
		puts []string
		want string
	}{{
		"root-empty",
		[]string{
			"", "",
		},
		`
null
`,
	}, {
		"root",
		[]string{
			"", "ff00ff",
		},
		`
leaf [ff00ff]
`,
	}, {
		"root-replace",
		[]string{
			"", "ff00ff",
			"", "00ff00",
		},
		`
leaf [00ff00]
				`,
	}, {
		"simple-leaf",
		[]string{
			"12", "ff00ff",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
`,
	}, {
		"simple-leaf-replace",
		[]string{
			"12", "ff00ff",
			"12", "00ff00",
		},
		`
branch
  edge 12 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
    leaf 12 [00ff00]
`,
	}, {
		"simple-leaf-set-root",
		[]string{
			"12", "ff00ff",
			"", "00ff00",
		},
		`
branch [00ff00]
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
`,
	}, {
		"fork",
		[]string{
			"12", "ff00ff",
			"20", "00ff00",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
  edge 20 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
    leaf 20 [00ff00]
`,
	}, {
		"trident",
		[]string{
			"12", "ff00ff",
			"20", "00ff00",
			"32", "ff0000",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
  edge 20 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
    leaf 20 [00ff00]
  edge 32 QmdWWKR6EXJqDUzhzPKW9XhAsgPhUFkADS4sbSYAK98N9h
    leaf 32 [ff0000]
`,
	}, {
		"long-edge",
		[]string{
			"1234", "ff00ff",
		},
		`
branch
  edge 1234 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 1234 [ff00ff]
`,
	}, {
		"long-edge-append",
		[]string{
			"1234", "ff00ff",
			"123456", "00ff00",
		},
		`
branch
  edge 1234 QmWNVxZFkxCREUtKniU1HX2MVdTo2kRpvKLMU2GDptULyV
    branch 1234 [ff00ff]
      edge 56 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
        leaf 123456 [00ff00]
`,
	}, {
		"long-edge-split",
		[]string{
			"1234", "ff00ff",
			"12", "00ff00",
		},
		`
branch
  edge 12 QmbGuGJsGwbTDJJKom8zfNc6AZ5H69rrX8Z6fbs5f5ZF6P
    branch 12 [00ff00]
      edge 34 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
        leaf 1234 [ff00ff]
`,
	}, {
		"longer-edge-split-top",
		[]string{
			"123456", "ff00ff",
			"12", "00ff00",
		},
		`
branch
  edge 12 QmPXVxR75GiF6PkBSBRGsF9gcMgWcNmKjMiU4jQMbhbjig
    branch 12 [00ff00]
      edge 3456 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
        leaf 123456 [ff00ff]
`,
	}, {
		"longer-edge-split-bottom",
		[]string{
			"123456", "ff00ff",
			"1234", "00ff00",
		},
		`
branch
  edge 1234 QmWWwZgtwuZheqgdVxvpCjoL3W8VjRvQmDkieymqmEB5nh
    branch 1234 [00ff00]
      edge 56 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
        leaf 123456 [ff00ff]
`,
	}, {
		"longer-edge-fork-top",
		[]string{
			"123456", "ff00ff",
			"1345", "00ff00",
		},
		`
branch
  edge 1 QmZbLfHbvHUgu8YGEVinFqZNhPFxLsXSZ5GKEA3LXLZy2i
    branch 1
      edge 23456 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
        leaf 123456 [ff00ff]
      edge 345 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
        leaf 1345 [00ff00]
`,
	}, {
		"longer-edge-fork-bottom",
		[]string{
			"123456", "ff00ff",
			"12345789", "00ff00",
		},
		`
branch
  edge 12345 QmVCitpZg671C643uyAmcXRKBaUhPFEqKLYEKB4zqQYwZb
    branch 12345
      edge 6 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
        leaf 123456 [ff00ff]
      edge 789 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
        leaf 12345789 [00ff00]
`,
	}, {
		"tree",
		[]string{
			"00", "00",
			"01", "01",
			"0100", "02",
			"0200", "03",
			"0110", "04",
			"0120", "05",
			"0121", "06",
			"11", "07",
		},
		`
branch
  edge 0 QmSBrchF7jmerT7Gmb1ygqTsQSaqypJbfhkS8xeCo4vReC
    branch 0
      edge 0 QmR1tQAHxAJuDCHU78BmHBjAi25GstYQc7AgHVUinJQpUe
        leaf 00 [00]
      edge 1 QmP8MQXqQrvb3HZPyaDeKPjtERQ43yJULEVBkkJjXqm36G
        branch 01 [01]
          edge 00 QmS2rSVoqRsiwDgHj5Eno4Gw4KbNvtR4F6m763ufxpSwaK
            leaf 0100 [02]
          edge 10 QmV1sPDuGWzC2HdrVviuD3oUr5c21gywRWzjWXdSFCzozJ
            leaf 0110 [04]
          edge 2 QmcRpdc9pzQ3hPgx8JVCGr8VajJdCfk1G8tt6qG1zg6s6t
            branch 012
              edge 0 QmUYkiDvL8XXcahsPVbECq7pAZbXr4WZJpCNkegqN9v9it
                leaf 0120 [05]
              edge 1 QmY8stJR4WL2nxRKXkBTYMkSKyu56ZrVFz7kq2NhWU5Euy
                leaf 0121 [06]
      edge 200 QmSWAksxa7qTPs2k4UPtVaE25Pbsku2bresa9sAjqjGSaQ
        leaf 0200 [03]
  edge 11 QmcExRPrQeMA7Y6U8Zb4QEEKzZJe9htS6Ax7AwKLNCyWn5
    leaf 11 [07]
`,
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

				if len(val) > 0 {
					got, err := trie.Get(key)
					require.NoErrorf(t, err, "hex.Get(%x)", key)
					assert.Equalf(t, val, got, "hex.Get(%x)", key)
				}
			}

			require.NoError(t, trie.Commit(), "trie.Commit()")

			got, err := trie.Dump()
			require.NoError(t, err, "trie.dump()")
			assert.Equal(t, strings.TrimSpace(tt.want), strings.TrimSpace(got))
		})
	}
}

func TestTrie_Delete(t *testing.T) {
	tests := []struct {
		name    string
		puts    []string
		deletes []string
		want    string
	}{{
		"root",
		[]string{
			"", "ff00ff",
		},
		[]string{
			"",
		},
		`
null		
`,
	}, {
		"root-nop",
		[]string{
			"", "",
		},
		[]string{
			"",
		},
		`
null		
`,
	}, {
		"simple-leaf",
		[]string{
			"12", "ff00ff",
		},
		[]string{
			"12",
		},
		`
null		
`,
	}, {
		"simple-root",
		[]string{
			"12", "ff00ff",
			"", "00ff00",
		},
		[]string{
			"",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
`,
	}, {
		"edge-top",
		[]string{
			"12", "ff00ff",
			"1234", "00ff00",
		},
		[]string{
			"12",
		},
		`
branch
  edge 1234 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
    leaf 1234 [00ff00]
`,
	}, {
		"edge-bottom",
		[]string{
			"12", "ff00ff",
			"1234", "00ff00",
		},
		[]string{
			"1234",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
`,
	}, {
		"node-nop",
		[]string{
			"12", "ff00ff",
		},
		[]string{
			"01",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
`,
	}, {
		"leaf-nop",
		[]string{
			"12", "ff00ff",
		},
		[]string{
			"1211",
		},
		`
branch
  edge 12 QmboVHLBpwP99VJhtWSn91rVtKgQTxi58u7sQ8L3EfpuMU
    leaf 12 [ff00ff]
`,
	}, {
		"edge-nop",
		[]string{
			"12", "ff00ff",
			"123456", "00ff00",
		},
		[]string{
			"1234",
		},
		`
branch
  edge 12 QmQbJdDePqyiHRB3XbcjXGWZApbhpTtY9njMvjgkqbaqvn
    branch 12 [ff00ff]
      edge 3456 QmVZHSgCKTSi3RDK2DhAsirGm8EjtYXpHMD9kkasqpnRUN
        leaf 123456 [00ff00]
`,
	}, {
		"collapse-root",
		[]string{
			"123456", "ff0000",
			"123478", "00ff00",
		},
		[]string{
			"123456",
			"123478",
		},
		`
null
`,
	}, {
		"collapse-with-fork",
		[]string{
			"12", "00",
			"1234", "01",
			"123456", "02",
			"123478", "03",
		},
		[]string{
			"1234",
		},
		`
branch
  edge 12 QmdAejXbrTgZCnhgAXQf5iCaf3YZKZPQU8AvXZJygwpCmU
    branch 12 [00]
      edge 34 QmfLCZAbV3FCSzNqYe365WqrmRfLqppfG58kt6dUNGP3sH
        branch 1234
          edge 56 QmS2rSVoqRsiwDgHj5Eno4Gw4KbNvtR4F6m763ufxpSwaK
            leaf 123456 [02]
          edge 78 QmSWAksxa7qTPs2k4UPtVaE25Pbsku2bresa9sAjqjGSaQ
            leaf 123478 [03]
`,
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

				if len(val) > 0 {
					got, err := trie.Get(key)
					require.NoErrorf(t, err, "hex.Get(%x)", key)
					assert.Equalf(t, val, got, "hex.Get(%x)", key)
				}
			}

			for _, v := range tt.deletes {
				key, err := hex.DecodeString(v)
				require.NoError(t, err, "hex.DecodeString(key)")
				require.NoErrorf(t, trie.Delete(key), "trie.Delete(%x)", key)
			}

			require.NoError(t, trie.Commit(), "trie.Commit()")

			got, err := trie.Dump()
			require.NoError(t, err, "trie.dump()")
			assert.Equal(t, strings.TrimSpace(tt.want), strings.TrimSpace(got))
		})
	}
}
