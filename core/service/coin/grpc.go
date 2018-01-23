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
	"crypto/sha256"
	"encoding/json"

	"github.com/pkg/errors"
	rpcpb "github.com/stratumn/alice/grpc/coin"
	pb "github.com/stratumn/alice/pb/coin"
)

var (
	// ErrEmptyTx is returned when the transaction is nil.
	ErrEmptyTx = errors.New("tx is empty")

	// ErrInvalidTxValue is returned when the transaction value is 0.
	ErrInvalidTxValue = errors.New("invalid tx value")

	// ErrInvalidTxSender is returned when the transaction sender is invalid.
	ErrInvalidTxSender = errors.New("invalid tx sender")

	// ErrInvalidTxRecipient is returned when the transaction recipient is invalid.
	ErrInvalidTxRecipient = errors.New("invalid tx recipient")

	// ErrMissingTxSignature is returned when the transaction signature is missing.
	ErrMissingTxSignature = errors.New("missing tx signature")
)

// grpcServer is a gRPC server for the coin service.
type grpcServer struct {
}

// Transaction sends a coin transaction to the consensus engine.
func (s grpcServer) Transaction(ctx context.Context, req *pb.Transaction) (response *rpcpb.TransactionResp, err error) {
	if err := s.validateTx(req); err != nil {
		return nil, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	txHash := sha256.Sum256(b)
	txResponse := &rpcpb.TransactionResp{
		TxHash: txHash[:],
	}

	return txResponse, nil
}

// validateTx validates a transaction from the point of view of the RPC server.
// It doesn't validate signatures, it's the protocol's responsibility to do it.
// Otherwise we would validate signatures twice for no good reason.
func (s grpcServer) validateTx(tx *pb.Transaction) error {
	if tx == nil {
		return ErrEmptyTx
	}

	if tx.Value <= 0 {
		return ErrInvalidTxValue
	}

	if tx.From == nil {
		return ErrInvalidTxSender
	}

	if tx.To == nil {
		return ErrInvalidTxRecipient
	}

	if tx.Signature == nil || tx.Signature.PublicKey == nil || tx.Signature.Signature == nil {
		return ErrMissingTxSignature
	}

	return nil
}
