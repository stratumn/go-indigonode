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

package model

import pb "github.com/stratumn/alice/pb/coin"

// Header is a block's header.
type Header struct {
	*pb.Header
}

// ToProtoModel converts the header service's proto model
func (h *Header) ToProtoModel() *pb.Header {
	return h.Header
}

// Transaction is a coin transaction.
type Transaction struct {
	*pb.Transaction
}

// ToProtoModel converts the header service's proto model
func (t *Transaction) ToProtoModel() *pb.Transaction {
	return t.Transaction
}

// Block is a block in the chain.
type Block struct {
	*pb.Block
}

// BlockNumber returns the block's number
func (b *Block) BlockNumber() uint64 {
	return b.Header.BlockNumber
}

// ToProtoModel converts the header service's proto model
func (b *Block) ToProtoModel() *pb.Block {
	return b.Block
}
