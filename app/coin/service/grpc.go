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

package service

import (
	"context"

	"github.com/pkg/errors"
	rpcpb "github.com/stratumn/alice/app/coin/grpc"
	"github.com/stratumn/alice/app/coin/pb"
	"github.com/stratumn/alice/app/coin/protocol/coinutil"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// grpcServer is a gRPC server for the coin service.
type grpcServer struct {
	DoGetAccount           func([]byte) (*pb.Account, error)
	AddTransaction         func(*pb.Transaction) error
	GetAccountTransactions func([]byte) ([]*pb.Transaction, error)
	GetBlockchain          func(uint64, []byte, uint32) ([]*pb.Block, error)
	GetTransactionPool     func(uint32) (uint64, []*pb.Transaction, error)
}

// Account returns an account.
func (s grpcServer) GetAccount(ctx context.Context, req *rpcpb.AccountReq) (*pb.Account, error) {
	// Just for validation.
	_, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s.DoGetAccount(req.PeerId)
}

// Transaction sends a coin transaction to the consensus engine.
func (s grpcServer) SendTransaction(ctx context.Context, req *pb.Transaction) (response *rpcpb.TransactionResp, err error) {
	err = s.AddTransaction(req)
	if err != nil {
		return nil, err
	}

	txHash, err := coinutil.HashTransaction(req)
	if err != nil {
		return nil, err
	}

	txResponse := &rpcpb.TransactionResp{
		TxHash: txHash,
	}

	return txResponse, nil
}

func (s grpcServer) AccountTransactions(req *rpcpb.AccountTransactionsReq, ss rpcpb.Coin_AccountTransactionsServer) error {
	_, err := peer.IDFromBytes(req.PeerId)
	if err != nil {
		return errors.WithStack(err)
	}

	txs, err := s.GetAccountTransactions(req.PeerId)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, txs := range txs {
		err := ss.Send(txs)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s grpcServer) TransactionPool(ctx context.Context, req *rpcpb.TransactionPoolReq) (*rpcpb.TransactionPoolResp, error) {
	txCount, txs, err := s.GetTransactionPool(req.Count)
	if err != nil {
		return nil, err
	}

	return &rpcpb.TransactionPoolResp{Count: txCount, Txs: txs}, nil
}

func (s grpcServer) Blockchain(ctx context.Context, req *rpcpb.BlockchainReq) (*rpcpb.BlockchainResp, error) {
	blocks, err := s.GetBlockchain(req.BlockNumber, req.HeaderHash, req.Count)
	if err != nil {
		return nil, err
	}

	return &rpcpb.BlockchainResp{Blocks: blocks}, nil
}
