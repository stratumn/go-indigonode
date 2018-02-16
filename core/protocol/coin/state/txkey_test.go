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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	txKey := &TxKey{1, []byte("123")}

	txKeyB := txKey.Marshal()

	assert.Equal(t, []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x31, 0x32, 0x33}, txKeyB, "txKey.Marshal()")
}

func TestUnmarshal(t *testing.T) {
	txKey := &TxKey{}
	err := txKey.Unmarshal([]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x31, 0x32, 0x33})

	require.NoError(t, err, "txKey.Unmarshal([]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x31, 0x32, 0x33})")
	assert.Equal(t, uint64(1), txKey.TxIdx)
	assert.Equal(t, []byte("123"), txKey.BlkHash)
}

func TestUnmarshalError(t *testing.T) {
	txKey := &TxKey{}
	err := txKey.Unmarshal([]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	assert.Error(t, err, "txKey.Unmarshal([]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})")
	assert.Equal(t, ErrInvalidTxKey, err, "txKey.Unmarshal([]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})")
}
