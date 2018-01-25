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

//go:generate mockgen -package mockcoin -destination mockcoin/mockchainreader.go github.com/stratumn/alice/core/protocol/coin ChainReader

package coin

import (
	pb "github.com/stratumn/alice/pb/coin"
)

// ChainConfig describes the blockchain's chain configuration.
type ChainConfig struct {
}

// ChainReader defines a small collection of methods needed to access the local
// blockchain.
type ChainReader interface {
	// Config retrieves the blockchain's chain configuration.
	Config() *ChainConfig

	// CurrentHeader retrieves the current header from the local chain.
	CurrentHeader() *pb.Header

	// GetHeaderByNumber retrieves block headers from the database by number.
	// In case of forks there might be multiple headers with the same number.
	GetHeaderByNumber(number uint64) []*pb.Header

	// GetHeaderByHash retrieves a block header from the database by its hash.
	GetHeaderByHash(hash []byte) *pb.Header

	// GetBlock retrieves a block from the database by hash and number.
	GetBlock(hash []byte, number uint64) *pb.Block
}
