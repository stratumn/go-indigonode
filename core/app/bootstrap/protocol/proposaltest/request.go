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

package proposaltest

import (
	"testing"

	"github.com/stratumn/go-indigonode/core/app/bootstrap/pb"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// NewAddRequest creates a new request to add the given peer.
func NewAddRequest(t *testing.T, peerID peer.ID) *proposal.Request {
	nodeID := &pb.NodeIdentity{
		PeerId:        []byte(peerID),
		PeerAddr:      test.GeneratePeerMultiaddr(t, peerID).Bytes(),
		IdentityProof: []byte("much proof very wow"),
	}

	r, err := proposal.NewAddRequest(nodeID)
	require.NoError(t, err, "proposal.NewAddRequest()")

	return r
}

// NewRemoveRequest creates a new request to remove the given peer.
func NewRemoveRequest(t *testing.T, peerID peer.ID) *proposal.Request {
	nodeID := &pb.NodeIdentity{
		PeerId:        []byte(peerID),
		IdentityProof: []byte("much proof very wow"),
	}

	r, err := proposal.NewRemoveRequest(nodeID)
	require.NoError(t, err, "proposal.NewRemoveRequest()")

	return r
}
