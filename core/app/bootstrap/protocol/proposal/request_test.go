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
	"encoding/json"
	"testing"
	"time"

	"github.com/stratumn/go-indigonode/core/app/bootstrap/pb"
	"github.com/stratumn/go-indigonode/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/go-indigonode/test"
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

func TestRequest_UpdateProposal(t *testing.T) {
	t.Run("add-node", func(t *testing.T) {
		req := &proposal.Request{
			Type:      proposal.AddNode,
			PeerID:    test.GeneratePeerID(t),
			PeerAddr:  test.GenerateMultiaddr(t),
			Challenge: []byte("such challenge"),
			Info:      []byte("very info"),
			Expires:   time.Now().UTC().Add(10 * time.Minute),
		}

		prop := req.ToUpdateProposal()
		require.NotNil(t, prop)
		require.NotNil(t, prop.NodeDetails)

		assert.Equal(t, pb.UpdateType_AddNode, prop.UpdateType)
		assert.Equal(t, req.Challenge, prop.Challenge)
		assert.Equal(t, []byte(req.PeerID), prop.NodeDetails.PeerId)
		assert.Equal(t, req.PeerAddr.Bytes(), prop.NodeDetails.PeerAddr)
		assert.Equal(t, req.Info, prop.NodeDetails.IdentityProof)

		req2 := &proposal.Request{}
		err := req2.FromUpdateProposal(prop)
		require.NoError(t, err)

		assert.Equal(t, req.Type, req2.Type)
		assert.Equal(t, req.PeerID, req2.PeerID)
		assert.Equal(t, req.PeerAddr, req2.PeerAddr)
		assert.Equal(t, req.Info, req2.Info)
		assert.Equal(t, req.Challenge, req2.Challenge)
	})

	t.Run("remove-node", func(t *testing.T) {
		req := &proposal.Request{
			Type:      proposal.RemoveNode,
			PeerID:    test.GeneratePeerID(t),
			Challenge: []byte("such challenge"),
			Expires:   time.Now().UTC().Add(10 * time.Minute),
		}

		prop := req.ToUpdateProposal()
		require.NotNil(t, prop)
		require.NotNil(t, prop.NodeDetails)

		assert.Equal(t, pb.UpdateType_RemoveNode, prop.UpdateType)
		assert.Equal(t, req.Challenge, prop.Challenge)
		assert.Equal(t, []byte(req.PeerID), prop.NodeDetails.PeerId)

		req2 := &proposal.Request{}
		err := req2.FromUpdateProposal(prop)
		require.NoError(t, err)

		assert.Equal(t, req.Type, req2.Type)
		assert.Equal(t, req.PeerID, req2.PeerID)
		assert.Equal(t, req.Challenge, req2.Challenge)
	})
}

func TestRequest_MarshalJSON(t *testing.T) {
	req := &proposal.Request{
		Type:      proposal.AddNode,
		PeerID:    test.GeneratePeerID(t),
		PeerAddr:  test.GenerateMultiaddr(t),
		Challenge: []byte("such challenge"),
		Info:      []byte("very info"),
		Expires:   time.Now().UTC().Add(10 * time.Minute),
	}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	var deserialized proposal.Request
	err = json.Unmarshal(b, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, req.Type, deserialized.Type)
	assert.Equal(t, req.PeerID, deserialized.PeerID)
	assert.Equal(t, req.PeerAddr, deserialized.PeerAddr)
	assert.Equal(t, req.Info, deserialized.Info)
	assert.Equal(t, req.Challenge, deserialized.Challenge)
	assert.Equal(t, req.Expires, deserialized.Expires)
}
