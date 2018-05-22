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
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protector/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var (
	testKey crypto.PrivKey
)

func init() {
	testKey, _, _ = crypto.GenerateEd25519Key(rand.Reader)
}

func TestLocalConfig_InitConfig_Success(t *testing.T) {
	ctx := context.Background()

	t.Run("creates-config-folder", func(t *testing.T) {
		configDir, _ := ioutil.TempDir("", "alice")
		require.NoError(t, os.Remove(configDir))

		conf, err := protector.InitLocalConfig(
			ctx,
			filepath.Join(configDir, "config.json"),
			testKey,
			protector.NewPrivateNetwork(nil),
			nil,
		)
		require.NoError(t, err)
		assert.NotNil(t, conf)

		_, err = os.Stat(configDir)
		assert.NoError(t, err)
	})

	t.Run("configures-protector-listener", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		p := mocks.NewMockProtector(ctrl)
		testChan := make(chan struct{})
		p.EXPECT().ListenForUpdates(gomock.Any()).Times(1).Do(func(<-chan protector.NetworkUpdate) {
			testChan <- struct{}{}
		})

		configDir, _ := ioutil.TempDir("", "alice")
		_, err := protector.InitLocalConfig(
			ctx,
			filepath.Join(configDir, "config.json"),
			testKey,
			p,
			nil,
		)
		require.NoError(t, err)

		select {
		case <-time.After(5 * time.Millisecond):
			assert.Fail(t, "protector.ListenForUpdates()")
		case <-testChan:
		}
	})

	t.Run("loads-existing-config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		configDir, _ := ioutil.TempDir("", "alice")
		configPath := filepath.Join(configDir, "config.json")

		peerID := generatePeerID()
		peerAddr1 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8903")
		peerAddr2 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8913")

		configData := protector.ConfigData{
			PeersAddrs: map[string][]string{
				peerID.Pretty(): []string{peerAddr1.String(), peerAddr2.String()},
			},
		}

		require.NoError(t, configData.Flush(ctx, configPath, testKey))

		peerStore := mocks.NewMockPeerstore(ctrl)
		peerStore.EXPECT().AddAddr(peerID, peerAddr1, gomock.Any()).Times(1)
		peerStore.EXPECT().AddAddr(peerID, peerAddr2, gomock.Any()).Times(1)

		p := protector.NewPrivateNetwork(nil)
		conf, err := protector.InitLocalConfig(ctx, configPath, testKey, p, peerStore)
		require.NoError(t, err)
		assert.NotNil(t, conf)

		waitUntilAllowed(t, p, peerID, 1)
		allowed := conf.AllowedPeers()
		assert.ElementsMatch(t, []peer.ID{peerID}, allowed)
	})
}

