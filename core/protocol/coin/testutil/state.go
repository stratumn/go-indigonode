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

	peer "gx/ipfs/Qma7H6RW8wRrfZpNSXwxYGcd1E149s42FpWNpDNieSVrnU/go-libp2p-peer"
)

// SimpleState is a simple implementation of the State interface.
// It allows tests to easily customize users' balances.
type SimpleState struct {
	mu       sync.RWMutex
	balances map[peer.ID]uint64
}

// NewSimpleState returns a SimpleState ready to use in tests.
func NewSimpleState() *SimpleState {
	return &SimpleState{
		balances: make(map[peer.ID]uint64),
	}
}

// GetBalance returns the balance of a peer.
func (s *SimpleState) GetBalance(pubKey []byte) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, err := peer.IDFromBytes(pubKey)
	if err != nil {
		return 0
	}

	return s.balances[id]
}

// AddBalance adds coins to a user account.
func (s *SimpleState) AddBalance(pubKey []byte, amount uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := peer.IDFromBytes(pubKey)
	if err != nil {
		return errors.WithStack(err)
	}

	s.balances[id] += amount
	return nil
}
