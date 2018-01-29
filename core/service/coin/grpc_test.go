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
	"testing"

	pb "github.com/stratumn/alice/pb/coin"
	"github.com/stretchr/testify/assert"
)

func TestGRPCServer(t *testing.T) {
	added := false
	server := &grpcServer{
		func(_ *pb.Transaction) error {
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

	assert.NoError(t, err, "server.Tranaction()")
	assert.NotNil(t, txResp.TxHash, "TransactionResp.TxHash")
	assert.True(t, added, "added")
}
