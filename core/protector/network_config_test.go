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

package protector_test

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protector/mocks"
	"github.com/stratumn/alice/core/protector/pb"
	"github.com/stratumn/alice/test"
	libp2pmocks "github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
)

func generateValidPeerAddrs(t *testing.T, peerID peer.ID) *pb.PeerAddrs {
	if peerID == "" {
		return &pb.PeerAddrs{
			Addresses: []string{test.GenerateMultiaddr(t).String()},
		}
	}

	return &pb.PeerAddrs{
		Addresses: []string{test.GeneratePeerMultiaddr(t, peerID).String()},
	}
}

func TestInMemoryConfig(t *testing.T) {
	ctx := context.Background()

	signerKey := test.GeneratePrivateKey(t)

	peer1 := test.GeneratePeerID(t)
	peerAddr1 := test.GeneratePeerMultiaddr(t, peer1)
	peer2 := test.GeneratePeerID(t)
	peerAddr2 := test.GeneratePeerMultiaddr(t, peer2)

	t.Run("New()", func(t *testing.T) {
		t.Run("rejects-invalid-config", func(t *testing.T) {
			_, err := protector.NewInMemoryConfig(
				ctx,
				&pb.NetworkConfig{NetworkState: 42},
			)
			assert.EqualError(t, err, pb.ErrInvalidNetworkState.Error())
		})

		t.Run("creates-valid-config", func(t *testing.T) {
			confData := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)
			confData.Participants[peer1.Pretty()] = &pb.PeerAddrs{
				Addresses: []string{peerAddr1.String()},
			}

			networkConfig, err := protector.NewInMemoryConfig(ctx, confData)
			require.NoError(t, err, "protector.NewInMemoryConfig()")

			assert.Equal(t, networkConfig.NetworkState(ctx), pb.NetworkState_BOOTSTRAP)
			assert.True(t, networkConfig.IsAllowed(ctx, peer1))
			assert.False(t, networkConfig.IsAllowed(ctx, peer2))
		})
	})

	t.Run("AddPeer()", func(t *testing.T) {
		networkConfig, _ := protector.NewInMemoryConfig(
			ctx,
			pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
		)

		err := networkConfig.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peerAddr1})
		require.NoError(t, err, "networkConfig.AddPeer()")

		assert.True(t, networkConfig.IsAllowed(ctx, peer1))

		err = networkConfig.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peerAddr1})
		require.NoError(t, err, "networkConfig.AddPeer()")
	})

	t.Run("RemovePeer()", func(t *testing.T) {
		networkConfig, _ := protector.NewInMemoryConfig(
			ctx,
			pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
		)

		err := networkConfig.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peerAddr1})
		require.NoError(t, err, "networkConfig.AddPeer()")

		err = networkConfig.RemovePeer(ctx, peer1)
		require.NoError(t, err, "networkConfig.RemovePeer()")

		assert.False(t, networkConfig.IsAllowed(ctx, peer1))
	})

	t.Run("AllowedPeers()", func(t *testing.T) {
		networkConfig, _ := protector.NewInMemoryConfig(
			ctx,
			pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
		)

		assert.Nil(t, networkConfig.AllowedPeers(ctx))

		networkConfig.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peerAddr1})
		networkConfig.AddPeer(ctx, peer2, []multiaddr.Multiaddr{peerAddr2})

		assert.Len(t, networkConfig.AllowedPeers(ctx), 2)
		assert.True(t, networkConfig.IsAllowed(ctx, peer1))
		assert.True(t, networkConfig.IsAllowed(ctx, peer2))
	})

	t.Run("SetNetworkState()", func(t *testing.T) {
		t.Run("rejects-invalid-state", func(t *testing.T) {
			networkConfig, _ := protector.NewInMemoryConfig(
				ctx,
				pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
			)

			err := networkConfig.SetNetworkState(ctx, 42)
			assert.EqualError(t, err, pb.ErrInvalidNetworkState.Error())
		})

		t.Run("sets-state", func(t *testing.T) {
			networkConfig, _ := protector.NewInMemoryConfig(
				ctx,
				pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
			)

			assert.Equal(t, pb.NetworkState_BOOTSTRAP, networkConfig.NetworkState(ctx))

			err := networkConfig.SetNetworkState(ctx, pb.NetworkState_PROTECTED)
			require.NoError(t, err, "networkConfig.SetNetworkState()")
			assert.Equal(t, pb.NetworkState_PROTECTED, networkConfig.NetworkState(ctx))
		})
	})

	t.Run("Sign()", func(t *testing.T) {
		networkConfig, _ := protector.NewInMemoryConfig(
			ctx,
			pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
		)

		err := networkConfig.Sign(ctx, signerKey)
		assert.NoError(t, err, "networkConfig.Sign()")
	})

	t.Run("Copy()", func(t *testing.T) {
		networkConfig, _ := protector.NewInMemoryConfig(
			ctx,
			pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
		)

		networkConfig.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peerAddr1})
		networkConfig.SetNetworkState(ctx, pb.NetworkState_PROTECTED)

		copy := networkConfig.Copy(ctx)
		assert.Equal(t, pb.NetworkState_PROTECTED, copy.NetworkState)
		assert.Len(t, copy.Participants, 1)
		assert.Contains(t, copy.Participants, peer1.Pretty())
		assert.Len(t, copy.Participants[peer1.Pretty()].Addresses, 1)
		assert.Equal(t, peerAddr1.String(), copy.Participants[peer1.Pretty()].Addresses[0])
	})

	t.Run("Reset()", func(t *testing.T) {
		conf, err := protector.NewInMemoryConfig(ctx, pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP))
		require.NoError(t, err, "protector.NewInMemoryConfig()")

		t.Run("rejects-invalid-network-state", func(t *testing.T) {
			assert.EqualError(t,
				conf.Reset(ctx, &pb.NetworkConfig{NetworkState: 42}),
				pb.ErrInvalidNetworkState.Error(),
			)
		})

		t.Run("rejects-invalid-peer-id", func(t *testing.T) {
			assert.EqualError(t,
				conf.Reset(ctx, &pb.NetworkConfig{
					Participants: map[string]*pb.PeerAddrs{
						"b4tm4n": generateValidPeerAddrs(t, ""),
					},
				}),
				pb.ErrInvalidPeerID.Error(),
			)
		})

		t.Run("rejects-invalid-addr", func(t *testing.T) {
			peerID := test.GeneratePeerID(t)

			assert.EqualError(t,
				conf.Reset(ctx, &pb.NetworkConfig{
					Participants: map[string]*pb.PeerAddrs{
						peerID.Pretty(): &pb.PeerAddrs{Addresses: []string{"not/a/multi/addr"}},
					},
				}),
				pb.ErrInvalidPeerAddr.Error(),
			)
		})

		t.Run("accepts-valid-config", func(t *testing.T) {
			peer1 := test.GeneratePeerID(t)
			peer2 := test.GeneratePeerID(t)

			err := conf.Reset(ctx, &pb.NetworkConfig{
				NetworkState: pb.NetworkState_PROTECTED,
				Participants: map[string]*pb.PeerAddrs{
					peer1.Pretty(): generateValidPeerAddrs(t, peer1),
					peer2.Pretty(): generateValidPeerAddrs(t, peer2),
				},
			})
			require.NoError(t, err, "conf.Reset()")

			assert.Equal(t, pb.NetworkState_PROTECTED, conf.NetworkState(ctx))
			assert.Len(t, conf.AllowedPeers(ctx), 2)
			assert.True(t, conf.IsAllowed(ctx, peer1))
			assert.True(t, conf.IsAllowed(ctx, peer2))
		})
	})
}

