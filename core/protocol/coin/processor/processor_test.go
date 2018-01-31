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

package processor

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/coin/state"
	"github.com/stratumn/alice/core/protocol/coin/testutil"
	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_Process(t *testing.T) {
	alice := []byte("alice")
	bob := []byte("bob")
	charlie := []byte("charlie")

	p := NewProcessor()
	s := testutil.NewSimpleState(t)

	assert.NoError(t, s.SetBalance(alice, 20), "s.SetBalance(alice)")

	block := &pb.Block{
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    bob,
			Value: 10,
		}, {
			From:  bob,
			To:    charlie,
			Value: 5,
		}, {
			From:  charlie,
			To:    alice,
			Value: 2,
		}},
	}

	assert.NoError(t, p.Process(block, s, nil))

	v, err := s.GetBalance(alice)
	assert.NoError(t, err, "s.GetBalance(alice)")
	assert.EqualValues(t, 20-10+2, v, "s.GetBalance(alice)")

	v, err = s.GetBalance(bob)
	assert.NoError(t, err, "s.GetBalance(bob)")
	assert.EqualValues(t, 10-5, v, "s.GetBalance(bob)")

	v, err = s.GetBalance(charlie)
	assert.NoError(t, err, "s.GetBalance(charlie)")
	assert.EqualValues(t, 5-2, v, "s.GetBalance(charlie)")
}

func TestProcessor_Process_amountTooBig(t *testing.T) {
	alice := []byte("alice")
	bob := []byte("bob")

	p := NewProcessor()
	s := testutil.NewSimpleState(t)

	assert.NoError(t, s.SetBalance(alice, 20), "s.SetBalance(alice)")

	block := &pb.Block{
		Transactions: []*pb.Transaction{{
			From:  alice,
			To:    bob,
			Value: 10,
		}, {
			From:  alice,
			To:    bob,
			Value: 11,
		}},
	}

	assert.EqualError(t, p.Process(block, s, nil), state.ErrAmountTooBig.Error())

	// Make sure state wasn't updated.
	v, err := s.GetBalance(alice)
	assert.NoError(t, err, "s.GetBalance(alice)")
	assert.EqualValues(t, 20, v, "s.GetBalance(alice)")

	v, err = s.GetBalance(bob)
	assert.NoError(t, err, "s.GetBalance(bob)")
	assert.EqualValues(t, 0, v, "s.GetBalance(bob)")
}
