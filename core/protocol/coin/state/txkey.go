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
