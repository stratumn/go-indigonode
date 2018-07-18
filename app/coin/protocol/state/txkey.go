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

package state

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

var (
	// ErrInvalidTxKey is returned when the transaction key cannot
	// be unmarshalled
	ErrInvalidTxKey = errors.New("txkey byte representation is invalid")
)

// TxKey is used to save a user transactions.
type TxKey struct {
	TxIdx   uint64
	BlkHash []byte
}

// Marshal converts the struct into a byte slice.
func (t *TxKey) Marshal() []byte {
	return append(encodeUint64(t.TxIdx), t.BlkHash...)
}

// Unmarshal creates a TxKey from a byte slice.
func (t *TxKey) Unmarshal(data []byte) error {
	if len(data) < 9 {
		return ErrInvalidTxKey
	}

	t.TxIdx = binary.LittleEndian.Uint64(data[:8])
	t.BlkHash = data[8:]

	return nil
}
