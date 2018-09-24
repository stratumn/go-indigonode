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

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/cstesting"
	pb "github.com/stratumn/go-indigonode/app/indigo/pb/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/audit"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/constants"
	protectormocks "github.com/stratumn/go-indigonode/core/protector/mocks"
	protectorpb "github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/core/streamutil/mockstream"
	"github.com/stratumn/go-indigonode/core/streamutil/streamtest"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stratumn/go-indigonode/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	"gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore/pstoremem"
)

func TestPrivate_NodeID(t *testing.T) {
	t.Run("panics-if-not-joined", func(t *testing.T) {
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		assert.Panics(t, func() { networkMgr.NodeID() })
	})

	t.Run("returns-host-ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		peerID := test.GeneratePeerID(t)

		h := mocks.NewMockHost(ctrl)
		h.EXPECT().ID().Return(peerID)

		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		networkMgr.Join(context.Background(), "", h)

		id := networkMgr.NodeID()
		assert.Equal(t, peerID, id)
	})
}

func TestPrivate_JoinLeave(t *testing.T) {
	t.Run("join-then-leave", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mocks.NewMockHost(ctrl)
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		err := networkMgr.Join(context.Background(), "", h)
		require.NoError(t, err, "networkMgr.Join()")

		err = networkMgr.Leave(context.Background(), "")
		require.NoError(t, err, "networkMgr.Leave()")
	})

	t.Run("join-multiple-times", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mocks.NewMockHost(ctrl)
		networkMgr := store.NewPrivateNetworkManager(nil, nil)

		for i := 0; i < 3; i++ {
			err := networkMgr.Join(context.Background(), "", h)
			require.NoError(t, err, "networkMgr.Join()")
		}
	})

	t.Run("leave-multiple-times", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mocks.NewMockHost(ctrl)
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		err := networkMgr.Join(context.Background(), "", h)
		require.NoError(t, err, "networkMgr.Join()")

		for i := 0; i < 3; i++ {
			err = networkMgr.Leave(context.Background(), "")
			require.NoError(t, err, "networkMgr.Leave()")
		}
	})

	t.Run("leave-without-joining", func(t *testing.T) {
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		err := networkMgr.Leave(context.Background(), "")
		require.NoError(t, err, "networkMgr.Leave()")
	})
}

func TestPrivate_Publish(t *testing.T) {
	t.Run("rejected-during-bootstrap", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		networkCfg := protectormocks.NewMockNetworkConfig(ctrl)
		networkMgr := store.NewPrivateNetworkManager(nil, networkCfg)

		networkCfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_BOOTSTRAP)

		err := networkMgr.Publish(context.Background(), cstesting.RandomLink())
		assert.EqualError(t, err, store.ErrNetworkNotReady.Error())
	})

	t.Run("send-to-participants", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		hostKey := test.GeneratePrivateKey(t)
		hostID := test.GetPeerIDFromKey(t, hostKey)
		peer1 := test.GeneratePeerID(t)

		l := cstesting.RandomLink()
		constants.SetLinkNodeID(l, hostID)

		pstore := pstoremem.NewPeerstore()
		pstore.AddPrivKey(hostID, hostKey)

		h := mocks.NewMockHost(ctrl)
		h.EXPECT().ID().Return(hostID).AnyTimes()
		h.EXPECT().Peerstore().Return(pstore)

		networkCfg := protectormocks.NewMockNetworkConfig(ctrl)
		networkCfg.EXPECT().NetworkState(gomock.Any()).Return(protectorpb.NetworkState_PROTECTED)
		networkCfg.EXPECT().AllowedPeers(gomock.Any()).Return([]peer.ID{hostID, peer1})

		codec := mockstream.NewMockCodec(ctrl)
		codec.EXPECT().Encode(gomock.Any()).Do(func(n interface{}) error {
			encoded, ok := n.(*pb.Segment)
			require.True(t, ok)

			segment, err := encoded.ToSegment()
			require.NoError(t, err, "encoded.ToSegment()")

			networkEvidences := segment.Meta.FindEvidences(audit.PeerSignatureBackend)
			require.Len(t, networkEvidences, 1)

			valid := networkEvidences[0].Proof.Verify(segment.GetLinkHash()[:])
			require.True(t, valid)

			return nil
		})

		stream := mockstream.NewMockStream(ctrl)
		stream.EXPECT().Codec().Return(codec).AnyTimes()
		stream.EXPECT().Close()

		streamProvider := mockstream.NewMockProvider(ctrl)
		streamtest.ExpectStreamPeerAndProtocol(t, streamProvider, peer1, store.IndigoLinkProtocolID, stream, nil)

		networkMgr := store.NewPrivateNetworkManager(streamProvider, networkCfg)
		networkMgr.Join(context.Background(), "", h)

		err := networkMgr.Publish(context.Background(), l)
		assert.NoError(t, err, "networkMgr.Publish()")
	})
}