func TestConfigSigner(t *testing.T) {
	ctx := context.Background()

	signerKey := test.GeneratePrivateKey(t)
	signerID := test.GetPeerIDFromKey(t, signerKey)

	peer1 := test.GeneratePeerID(t)
	peer1Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer1)}
	peer2 := test.GeneratePeerID(t)
	peer2Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer2)}

	inMemoryConfig, _ := protector.NewInMemoryConfig(ctx, pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP))
	networkConfig := protector.WrapWithSignature(inMemoryConfig, signerKey)

	t.Run("AddPeer()", func(t *testing.T) {
		oldSig := networkConfig.Copy(ctx).Signature

		err := networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		copy := networkConfig.Copy(ctx)
		assert.NotEqual(t, oldSig, copy.Signature)
		assert.True(t, copy.ValidateSignature(ctx, signerID))
	})

	t.Run("RemovePeer()", func(t *testing.T) {
		err := networkConfig.AddPeer(ctx, peer2, peer2Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		oldSig := networkConfig.Copy(ctx).Signature

		err = networkConfig.RemovePeer(ctx, peer2)
		require.NoError(t, err, "networkConfig.RemovePeer()")

		copy := networkConfig.Copy(ctx)
		assert.NotEqual(t, oldSig.Signature, copy.Signature.Signature)
		assert.True(t, copy.ValidateSignature(ctx, signerID))
	})

	t.Run("SetNetworkState()", func(t *testing.T) {
		t.Run("does-not-sign-if-set-fails", func(t *testing.T) {
			oldSig := networkConfig.Copy(ctx).Signature

			err := networkConfig.SetNetworkState(ctx, 42)
			require.EqualError(t, err, pb.ErrInvalidNetworkState.Error())

			newSig := networkConfig.Copy(ctx).Signature
			assert.Equal(t, oldSig.Signature, newSig.Signature)
		})

		t.Run("updates-signature", func(t *testing.T) {
			oldSig := networkConfig.Copy(ctx).Signature

			err := networkConfig.SetNetworkState(ctx, pb.NetworkState_PROTECTED)
			require.NoError(t, err, "networkConfig.SetNetworkState()")

			copy := networkConfig.Copy(ctx)
			assert.NotEqual(t, oldSig.Signature, copy.Signature.Signature)
			assert.True(t, copy.ValidateSignature(ctx, signerID))
		})
	})

	t.Run("Reset()", func(t *testing.T) {
		err := networkConfig.Reset(ctx, pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP))
		require.NoError(t, err, "networkConfig.Reset()")

		copy := networkConfig.Copy(ctx)
		assert.True(t, copy.ValidateSignature(ctx, signerID))
	})
}

