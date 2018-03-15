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
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/stratumn/alice/core/protocol/coin/coinutil"
	pb "github.com/stratumn/alice/pb/coin"

	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
)

// GetGenesisBlock returns the default genesis block.
func GetGenesisBlock() (*pb.Block, error) {
	stefan, err := peer.IDB58Decode("QmYabPJqc6WQWPdE46ttToyuWEwMhxupMQa65p5FBR1BRv")
	if err != nil {
		return nil, err
	}
	tbast, err := peer.IDB58Decode("QmQnYf23kQ7SvuPZ3mQcg3RuJMr9E39fBvm89Nz4bevJdt")
	if err != nil {
		return nil, err
	}
	bejito, err := peer.IDB58Decode("QmaYTBbpANr4gooRRJXBhxnfS9sC8ChQCYnMqwX8cPTHqH")
	if err != nil {
		return nil, err
	}
	simon, err := peer.IDB58Decode("QmUpqc24f4RViYmmXxXrLrFQ6vzDW1M3dsu65KCa7ZJ159")
	if err != nil {
		return nil, err
	}
	alex, err := peer.IDB58Decode("QmTJBbyuh3EFeoiYQF6FWU17og7YG52UfnBwTEcbigGhgW")
	if err != nil {
		return nil, err
	}
	such, err := peer.IDB58Decode("Qmcso1m2v6r9jZv8swVD7LCDPNzRoX8RohHWYqG8dP8cxr")
	if err != nil {
		return nil, err
	}

	txs := []*pb.Transaction{
		&pb.Transaction{
			To:    []byte(stefan),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(tbast),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(bejito),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(alex),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(such),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(simon),
			Value: uint64(41999),
		},
	}

	merkleRoot, err := coinutil.TransactionRoot(txs)
	if err != nil {
		return nil, err
	}

	// 2018-02-23 10:00:00
	ts, err := ptypes.TimestampProto(time.Unix(1519376400, 0))
	if err != nil {
		return nil, err
	}

	// GenesisBlock is the genesis block.
	genesisBlock := &pb.Block{
		Header: &pb.Header{
			Nonce:      42,
			Version:    1,
			MerkleRoot: merkleRoot,
			Timestamp:  ts,
		},
		Transactions: txs,
	}

	return genesisBlock, nil
}
