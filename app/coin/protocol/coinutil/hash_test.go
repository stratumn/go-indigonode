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

package coinutil

import (
	"encoding/hex"
	"testing"

	"github.com/stratumn/go-node/app/coin/pb"
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

	expected, _ := hex.DecodeString("80cc5e94aa7c0face84f77261516435047fc84e53cd8cf4bde57c8c8de56582c")
	assert.NoError(t, err, "TransactionRoot()")
	assert.Equal(t, expected, tr, "TransactionRoot()")
}
