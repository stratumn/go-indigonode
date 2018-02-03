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
	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
	"testing"

	rpcpb "github.com/stratumn/alice/grpc/coin"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPID = "QmVhJVRSYHNSHgR9dJNbDxu6G7GPPqJAeiJoVRvcexGNf9"

func TestGRPCServer_GetAccount(t *testing.T) {
	pid, err := peer.IDB58Decode(testPID)
	require.NoError(t, err, "peer.IDB58Decode(testPID)")

	account := &pb.Account{
		Balance: 10,
		Nonce:   1,
	}

	server := &grpcServer{
		func([]byte) (*pb.Account, error) {
			return account, nil
		},
		nil,
	}

	res, err := server.Account(context.Background(), &rpcpb.AccountReq{
		PeerId: []byte(pid),
	})
	require.NoError(t, err, "server.GetAccount()")
	require.Equal(t, account, res)
}

func TestGRPCServer_Transaction(t *testing.T) {
	added := false
	server := &grpcServer{
		nil,
		func(_ *pb.Transaction) error {
			added = true
			return nil
		},
	}

	txResp, err := server.Transaction(
		context.Background(),
		&pb.Transaction{
			Value: 42,
			Nonce: 42,
		})

	assert.NoError(t, err, "server.Transaction()")
	assert.NotNil(t, txResp.TxHash, "TransactionResp.TxHash")
	assert.True(t, added, "added")
}
