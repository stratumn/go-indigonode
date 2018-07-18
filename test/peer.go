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

package test

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

// GeneratePrivateKey generates a valid private key.
func GeneratePrivateKey(t *testing.T) crypto.PrivKey {
	sk, _, err := crypto.GenerateEd25519Key(crand.Reader)
	require.NoError(t, err, "crypto.GenerateEd25519Key()")

	return sk
}

// GetPeerIDFromKey returns the peer ID associated to a key.
func GetPeerIDFromKey(t *testing.T, sk crypto.PrivKey) peer.ID {
	peerID, err := peer.IDFromPrivateKey(sk)
	require.NoError(t, err, "peer.IDFromPrivateKey()")

	return peerID
}

// GeneratePeerID generates a valid peerID.
func GeneratePeerID(t *testing.T) peer.ID {
	sk := GeneratePrivateKey(t)
	return GetPeerIDFromKey(t, sk)
}

// GeneratePeerMultiaddr generates a semantically valid multiaddr for the given peer.
func GeneratePeerMultiaddr(t *testing.T, peerID peer.ID) multiaddr.Multiaddr {
	peerAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf(
		"/ip4/10.0.0.42/tcp/%d/ipfs/%s",
		8900+rand.Intn(10000),
		peerID.Pretty(),
	))
	require.NoError(t, err, "multiaddr.NewMultiaddr()")

	return peerAddr
}

// GeneratePeerMultiaddrs generates semantically valid multiaddrs for the given peer.
func GeneratePeerMultiaddrs(t *testing.T, peerID peer.ID) []multiaddr.Multiaddr {
	return []multiaddr.Multiaddr{
		GeneratePeerMultiaddr(t, peerID),
		GeneratePeerMultiaddr(t, peerID),
	}
}

// GenerateMultiaddr generates a valid multiaddr.
func GenerateMultiaddr(t *testing.T) multiaddr.Multiaddr {
	peerID := GeneratePeerID(t)
	return GeneratePeerMultiaddr(t, peerID)
}

// GenerateMultiaddrs generates semantically valid multiaddrs.
func GenerateMultiaddrs(t *testing.T) []multiaddr.Multiaddr {
	return []multiaddr.Multiaddr{
		GenerateMultiaddr(t),
		GenerateMultiaddr(t),
	}
}
