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

package coinutil

import (
	"encoding/hex"
	"testing"

	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

func Test_HashHeader(t *testing.T) {
	header := &pb.Header{Nonce: 42}

	h, err := HashHeader(header)
	expected, _ := hex.DecodeString("1220f18e9301eae406bbc26e46022df9bdf29f1b6393f8a2f47f03961797a6897d4e")

	assert.NoError(t, err, "HashHeader()")
	assert.Equal(t, expected, []byte(h), "HashHeader()")
}

func Test_HashTransaction(t *testing.T) {
	transaction := &pb.Transaction{Nonce: 42}

	tx, err := HashTransaction(transaction)
	expected, _ := hex.DecodeString("122064e752669b24739d5a68c49d86c86ee8371eacbb9e61a46741f62f2f9d4ab4f4")

	assert.NoError(t, err, "HashHeader()")
	assert.Equal(t, expected, []byte(tx), "HashHeader()")
}

func Test_TransactionRoot(t *testing.T) {
	tx1 := &pb.Transaction{Nonce: 1}
	tx2 := &pb.Transaction{Nonce: 2}
	tx3 := &pb.Transaction{Nonce: 3}

	txs := []*pb.Transaction{tx1, tx2, tx3}

	tr, err := TransactionRoot(txs)
	expected, _ := hex.DecodeString("8598c431bbdfecab85e06c4d6450aeed500bbeb60e51f3035ebeaf7cd6478f83")
	assert.NoError(t, err, "TransactionRoot()")
	assert.Equal(t, expected, tr, "TransactionRoot()")

}
