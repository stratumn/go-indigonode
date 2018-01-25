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

package coin

import (
	"context"
	"encoding/json"
	"testing"

	rpcpb "github.com/stratumn/alice/grpc/coin"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
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

type grpcTestCase struct {
	name          string
	tx            func() *pb.Transaction
	validate      func(*rpcpb.TransactionResp, error)
	expectedAdded bool
}

// newTransaction creates a valid transaction.
func newTransaction(t *testing.T) *pb.Transaction {
	tx := &pb.Transaction{
		From:  []byte(testSenderPID),
		To:    []byte(testRecipientPID),
		Value: 1,
		Nonce: 42,
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

func TestGRPCServer(t *testing.T) {
	testCases := []grpcTestCase{{
		"empty-tx",
		func() *pb.Transaction { return nil },
		func(txResp *rpcpb.TransactionResp, err error) {
			assert.EqualError(t, err, ErrEmptyTx.Error())
			assert.Nil(t, txResp)
		},
		false,
	}, {
		"empty-value",
		func() *pb.Transaction {
			tx := newTransaction(t)
			tx.Value = 0
			return tx
		},
		func(txResp *rpcpb.TransactionResp, err error) {
			assert.EqualError(t, err, ErrInvalidTxValue.Error())
			assert.Nil(t, txResp)
		},
		false,
	}, {
		"missing-to",
		func() *pb.Transaction {
			tx := newTransaction(t)
			tx.To = nil
			return tx
		},
		func(txResp *rpcpb.TransactionResp, err error) {
			assert.EqualError(t, err, ErrInvalidTxRecipient.Error())
			assert.Nil(t, txResp)
		},
		false,
	}, {
		"missing-from",
		func() *pb.Transaction {
			tx := newTransaction(t)
			tx.From = nil
			return tx
		},
		func(txResp *rpcpb.TransactionResp, err error) {
			assert.EqualError(t, err, ErrInvalidTxSender.Error())
			assert.Nil(t, txResp)
		},
		false,
	}, {
		"missing-signature",
		func() *pb.Transaction {
			tx := newTransaction(t)
			tx.Signature = nil
			return tx
		},
		func(txResp *rpcpb.TransactionResp, err error) {
			assert.EqualError(t, err, ErrMissingTxSignature.Error())
			assert.Nil(t, txResp)
		},
		false,
	}, {
		"valid-tx",
		func() *pb.Transaction {
			tx := newTransaction(t)
			return tx
		},
		func(txResp *rpcpb.TransactionResp, err error) {
			assert.NoError(t, err)
			assert.NotNil(t, txResp)
			assert.NotNil(t, txResp.TxHash)
		},
		true,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			added := false
			server := &grpcServer{
				func(_ *pb.Transaction) error {
					added = true
					return nil
				},
			}

			txResp, err := server.Transaction(context.Background(), tt.tx())
			tt.validate(txResp, err)
			assert.Equal(t, tt.expectedAdded, added, "added")
		})
	}
}
