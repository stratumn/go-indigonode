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

package bootstrap

import (
	"context"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// PublicNetworkHandler is the handler for public networks.
// Public networks are completely open, so this handler doesn't do anything.
type PublicNetworkHandler struct{}

// AddNode can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) AddNode(context.Context, peer.ID, multiaddr.Multiaddr, []byte) error {
	return ErrInvalidOperation
}

// Accept can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) Accept(context.Context, peer.ID) error {
	return ErrInvalidOperation
}

// Reject can't be used in a public network.
// You can directly connect to any node freely.
func (h *PublicNetworkHandler) Reject(context.Context, peer.ID) error {
	return ErrInvalidOperation
}

// CompleteBootstrap can't be used in a public network.
// There is no bootstrapping phase in such networks.
func (h *PublicNetworkHandler) CompleteBootstrap(context.Context) error {
	return ErrInvalidOperation
}

// Close doesn't do anything.
func (h *PublicNetworkHandler) Close(context.Context) {}
