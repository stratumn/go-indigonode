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

//go:generate mockgen -package mockvalidator -destination mockvalidator/mockvalidator.go github.com/stratumn/alice/core/protocol/coin/validator Validator

package validator

import (
	"github.com/stratumn/alice/core/protocol/coin/state"
	pb "github.com/stratumn/alice/pb/coin"
)

// Validator is an interface which defines the standard for block and
// transaction validation.
// It is only responsible for validating the block contents, as the header
// validation is done by the specific consensus engines.
type Validator interface {
	// ValidateTx validates a transaction.
	// If state is nil, ValidateTx only validates that the
	// transaction is well-formed and properly signed.
	ValidateTx(tx *pb.Transaction, state *state.Reader) error
	// ValidateBlock validates the contents of a block.
	ValidateBlock(block *pb.Block, state *state.Reader) error
}
