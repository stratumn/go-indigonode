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

package validator

import (
	"bytes"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/app/coin/pb"
	"github.com/stratumn/alice/app/coin/protocol/chain"
	"github.com/stratumn/alice/app/coin/protocol/coinutil"
	"github.com/stratumn/alice/app/coin/protocol/engine"
	"github.com/stratumn/alice/app/coin/protocol/state"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
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

	// ErrInvalidTxSignature is returned when the transaction signature is invalid.
	ErrInvalidTxSignature = errors.New("invalid tx signature")

	// ErrTxSignatureNotHandled is returned when the transaction signature scheme isn't implemented.
	ErrTxSignatureNotHandled = errors.New("tx signature scheme not supported yet")

	// ErrTooManyTxs is returned when the sender tries to put too many transactions in a block.
	ErrTooManyTxs = errors.New("too many txs in proposed block")

	// ErrInvalidMerkleRoot is returned when the merkle root doesn't represent the block transactions.
	ErrInvalidMerkleRoot = errors.New("invalid merkle root")

	// ErrMultipleMinerRewards is returned when a block contains multiple miner rewards.
	ErrMultipleMinerRewards = errors.New("only one miner reward transaction is allowed per block")

	// ErrInvalidMinerReward is returned when the miner reward is invalid.
	ErrInvalidMinerReward = errors.New("miner reward is invalid")

	// ErrBlockHeightZero is returned we receive a block with height 0.
	ErrBlockHeightZero = errors.New("block should have height higher than 0")
)

// Validator is an interface which defines the standard for block and
// transaction validation.
// It is only responsible for validating the block contents, as the header
// validation is done by the specific consensus engines.
type Validator interface {
	// MaxTxPerBlock returns the maximum number of transactions
	// allowed in a block.
	MaxTxPerBlock() uint32
	// ValidateTx validates a transaction.
	// If state is nil, ValidateTx only validates that the
	// transaction is well-formed and properly signed.
	ValidateTx(tx *pb.Transaction, state state.Reader) error
	// ValidateBlock validates the contents of a block.
	ValidateBlock(block *pb.Block, state state.Reader) error
	// ValidateTransactions validates a list of transactions.
	ValidateTransactions(transactions []*pb.Transaction, state state.Reader) error
}

// GossipValidator validates everything including the block header except balances.
type GossipValidator struct {
	chReader chain.Reader
	*BalanceValidator
}

// ValidateTx validates that the transaction is well-formed and properly signed.
func (g *GossipValidator) ValidateTx(tx *pb.Transaction, state state.Reader) error {
	return g.BalanceValidator.ValidateTx(tx, nil)
}

// ValidateBlock validates the transactions contained in a block and the block header.
func (g *GossipValidator) ValidateBlock(block *pb.Block, state state.Reader) error {
	if err := g.BalanceValidator.engine.VerifyHeader(g.chReader, block.Header); err != nil {
		return err
	}
	return g.BalanceValidator.ValidateBlock(block, nil)
}

// ValidateTransactions validates that transactions are well-formed and properly signed.
func (g *GossipValidator) ValidateTransactions(transactions []*pb.Transaction, state state.Reader) error {
	return g.BalanceValidator.ValidateTransactions(transactions, nil)
}

// NewGossipValidator returns a GossipValidator.
func NewGossipValidator(maxTxPerBlock uint32, engine engine.PoW, chReader chain.Reader) Validator {
	balanceValidator := NewBalanceValidator(maxTxPerBlock, engine).(*BalanceValidator)

	return &GossipValidator{chReader, balanceValidator}
}

// BalanceValidator validates coin transactions.
// It verifies that transactions are well-formed and signed,
// and that users don't spend more coins than they have.
type BalanceValidator struct {
	engine        engine.PoW
	maxTxPerBlock uint32
}

// NewBalanceValidator creates a BalanceValidator.
func NewBalanceValidator(maxTxPerBlock uint32, engine engine.PoW) Validator {
	return &BalanceValidator{
		engine:        engine,
		maxTxPerBlock: maxTxPerBlock,
	}
}

// MaxTxPerBlock returns the maximum number of transactions
// allowed in a block.
func (v *BalanceValidator) MaxTxPerBlock() uint32 {
	return v.maxTxPerBlock
}