func TestConfigSaver(t *testing.T) {
	ctx := context.Background()

	dir, _ := ioutil.TempDir("", "alice")
	configPath := path.Join(dir, "config.json")

	signerKey := test.GeneratePrivateKey(t)
	signerID := test.GetPeerIDFromKey(t, signerKey)

	peer1 := test.GeneratePeerID(t)
	peer1Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer1)}
	peer2 := test.GeneratePeerID(t)
	peer2Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer2)}

	inMemoryConfig, _ := protector.NewInMemoryConfig(ctx, pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP))
	networkConfig := protector.WrapWithSaver(
		protector.WrapWithSignature(inMemoryConfig, signerKey),
		configPath,
	)

	saved := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)

	t.Run("AddPeer()/RemovePeer()", func(t *testing.T) {
		err := networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		err = networkConfig.AddPeer(ctx, peer2, peer2Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		err = saved.LoadFromFile(ctx, configPath, signerID)
		require.NoError(t, err, "saved.LoadFromFile()")

		assert.Equal(t, pb.NetworkState_BOOTSTRAP, saved.NetworkState)
		assert.Len(t, saved.Participants, 2)
		assert.Contains(t, saved.Participants, peer1.Pretty())
		assert.Contains(t, saved.Participants, peer2.Pretty())

		err = networkConfig.RemovePeer(ctx, peer1)
		require.NoError(t, err, "networkConfig.RemovePeer()")

		err = saved.LoadFromFile(ctx, configPath, signerID)
		require.NoError(t, err, "saved.LoadFromFile()")

		assert.Equal(t, pb.NetworkState_BOOTSTRAP, saved.NetworkState)
		assert.Len(t, saved.Participants, 1)
		assert.Contains(t, saved.Participants, peer2.Pretty())
	})

	t.Run("SetNetworkState()", func(t *testing.T) {
		t.Run("does-not-save-if-set-fails", func(t *testing.T) {
			networkState := networkConfig.NetworkState(ctx)

			err := networkConfig.SetNetworkState(ctx, 42)
			require.EqualError(t, err, pb.ErrInvalidNetworkState.Error())

			err = saved.LoadFromFile(ctx, configPath, signerID)
			require.NoError(t, err, "saved.LoadFromFile()")

			assert.Equal(t, networkState, saved.NetworkState)
		})

		t.Run("saves-valid-change", func(t *testing.T) {
			err := networkConfig.SetNetworkState(ctx, pb.NetworkState_PROTECTED)
			require.NoError(t, err, "networkConfig.SetNetworkState()")

			err = saved.LoadFromFile(ctx, configPath, signerID)
			require.NoError(t, err, "saved.LoadFromFile()")

			assert.Equal(t, pb.NetworkState_PROTECTED, saved.NetworkState)
		})
	})

	t.Run("Reset()", func(t *testing.T) {
		err := networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		err = networkConfig.Reset(ctx, pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP))
		require.NoError(t, err, "networkConfig.Reset()")

		err = saved.LoadFromFile(ctx, configPath, signerID)
		require.NoError(t, err, "saved.LoadFromFile()")

		assert.Equal(t, pb.NetworkState_BOOTSTRAP, saved.NetworkState)
		assert.Nil(t, saved.Participants)
	})
}

