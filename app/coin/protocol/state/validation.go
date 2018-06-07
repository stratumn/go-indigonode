package state

import (
	"github.com/pkg/errors"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"

	"github.com/stratumn/alice/app/coin/pb"
)

var (
	// ErrInvalidTxNonce is returned when the transaction nonce isn't greater
	// than the last nonce used by the sender's account.
	ErrInvalidTxNonce = errors.New("invalid tx nonce")

	// ErrInsufficientBalance is returned when the sender tries to send more coins than he has.
	ErrInsufficientBalance = errors.New("tx sender does not have enough coins to send")
)

// ValidateBalance validates transaction nonce and balance against
// a given state.
func ValidateBalance(s Reader, tx *pb.Transaction) error {
	account, err := s.GetAccount(tx.From)
	if err != nil {
		return err
	}

	if tx.Nonce <= account.Nonce {
		return ErrInvalidTxNonce
	}

	if account.Balance < tx.Value+tx.Fee {
		return ErrInsufficientBalance
	}

	return nil
}

// ValidateBalances validates transactions nonce and balance of a
// block against a given state.
func ValidateBalances(s Reader, transactions []*pb.Transaction) error {
	// Aggregate transactions from the same sender and verify balance.
	txs := make(map[peer.ID]*pb.Account)
	for _, tx := range transactions {
		if tx.From == nil {
			continue
		}

		// We need to validate each Tx individually as well so that balance
		// cannot wrap around because of int overflow.
		// Later we could use a more efficient implementation taking into
		// account received txs and ordering by nonce.
		if err := ValidateBalance(s, tx); err != nil {
			return err
		}

		senderID, err := peer.IDFromBytes(tx.From)
		if err != nil {
			return errors.WithStack(err)
		}

		_, ok := txs[senderID]
		if !ok {
			txs[senderID] = &pb.Account{}
		}

		account := txs[senderID]
		txs[senderID] = &pb.Account{
			Balance: account.Balance + tx.Value + tx.Fee,
			Nonce:   tx.Nonce,
		}
	}

	for from, val := range txs {
		err := ValidateBalance(s, &pb.Transaction{
			From:  []byte(from),
			Value: val.Balance,
			Nonce: val.Nonce,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