func TestPrivate_Listen(t *testing.T) {
	t.Run("set-protocol-handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mocks.NewMockHost(ctrl)
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		networkMgr.Join(ctx, "", h)

		h.EXPECT().SetStreamHandler(store.IndigoLinkProtocolID, gomock.Any())

		errChan := make(chan error)
		go func() {
			errChan <- networkMgr.Listen(ctx)
		}()

		h.EXPECT().RemoveStreamHandler(store.IndigoLinkProtocolID)
		cancel()

		err := <-errChan
		assert.EqualError(t, err, context.Canceled.Error())
	})

	t.Run("close-listeners", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := mocks.NewMockHost(ctrl)
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		networkMgr.Join(ctx, "", h)
		testChan := networkMgr.AddListener()

		h.EXPECT().SetStreamHandler(store.IndigoLinkProtocolID, gomock.Any()).AnyTimes()
		h.EXPECT().RemoveStreamHandler(store.IndigoLinkProtocolID).AnyTimes()

		errChan := make(chan error)
		go func() {
			errChan <- networkMgr.Listen(ctx)
		}()

		cancel()
		<-errChan

		_, ok := <-testChan
		assert.False(t, ok)
	})

	t.Run("ignore-invalid-message", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := mocks.NewMockStream(ctrl)
		codec := mockstream.NewMockCodec(ctrl)
		codec.EXPECT().Decode(gomock.Any()).Return(errors.New("invalid message"))

		networkMgr := store.NewPrivateNetworkManager(nil, nil).(*store.PrivateNetworkManager)
		linksChan := networkMgr.AddListener()

		err := networkMgr.HandleNewLink(context.Background(), test.NewSpan(), stream, codec)
		assert.EqualError(t, err, "invalid message")

		select {
		case <-linksChan:
			assert.Fail(t, "invalid message should be discarded")
		case <-time.After(15 * time.Millisecond):
			return
		}
	})

	t.Run("forward-segments-to-listeners", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		segment := cstesting.RandomSegment()
		encodedSegment, _ := pb.FromSegment(segment)

		stream := mocks.NewMockStream(ctrl)
		codec := mockstream.NewMockCodec(ctrl)
		streamtest.ExpectDecodeSegment(t, codec, encodedSegment)

		networkMgr := store.NewPrivateNetworkManager(nil, nil).(*store.PrivateNetworkManager)
		linksChan := networkMgr.AddListener()

		err := networkMgr.HandleNewLink(context.Background(), test.NewSpan(), stream, codec)
		require.NoError(t, err)

		seg := <-linksChan
		assert.Equal(t, segment, seg)
	})
}

func TestPrivate_AddRemoveListeners(t *testing.T) {
	t.Run("remove-closes-channel", func(t *testing.T) {
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		testChan := networkMgr.AddListener()
		networkMgr.RemoveListener(testChan)

		_, ok := <-testChan
		assert.False(t, ok, "<-testChan")
	})

	t.Run("remove-unknown-channel", func(t *testing.T) {
		networkMgr := store.NewPrivateNetworkManager(nil, nil)
		privateChan := make(chan *cs.Segment)
		networkMgr.RemoveListener(privateChan)

		networkMgr.AddListener()
		networkMgr.RemoveListener(privateChan)
	})
}
