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

package streamtest

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/app/bootstrap/pb"
	"github.com/stratumn/alice/core/app/bootstrap/protocol/proposal"
	"github.com/stratumn/alice/core/streamutil/mockstream"
	protectorpb "github.com/stratumn/alice/pb/protector"
	"github.com/stretchr/testify/require"

	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// ExpectDecodeNodeID configures a mock codec to decode the given nodeID.
func ExpectDecodeNodeID(t *testing.T, codec *mockstream.MockCodec, nodeID *pb.NodeIdentity) {
	codec.EXPECT().Decode(gomock.Any()).Do(func(n interface{}) error {
		nid, ok := n.(*pb.NodeIdentity)
		require.True(t, ok, "n.(*pb.NodeIdentity)")

		nid.PeerId = nodeID.PeerId
		nid.PeerAddr = nodeID.PeerAddr
		nid.IdentityProof = nodeID.IdentityProof

		return nil
	})
}

// ExpectDecodeVote configures a mock codec to decode the given vote.
func ExpectDecodeVote(t *testing.T, codec *mockstream.MockCodec, vote *pb.Vote) {
	codec.EXPECT().Decode(gomock.Any()).Do(func(n interface{}) error {
		v, ok := n.(*pb.Vote)
		require.True(t, ok, "n.(*pb.Vote)")

		v.Challenge = vote.Challenge
		v.PeerId = vote.PeerId
		v.Signature = vote.Signature
		v.UpdateType = vote.UpdateType

		return nil
	})
}

// ExpectDecodeConfig configures a mock codec to decode the given network config.
func ExpectDecodeConfig(t *testing.T, codec *mockstream.MockCodec, cfg *protectorpb.NetworkConfig) {
	codec.EXPECT().Decode(gomock.Any()).Do(func(n interface{}) error {
		c, ok := n.(*protectorpb.NetworkConfig)
		require.True(t, ok, "n.(*protectorpb.NetworkConfig)")

		*c = *cfg
		return nil
	})
}

// ExpectDecodeUpdateProp configures a mock codec to decode the given network proposal.
func ExpectDecodeUpdateProp(t *testing.T, codec *mockstream.MockCodec, prop *pb.UpdateProposal) {
	codec.EXPECT().Decode(gomock.Any()).Do(func(n interface{}) error {
		p, ok := n.(*pb.UpdateProposal)
		require.True(t, ok, "n.(*pb.UpdateProposal)")

		*p = *prop
		return nil
	})
}

// ExpectEncodeAck configures a mock codec to encode an Ack
// with the given error.
func ExpectEncodeAck(t *testing.T, codec *mockstream.MockCodec, err error) {
	if err != nil {
		codec.EXPECT().Encode(&pb.Ack{Error: err.Error()})
	} else {
		codec.EXPECT().Encode(&pb.Ack{})
	}
}

// ExpectEncodeAllowed configures a mock codec to verify that
// the encoded network configuration contains a specific peer.
func ExpectEncodeAllowed(t *testing.T, codec *mockstream.MockCodec, peerID peer.ID) {
	codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
		cfg, ok := n.(*protectorpb.NetworkConfig)
		require.True(t, ok, "n.(*protectorpb.NetworkConfig)")

		p, ok := cfg.Participants[peerID.Pretty()]
		require.True(t, ok)
		require.True(t, len(p.Addresses) > 0)

		return nil
	})
}

// ExpectEncodeNetworkState configures a mock codec to verify that
// the encoded network configuration has the given network state.
func ExpectEncodeNetworkState(t *testing.T, codec *mockstream.MockCodec, state protectorpb.NetworkState) {
	codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
		cfg, ok := n.(*protectorpb.NetworkConfig)
		require.True(t, ok, "n.(*protectorpb.NetworkConfig)")
		require.Equal(t, state, cfg.NetworkState)

		return nil
	})
}

// ExpectEncodeVote configures a mock codec to verify that
// the encoded vote is a valid vote for the given request.
func ExpectEncodeVote(t *testing.T, codec *mockstream.MockCodec, r *proposal.Request) {
	codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
		vote, ok := n.(*pb.Vote)
		require.True(t, ok, "n.(*pb.Vote)")

		v := &proposal.Vote{}
		err := v.FromProtoVote(vote)
		require.NoError(t, err)
		require.NoError(t, v.Verify(r))

		return nil
	})
}
