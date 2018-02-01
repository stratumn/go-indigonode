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

package state

// State stores users' account balances.
type State interface {
	Reader
	Writer
}

// Account describes a user account.
type Account struct {
	// Balance is the number of coins the user has.
	Balance uint64
	// Nonce is the latest transaction nonce seen by the system.
	// It can only increase.
	Nonce uint64
}

// Reader gives read access to users' account balances.
type Reader interface {
	// GetAccount gets the account details of a user identified
	// by his public key.
	GetAccount(pubKey []byte) Account
}

// Writer gives write access to users' account balances.
type Writer interface {
}
