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

package proposaltest

import (
	"testing"

	"github.com/stratumn/go-indigonode/core/app/bootstrap/pb"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
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