func TestConfigProtectUpdater(t *testing.T) {
	ctx := context.Background()

	peer1 := test.GeneratePeerID(t)
	peer1Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer1)}
	peer2 := test.GeneratePeerID(t)
	peer2Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer2)}
	peer3 := test.GeneratePeerID(t)

	inMemoryConfig, _ := protector.NewInMemoryConfig(
		ctx,
		pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP),
	)

	t.Run("WrapWithProtectUpdater", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		p := mocks.NewMockProtector(ctrl)
		testChan := make(chan struct{})
		p.EXPECT().ListenForUpdates(gomock.Any()).Times(1).Do(func(<-chan protector.NetworkUpdate) {
			testChan <- struct{}{}
		})

		protector.WrapWithProtectUpdater(inMemoryConfig, p, nil)

		select {
		case <-time.After(10 * time.Millisecond):
			assert.Fail(t, "ListenForUpdates not called in time")
		case <-testChan:
		}
	})

	t.Run("AddPeer()", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		peerStore := libp2pmocks.NewMockPeerstore(ctrl)
		peerStore.EXPECT().AddAddrs(peer1, peer1Addrs, gomock.Any()).Times(1)
		p := protector.NewPrivateNetwork(peerStore)

		networkConfig := protector.WrapWithProtectUpdater(inMemoryConfig, p, peerStore)
		err := networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		waitUntilAllowed(t, p, peer1, 1)
	})

	t.Run("RemovePeer()", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		peerStore := libp2pmocks.NewMockPeerstore(ctrl)
		peerStore.EXPECT().AddAddrs(peer1, peer1Addrs, gomock.Any()).Times(1)
		peerStore.EXPECT().AddAddrs(peer2, peer2Addrs, gomock.Any()).Times(1)

		p := protector.NewPrivateNetwork(peerStore)

		networkConfig := protector.WrapWithProtectUpdater(inMemoryConfig, p, peerStore)

		err := networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")
		err = networkConfig.AddPeer(ctx, peer2, peer2Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		waitUntilAllowed(t, p, peer2, 2)

		err = networkConfig.RemovePeer(ctx, peer1)
		require.NoError(t, err, "networkConfig.RemovePeer()")

		waitUntilAllowed(t, p, peer2, 1)
	})

	t.Run("SetNetworkState()", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		p := mocks.NewMockStateAwareProtector(ctrl)
		p.EXPECT().ListenForUpdates(gomock.Any()).AnyTimes()

		networkConfig := protector.WrapWithProtectUpdater(inMemoryConfig, p, nil)

		p.EXPECT().SetNetworkState(gomock.Any(), pb.NetworkState_PROTECTED).Times(1)
		err := networkConfig.SetNetworkState(ctx, pb.NetworkState_PROTECTED)
		require.NoError(t, err, "networkConfig.SetNetworkState()")
	})

	t.Run("Reset()", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		peerStore := peerstore.NewPeerstore()
		p := protector.NewPrivateNetworkWithBootstrap(peerStore)

		networkConfig := protector.WrapWithProtectUpdater(inMemoryConfig, p, peerStore)

		err := networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")
		err = networkConfig.AddPeer(ctx, peer2, peer2Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		waitUntilAllowed(t, p, peer2, 2)

		err = networkConfig.Reset(ctx, &pb.NetworkConfig{
			NetworkState: pb.NetworkState_PROTECTED,
			Participants: map[string]*pb.PeerAddrs{
				peer1.Pretty(): generateValidPeerAddrs(t, peer1),
				peer3.Pretty(): generateValidPeerAddrs(t, peer3),
			},
		})
		require.NoError(t, err, "networkConfig.Reset()")

		waitUntilAllowed(t, p, peer3, 2)

		assert.ElementsMatch(t, []peer.ID{peer1, peer3}, p.AllowedPeers(ctx))

		rejectedConn := libp2pmocks.NewMockTransportConn(ctrl)
		rejectedConn.EXPECT().LocalMultiaddr().Times(1).Return(peer1Addrs[0])
		rejectedConn.EXPECT().RemoteMultiaddr().Times(1).Return(peer2Addrs[0])
		rejectedConn.EXPECT().Close().AnyTimes()

		_, err = p.Protect(rejectedConn)
		assert.EqualError(t, err, protector.ErrConnectionRefused.Error())
	})
}

