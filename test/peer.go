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

package test

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
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
