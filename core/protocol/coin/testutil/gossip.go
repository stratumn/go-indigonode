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

package testutil

import (
	"math/rand"

	pb "github.com/stratumn/alice/pb/coin"
)

// RandomGossipTx creates a random transaction broadcast between peers.
func RandomGossipTx() *pb.Gossip {
	return &pb.Gossip{
		Msg: &pb.Gossip_Tx{
			Tx: &pb.Transaction{
				From:  []byte("Alice"),
				To:    []byte("Bob"),
				Value: rand.Int63(),
				Nonce: rand.Int63(),
			},
		},
	}
}

// RandomGossipBlock creates a random block broadcast between peers.
func RandomGossipBlock() *pb.Gossip {
	return &pb.Gossip{
		Msg: &pb.Gossip_Block{
			Block: &pb.Block{
				Header: &pb.Header{
					Version:     1,
					BlockNumber: rand.Int63(),
				},
			},
		},
	}
}
