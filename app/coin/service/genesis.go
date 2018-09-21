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

package service

import (
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/coinutil"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
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
	jeremie, err := peer.IDB58Decode("QmfM51csY6bXJZXsuiFocV8nEVM11UKX6yyCwQPKCTeqQ9")
	if err != nil {
		return nil, err
	}
	conor, err := peer.IDB58Decode("QmYgntWu4v6JJPDd2Luw8V6pQRh7SaKAMRSFbQaKFRb7ra")
	if err != nil {
		return nil, err
	}
	pierre, err := peer.IDB58Decode("QmbsW6Rs9cpcupQk3kzHUsLRGA3c6eroenoAWwRuRMFMur")
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
			Value: uint64(42001),
		},
		&pb.Transaction{
			To:    []byte(conor),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(jeremie),
			Value: uint64(42000),
		},
		&pb.Transaction{
			To:    []byte(pierre),
			Value: uint64(42000),
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
