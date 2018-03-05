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
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	"testing"

	"github.com/golang/mock/gomock"
	rpcpb "github.com/stratumn/alice/grpc/coin"
	"github.com/stratumn/alice/grpc/coin/mockcoin"
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
		GetAccount: func([]byte) (*pb.Account, error) {
			return account, nil
		},
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
		AddTransaction: func(_ *pb.Transaction) error {
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

func TestGRPCServer_AccountTransactions(t *testing.T) {
	pid, err := peer.IDB58Decode(testPID)
	require.NoError(t, err, "peer.IDB58Decode(testPID)")

	tx := &pb.Transaction{
		Value: 42,
	}

	server := &grpcServer{
		GetAccountTransactions: func([]byte) ([]*pb.Transaction, error) {
			return []*pb.Transaction{tx}, nil
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	req, ss := &rpcpb.AccountTransactionsReq{
		PeerId: []byte(pid),
	}, mockcoin.NewMockCoin_AccountTransactionsServer(ctrl)

	ss.EXPECT().Send(&pb.Transaction{
		Value: 42,
	})

	assert.NoError(t, server.AccountTransactions(req, ss))
}

func TestGRPCServer_Blockchain(t *testing.T) {
	server := &grpcServer{
		GetBlockchain: func(uint64, []byte, uint32) ([]*pb.Block, error) {
			return []*pb.Block{
				&pb.Block{Header: &pb.Header{BlockNumber: 13, Nonce: 2}},
				&pb.Block{Header: &pb.Header{BlockNumber: 14, Nonce: 4}},
			}, nil
		},
	}

	blockchain, err := server.Blockchain(
		context.Background(),
		&rpcpb.BlockchainReq{BlockNumber: 14, Count: 2},
	)

	assert.NoError(t, err, "server.Blockchain()")
	assert.Len(t, blockchain.Blocks, 2, "blockchain.Blocks")
	assert.Equal(t, blockchain.Blocks[0].BlockNumber(), uint64(13), "BlockNumber()")
	assert.Equal(t, blockchain.Blocks[1].BlockNumber(), uint64(14), "BlockNumber()")
}

func TestGRPCServer_TransactionPool(t *testing.T) {
	server := &grpcServer{
		GetTransactionPool: func(uint32) (uint64, []*pb.Transaction, error) {
			return 42, []*pb.Transaction{
				&pb.Transaction{Value: 12},
				&pb.Transaction{Value: 15},
			}, nil
		},
	}

	txPool, err := server.TransactionPool(
		context.Background(),
		&rpcpb.TransactionPoolReq{Count: 2},
	)

	assert.NoError(t, err, "server.TransactionPool()")
	assert.Equal(t, uint64(42), txPool.Count, "Count")
	assert.Len(t, txPool.Txs, 2, "txPool.Txs")
	assert.Equal(t, uint64(12), txPool.Txs[0].Value, "Value")
	assert.Equal(t, uint64(15), txPool.Txs[1].Value, "Value")
}