func TestLocalConfig_InitConfig_Error(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                 string
		createPreviousConfig func(*testing.T, string)
		expectedError        error
	}{{
		"invalid-config",
		func(t *testing.T, configPath string) {
			err := ioutil.WriteFile(
				configPath,
				[]byte("not a json config"),
				0644,
			)
			require.NoError(t, err)
		},
		protector.ErrInvalidConfig,
	}, {
		"invalid-config-signature",
		func(t *testing.T, configPath string) {
			otherKey, _, _ := crypto.GenerateEd25519Key(rand.Reader)
			configData := protector.ConfigData{
				PeersAddrs: map[string][]string{
					generatePeerID().Pretty(): []string{"/ip4/127.0.0.1/tcp/8903"},
				},
			}

			require.NoError(t, configData.Flush(ctx, configPath, otherKey))
		},
		protector.ErrInvalidSignature,
	}, {
		"invalid-peer-id",
		func(t *testing.T, configPath string) {
			configData := protector.ConfigData{
				PeersAddrs: map[string][]string{
					"not-a-peer-id": []string{"/ip4/127.0.0.1/tcp/8903"},
				},
			}

			require.NoError(t, configData.Flush(ctx, configPath, testKey))
		},
		protector.ErrInvalidConfig,
	}, {
		"invalid-peer-address",
		func(t *testing.T, configPath string) {
			configData := protector.ConfigData{
				PeersAddrs: map[string][]string{
					generatePeerID().Pretty(): []string{"/not/a/multiaddr"},
				},
			}

			require.NoError(t, configData.Flush(ctx, configPath, testKey))
		},
		protector.ErrInvalidConfig,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			configDir, _ := ioutil.TempDir("", "alice")
			configPath := filepath.Join(configDir, "config.json")

			tt.createPreviousConfig(t, configPath)
			_, err := protector.InitLocalConfig(
				ctx,
				configPath,
				testKey,
				protector.NewPrivateNetwork(nil),
				nil,
			)

			assert.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestLocalConfig_AddPeer(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configDir, _ := ioutil.TempDir("", "alice")
	configPath := filepath.Join(configDir, "config.json")

	peer1 := generatePeerID()
	peer1Addr1 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8903")
	peer1Addr2 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8904")
	peer2 := generatePeerID()
	peer2Addr1 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8913")

	peerStore := mocks.NewMockPeerstore(ctrl)
	peerStore.EXPECT().AddAddrs(peer1, []multiaddr.Multiaddr{peer1Addr1, peer1Addr2}, gomock.Any()).Times(1)
	peerStore.EXPECT().AddAddrs(peer2, []multiaddr.Multiaddr{peer2Addr1}, gomock.Any()).Times(1)

	p := protector.NewPrivateNetwork(peerStore)
	conf, err := protector.InitLocalConfig(ctx, configPath, testKey, p, peerStore)
	require.NoError(t, err)

	require.NoError(t, conf.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peer1Addr1, peer1Addr2}))
	require.NoError(t, conf.AddPeer(ctx, peer2, []multiaddr.Multiaddr{peer2Addr1}))

	waitUntilAllowed(t, p, peer2, 2)
	assert.ElementsMatch(t, []peer.ID{peer1, peer2}, conf.AllowedPeers())
}

func TestLocalConfig_RemovePeer(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configDir, _ := ioutil.TempDir("", "alice")
	configPath := filepath.Join(configDir, "config.json")

	peer1 := generatePeerID()
	peer2 := generatePeerID()

	peerStore := mocks.NewMockPeerstore(ctrl)
	peerStore.EXPECT().AddAddrs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	p := protector.NewPrivateNetwork(peerStore)
	conf, err := protector.InitLocalConfig(ctx, configPath, testKey, p, peerStore)
	require.NoError(t, err)
	require.NoError(t, conf.AddPeer(ctx, peer1, nil))
	require.NoError(t, conf.AddPeer(ctx, peer2, nil))

	waitUntilAllowed(t, p, peer2, 2)

	conf.RemovePeer(ctx, peer1)
	waitUntilAllowed(t, p, peer2, 1)

	assert.ElementsMatch(t, []peer.ID{peer2}, conf.AllowedPeers())
}

func TestLocalConfig_Flush(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	peerStore := mocks.NewMockPeerstore(ctrl)
	peerStore.EXPECT().AddAddrs(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	p := protector.NewPrivateNetwork(peerStore)

	peer1 := generatePeerID()
	peer1Addr1 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8903")
	peer2 := generatePeerID()
	peer2Addr1 := multiaddr.StringCast("/ip4/127.0.0.1/tcp/8913")

	configDir, _ := ioutil.TempDir("", "alice")
	configPath := filepath.Join(configDir, "config.json")

	conf, err := protector.InitLocalConfig(ctx, configPath, testKey, p, peerStore)
	require.NoError(t, err)
	require.NoError(t, conf.AddPeer(ctx, peer1, []multiaddr.Multiaddr{peer1Addr1}))
	require.NoError(t, conf.AddPeer(ctx, peer2, []multiaddr.Multiaddr{peer2Addr1}))

	configData := protector.NewConfigData()
	require.NoError(t, configData.Load(ctx, configPath, testKey))
	assert.Len(t, configData.PeersAddrs, 2)
	assert.ElementsMatch(t, []string{peer1Addr1.String()}, configData.PeersAddrs[peer1.Pretty()])
	assert.ElementsMatch(t, []string{peer2Addr1.String()}, configData.PeersAddrs[peer2.Pretty()])

	require.NoError(t, conf.RemovePeer(ctx, peer1))
	require.NoError(t, configData.Load(ctx, configPath, testKey))
	assert.Len(t, configData.PeersAddrs, 1)
	assert.ElementsMatch(t, []string{peer2Addr1.String()}, configData.PeersAddrs[peer2.Pretty()])
}
