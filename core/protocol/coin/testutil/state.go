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
	"sync"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/protocol/coin/state"

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
)

// SimpleState is a simple implementation of the State interface.
// It allows tests to easily customize users' accounts.
type SimpleState struct {
	mu       sync.RWMutex
	accounts map[peer.ID]state.Account
}

// NewSimpleState returns a SimpleState ready to use in tests.
func NewSimpleState() *SimpleState {
	return &SimpleState{
		accounts: make(map[peer.ID]state.Account),
	}
}

// GetAccount returns the account of a peer.
func (s *SimpleState) GetAccount(pubKey []byte) state.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, err := peer.IDFromBytes(pubKey)
	if err != nil {
		return state.Account{}
	}

	return s.accounts[id]
}

// UpdateAccount updates a user account.
func (s *SimpleState) UpdateAccount(pubKey []byte, balance uint64, nonce uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := peer.IDFromBytes(pubKey)
	if err != nil {
		return errors.WithStack(err)
	}

	s.accounts[id] = state.Account{
		Balance: balance,
		Nonce:   nonce,
	}

	return nil
}
