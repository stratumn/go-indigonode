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

package testutil

import (
	"encoding/json"
	"testing"

	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	ic "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

var testSenderPrivateKey ic.PrivKey
var testSenderPublicKey ic.PubKey
var testSenderPID peer.ID
var testRecipientPID peer.ID

func init() {
	sk, pk, err := ic.GenerateKeyPair(ic.Ed25519, 0)
	if err != nil {
		panic(err)
	}

	testSenderPrivateKey = sk
	testSenderPublicKey = pk

	pid, err := peer.IDFromPrivateKey(testSenderPrivateKey)
	if err != nil {
		panic(err)
	}

	testSenderPID = pid

	pid, err = peer.IDB58Decode("QmYgPDzgtjRQFC4poxTD8fgVnVVf3wHQPjrmyph8Wy8ZmQ")
	if err != nil {
		panic(err)
	}

	testRecipientPID = pid
}

// NewTransaction creates a properly signed transaction.
func NewTransaction(t *testing.T, value, nonce int64) *pb.Transaction {
	tx := &pb.Transaction{
		From:  []byte(testSenderPID),
		To:    []byte(testRecipientPID),
		Value: value,
		Nonce: nonce,
	}

	b, err := json.Marshal(tx)
	require.NoError(t, err, "json.Marshal(tx)")

	sig, err := testSenderPrivateKey.Sign(b)
	require.NoError(t, err, "testSenderPrivateKey.Sign(b)")

	pk, err := testSenderPublicKey.Bytes()
	require.NoError(t, err, "testSenderPublicKey.Bytes()")

	tx.Signature = &pb.Signature{
		KeyType:   pb.KeyType_Ed25519,
		PublicKey: pk,
		Signature: sig,
	}

	return tx
}
