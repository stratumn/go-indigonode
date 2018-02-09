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
	expected, _ := hex.DecodeString("1220cfa28753bfaab814b6d79c4cd95898949d6aba143d03dc5d39d567c3f3de1ee3")

	assert.NoError(t, err, "HashHeader()")
	assert.Equal(t, expected, []byte(tx), "HashHeader()")
}

func Test_TransactionRoot(t *testing.T) {
	tx1 := &pb.Transaction{Nonce: 1}
	tx2 := &pb.Transaction{Nonce: 2}
	tx3 := &pb.Transaction{Nonce: 3}

	txs := []*pb.Transaction{tx1, tx2, tx3}

	tr, err := TransactionRoot(txs)
	expected, _ := hex.DecodeString("e1bb422f51063ab9f4583bf25683e88d22debdd04a81e3710ac841f9706721df")
	assert.NoError(t, err, "TransactionRoot()")
	assert.Equal(t, expected, tr, "TransactionRoot()")
}