// ValidateTx validates a transaction.
// If state is nil, ValidateTx only validates that the
// transaction is well-formed and properly signed.
func (v *BalanceValidator) ValidateTx(tx *pb.Transaction, s state.Reader) error {
	err := v.validateFormat(tx)
	if err != nil {
		return err
	}

	err = v.validateSignature(tx)
	if err != nil {
		return err
	}

	if s != nil {
		err = state.ValidateBalance(s, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateFormat validates the format of the incoming tx (fields, etc).
func (v *BalanceValidator) validateFormat(tx *pb.Transaction) error {
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

	if bytes.Equal(tx.To, tx.From) {
		return ErrInvalidTxRecipient
	}

	if tx.Signature == nil || tx.Signature.PublicKey == nil || tx.Signature.Signature == nil {
		return ErrMissingTxSignature
	}

	return nil
}

// validateSignature validates transaction signature.
func (v *BalanceValidator) validateSignature(tx *pb.Transaction) error {
	// Extract the payload part
	payload := &pb.Transaction{
		From:  tx.From,
		To:    tx.To,
		Value: tx.Value,
		Fee:   tx.Fee,
		Nonce: tx.Nonce,
	}

	b, err := payload.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}

	senderID, err := peer.IDFromBytes(tx.From)
	if err != nil {
		return errors.WithStack(err)
	}

	var validSig bool
	switch tx.Signature.KeyType {
	case pb.KeyType_Ed25519:
		sigKey, err := ic.UnmarshalPublicKey(tx.Signature.PublicKey)
		if err != nil {
			return errors.WithStack(err)
		}

		if !senderID.MatchesPublicKey(sigKey) {
			return ErrInvalidTxSignature
		}

		valid, err := sigKey.Verify(b, tx.Signature.Signature)
		if err != nil {
			return errors.WithStack(err)
		}

		validSig = valid
	default:
		return ErrTxSignatureNotHandled
	}

	if !validSig {
		return ErrInvalidTxSignature
	}

	return nil
}

// ValidateBlock validates a block.
func (v *BalanceValidator) ValidateBlock(block *pb.Block, s state.Reader) error {
	// Reject blocks with height 0: we do not accept another genesis block.
	if block.BlockNumber() == 0 {
		return ErrBlockHeightZero
	}

	if err := v.validateTxCount(block); err != nil {
		return err
	}

	if err := v.validateMerkleRoot(block); err != nil {
		return err
	}

	if err := v.validateMinerReward(block); err != nil {
		return err
	}

	return v.ValidateTransactions(block.Transactions, s)
}

// ValidateTransactions validates the transactions contained in a block.
func (v *BalanceValidator) ValidateTransactions(transactions []*pb.Transaction, s state.Reader) error {
	for _, tx := range transactions {
		if tx.From != nil {
			// Validate everything else than balance.
			err := v.ValidateTx(tx, nil)
			if err != nil {
				return err
			}
		}
	}

	if s != nil {
		return state.ValidateBalances(s, transactions)
	}

	return nil
}

// validateTxCount restricts the number of transactions that can be
// included in a block.
func (v *BalanceValidator) validateTxCount(block *pb.Block) error {
	if v.maxTxPerBlock < coinutil.GetTxCount(block) {
		return ErrTooManyTxs
	}

	return nil
}

// validateMerkleRoot verifies that the merkle root correctly
// hashes the block's transactions.
func (v *BalanceValidator) validateMerkleRoot(block *pb.Block) error {
	if len(block.Transactions) == 0 {
		return nil
	}

	expected, err := coinutil.TransactionRoot(block.Transactions)
	if err != nil {
		return err
	}

	if !bytes.Equal(expected, block.Header.MerkleRoot) {
		return ErrInvalidMerkleRoot
	}

	return nil
}

// validateMinerReward verifies that the miner reward (if included) is valid.
func (v *BalanceValidator) validateMinerReward(block *pb.Block) error {
	minerReward, err := coinutil.GetMinerReward(block)
	if err != nil {
		if err == coinutil.ErrMultipleMinerRewards {
			return ErrMultipleMinerRewards
		}

		return err
	}

	if minerReward != nil {
		totalFees := coinutil.GetBlockFees(block)
		if totalFees+v.engine.Reward() < minerReward.Value {
			return ErrInvalidMinerReward
		}
	}

	return nil
}
