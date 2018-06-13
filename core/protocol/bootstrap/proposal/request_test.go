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

package proposal_test

import (
	"testing"

	"github.com/stratumn/alice/core/protocol/bootstrap/proposal"
	pb "github.com/stratumn/alice/pb/bootstrap"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType_String(t *testing.T) {
	t.Run("unknown-type", func(t *testing.T) {
		var unknown proposal.Type = 42
		assert.Equal(t, "Unknown", unknown.String())
	})

	t.Run("add-node", func(t *testing.T) {
		assert.Equal(t, "Add", proposal.AddNode.String())
	})
}

func TestRequest_New(t *testing.T) {
	peerID := test.GeneratePeerID(t)
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)

	t.Run("add-invalid-peer-id", func(t *testing.T) {
		_, err := proposal.NewAddRequest(&pb.NodeIdentity{PeerId: []byte("b4tm4n")})
		assert.EqualError(t, err, proposal.ErrInvalidPeerID.Error())
	})

	t.Run("add-invalid-peer-addr", func(t *testing.T) {
		_, err := proposal.NewAddRequest(&pb.NodeIdentity{
			PeerId:   []byte(peerID),
			PeerAddr: []byte("not/a/multiaddr"),
		})
		assert.EqualError(t, err, proposal.ErrInvalidPeerAddr.Error())
	})

	t.Run("add-request", func(t *testing.T) {
		req, err := proposal.NewAddRequest(&pb.NodeIdentity{
			PeerId:        []byte(peerID),
			PeerAddr:      peerAddr.Bytes(),
			IdentityProof: []byte("that guy is b4tm4n"),
		})
		require.NoError(t, err)
		require.NotNil(t, req)

		assert.Equal(t, proposal.AddNode, req.Type)
		assert.Equal(t, peerID, req.PeerID)
		assert.Equal(t, peerAddr, req.PeerAddr)
		assert.Equal(t, []byte("that guy is b4tm4n"), req.Info)
	})

	t.Run("remove-invalid-peer-id", func(t *testing.T) {
		_, err := proposal.NewRemoveRequest(&pb.NodeIdentity{PeerId: []byte("b4tm4n")})
		assert.EqualError(t, err, proposal.ErrInvalidPeerID.Error())
	})

	t.Run("remove-request", func(t *testing.T) {
		req, err := proposal.NewRemoveRequest(&pb.NodeIdentity{
			PeerId: []byte(peerID),
		})
		require.NoError(t, err)
		require.NotNil(t, req)

		assert.Equal(t, proposal.RemoveNode, req.Type)
		assert.Equal(t, peerID, req.PeerID)
		assert.NotNil(t, req.Challenge)
	})
}