func TestLoadOrInitNetworkConfig(t *testing.T) {
	ctx := context.Background()

	dir, _ := ioutil.TempDir("", "alice")
	configPath := path.Join(dir, "config.json")

	signerKey := test.GeneratePrivateKey(t)

	peer1 := test.GeneratePeerID(t)
	peer1Addrs := []multiaddr.Multiaddr{test.GeneratePeerMultiaddr(t, peer1)}

	t.Run("creates-new-config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		peerStore := libp2pmocks.NewMockPeerstore(ctrl)
		p := protector.NewPrivateNetwork(peerStore)

		networkConfig, err := protector.LoadOrInitNetworkConfig(ctx, configPath, signerKey, p, peerStore)
		require.NoError(t, err, "protector.LoadOrInitNetworkConfig()")

		peerStore.EXPECT().AddAddrs(peer1, peer1Addrs, gomock.Any()).Times(1)
		err = networkConfig.AddPeer(ctx, peer1, peer1Addrs)
		require.NoError(t, err, "networkConfig.AddPeer()")

		err = networkConfig.SetNetworkState(ctx, pb.NetworkState_PROTECTED)
		require.NoError(t, err, "networkConfig.SetNetworkState()")

		waitUntilAllowed(t, p, peer1, 1)
		assert.True(t, networkConfig.IsAllowed(ctx, peer1))
	})

	t.Run("loads-existing-config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, err := os.Stat(configPath)
		require.NoError(t, err, "os.Stat(configPath)")

		peerStore := libp2pmocks.NewMockPeerstore(ctrl)
		p := protector.NewPrivateNetwork(peerStore)

		peerStore.EXPECT().AddAddrs(peer1, peer1Addrs, gomock.Any()).Times(1)

		networkConfig, err := protector.LoadOrInitNetworkConfig(ctx, configPath, signerKey, p, peerStore)
		require.NoError(t, err, "protector.LoadOrInitNetworkConfig()")

		assert.Equal(t, pb.NetworkState_PROTECTED, networkConfig.NetworkState(ctx))
		assert.Equal(t, []peer.ID{peer1}, networkConfig.AllowedPeers(ctx))
		assert.True(t, networkConfig.IsAllowed(ctx, peer1))
	})

	t.Run("rejects-invalid-existing-config", func(t *testing.T) {
		unknownKey := test.GeneratePrivateKey(t)

		p := protector.NewPrivateNetwork(nil)

		_, err := protector.LoadOrInitNetworkConfig(ctx, configPath, unknownKey, p, nil)
		require.EqualError(t, err, pb.ErrInvalidSignature.Error())
	})
}
