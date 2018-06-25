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
	"context"
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
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
  edge 12 QmUMoZdpE4dTbXaG1dG1yPnJLAmGMsbsCAYDsfTasR2QDW
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
    leaf 12 [ff00ff]
  edge 20 QmVRmPYxnrzLi6QvtxWuV78A4V9x6NyvSDgPZAASJJLGSb
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
    leaf 12 [ff00ff]
  edge 20 QmVRmPYxnrzLi6QvtxWuV78A4V9x6NyvSDgPZAASJJLGSb
    leaf 20 [00ff00]
  edge 32 QmXSmKp35SSPk8ZkJhW6kTCwTshFPV39p5KFPMmFZ79QLX
    leaf 32 [ff0000]
`,
	}, {
		"long-edge",
		[]string{
			"1234", "ff00ff",
		},
		`
branch
  edge 1234 QmVrrM5T3S9Zm92B9i6bc1e4XzVnKfRyGLpAia65k77cDf
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
  edge 1234 QmPsiacKR4KvUymPqCDMTLtmeexXN1qjB4jsNP6VhFLX9B
    branch 1234 [ff00ff]
      edge 56 QmQc672vjb6oZm2aH1CEsB1z8d1s3Khf4rfRCYqiHY7GBR
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
  edge 12 QmcoEKfoEifaKmEvuPt9PXvwfd5ELoiU31mceu4BdixbuW
    branch 12 [00ff00]
      edge 34 QmVrrM5T3S9Zm92B9i6bc1e4XzVnKfRyGLpAia65k77cDf
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
  edge 12 QmcP5waom711ppFCcBV6NG5SR2nnpndWBaeff9EQYYd8h9
    branch 12 [00ff00]
      edge 3456 QmfN2NpdkB2sGpc72ELWYunqc2KPm44JCvL1q8vP8qGXco
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
  edge 1234 QmcP5waom711ppFCcBV6NG5SR2nnpndWBaeff9EQYYd8h9
    branch 1234 [00ff00]
      edge 56 QmfN2NpdkB2sGpc72ELWYunqc2KPm44JCvL1q8vP8qGXco
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
  edge 1 QmU2MomKZuWa3RP8HApuDRL4rsxgoKE6QMPz7oCvEqhkvz
    branch 1
      edge 23456 QmfN2NpdkB2sGpc72ELWYunqc2KPm44JCvL1q8vP8qGXco
        leaf 123456 [ff00ff]
      edge 345 QmTR861dPCB2KboD96FQBy7xZsw5Ny2LvGdHdZLgKUwyD3
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
  edge 12345 Qmb87psWxaptZQVeyVZ4notqirxPuyDAg5fFxUKZtQbGSa
    branch 12345
      edge 6 QmfN2NpdkB2sGpc72ELWYunqc2KPm44JCvL1q8vP8qGXco
        leaf 123456 [ff00ff]
      edge 789 QmQh2DKArwEXe5f63GmQDLVCzGEvv9ohMxYzjRxibKKACc
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
  edge 0 QmPLz81zGZ6PMv3i3nFRAzMT6AzynCDqo66Qwz3au3SJx9
    branch 0
      edge 0 QmYUffAgALxiUQonbhAVXjknTq3dNf3AfHQGQ8P5xny7TU
        leaf 00 [00]
      edge 1 QmZHUh95Qttm2vTTMpb44AyFbuZb4YYa8VWeEbCjLYfamS
        branch 01 [01]
          edge 00 QmPqGSdVzXJXAvNUrnxiwBGoo8EwUD2v7c7GsJkXKMiHPA
            leaf 0100 [02]
          edge 10 QmeaHciZS5DW7C4jKgbPCsj7ZQP7JRgh1nSH4AaYh7Ts9K
            leaf 0110 [04]
          edge 2 Qmco5WuYgQB1KGZ6NrBq76g6YR3yksApNUdhqqVLcR8DkX
            branch 012
              edge 0 QmQC8dXYcujfCMkmX3Z2VMp2Rq1JcSxSKCdihUq31jk77W
                leaf 0120 [05]
              edge 1 QmTxhtdxbswwbmmc9BAN5rmnBD2D8VMATpMW6Agpc6cxgU
                leaf 0121 [06]
      edge 200 QmQLDd75hvqobp7K5oWa6rTK5cGo2EzYSfMxZnxSMNQZWG
        leaf 0200 [03]
  edge 11 QmcnCrYgiVLUcEA7j8fa1bGGyjsSUJ7ahAxRxrRHjuhCug
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
			require.NoError(t, trie.Check(context.Background()), "trie.Check()")

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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
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
  edge 1234 Qme9AssTmBYDoXZKvs8ns6kY9ExhKiEtW9X4pTFYiqnSq7
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
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
  edge 12 QmbKfbFixNmshxuxe9dxtRL9bMQa4tKXNeXC29sfk8FQUU
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
  edge 12 QmPsiacKR4KvUymPqCDMTLtmeexXN1qjB4jsNP6VhFLX9B
    branch 12 [ff00ff]
      edge 3456 QmQc672vjb6oZm2aH1CEsB1z8d1s3Khf4rfRCYqiHY7GBR
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
  edge 12 QmXP5fthqLtug1Yc97CuDzNNRzMrJXjjzjoXGuoH7JRi9o
    branch 12 [00]
      edge 34 QmdEWeq5gprJMvjTCVUM2iGdEQpVV8ZAjDPSeKdvA4nMud
        branch 1234
          edge 56 QmPJ2SBEzQNEnGVy7eNi8H1BsG6fUzZUd2ojrExPyCAC2Z
            leaf 123456 [02]
          edge 78 QmbYfDRCrMuHqeuYcFDSXtX3Aescgq21shCx8E8qnmR7na
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
			require.NoError(t, trie.Check(context.Background()), "trie.Check()")

			got, err := trie.Dump()
			require.NoError(t, err, "trie.dump()")
			assert.Equal(t, strings.TrimSpace(tt.want), strings.TrimSpace(got))
		})
	}
}
