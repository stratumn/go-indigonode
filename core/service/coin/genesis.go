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

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
)

// GetGenesisBlock returns the default genesis block.
func GetGenesisBlock() (*pb.Block, error) {
	pid, err := peer.IDB58Decode("QmXivMekek9JBn3fLTuQBwUuUqiCZYkzkw2uU5ZEFFEmhU")
	if err != nil {
		return nil, err
	}

	txs := []*pb.Transaction{
		&pb.Transaction{
			To:    []byte(pid),
			Value: uint64(9000),
		},
		&pb.Transaction{
			To:    []byte(pid),
			Value: uint64(10000),
		},
		&pb.Transaction{
			To:    []byte(pid),
			Value: uint64(11000),
		},
		&pb.Transaction{
			To:    []byte(pid),
			Value: uint64(12000),
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
