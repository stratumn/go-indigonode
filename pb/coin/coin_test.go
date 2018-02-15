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

package coin_test

import (
	"testing"

	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

func TestBlockModel(t *testing.T) {
	b := &pb.Block{}
	assert.Equal(t, uint64(0), b.BlockNumber(), "b.BlockNumber()")
	assert.Equal(t, []byte{}, b.PreviousHash(), "b.PreviousHash()")

	b = &pb.Block{Header: &pb.Header{BlockNumber: 42, PreviousHash: []byte("pizza")}}
	assert.Equal(t, b.Header.BlockNumber, b.BlockNumber(), "b.BlockNumber()")
	assert.Equal(t, b.Header.PreviousHash, b.PreviousHash(), "b.PreviousHash()")
}
