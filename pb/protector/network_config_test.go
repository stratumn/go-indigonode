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
	"path"
	"testing"

	pb "github.com/stratumn/alice/pb/protector"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

func generatePeerAddrs(t *testing.T, peerID peer.ID) *pb.PeerAddrs {
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)
	return &pb.PeerAddrs{
		Addresses: []string{peerAddr.String()},
	}
}

func TestNetworkConfig_Signature(t *testing.T) {
	sk1 := test.GeneratePrivateKey(t)
	peer1 := test.GetPeerIDFromKey(t, sk1)

	sk2 := test.GeneratePrivateKey(t)
	peer2 := test.GetPeerIDFromKey(t, sk2)

	testCases := []struct {
		name          string
		networkConfig func(*testing.T) *pb.NetworkConfig
		signerID      peer.ID
		expected      bool
	}{{
		"nil-config",
		func(*testing.T) *pb.NetworkConfig { return nil },
		peer1,
		false,
	}, {
		"invalid-signer",
		func(t *testing.T) *pb.NetworkConfig {
			networkConfig := pb.NewNetworkConfig(pb.NetworkState_PROTECTED)
			networkConfig.Participants[peer1.Pretty()] = generatePeerAddrs(t, peer1)
			networkConfig.Participants[peer2.Pretty()] = generatePeerAddrs(t, peer2)

			require.NoError(t, networkConfig.Sign(context.Background(), sk1))
			require.NotNil(t, networkConfig.Signature)

			return networkConfig
		},
		peer2,
		false,
	}, {
		"invalid-signature",
		func(t *testing.T) *pb.NetworkConfig {
			cfg1 := pb.NewNetworkConfig(pb.NetworkState_PROTECTED)
			require.NoError(t, cfg1.Sign(context.Background(), sk1))

			cfg2 := pb.NewNetworkConfig(pb.NetworkState_PROTECTED)
			cfg2.Participants[peer1.Pretty()] = generatePeerAddrs(t, peer1)
			require.NoError(t, cfg2.Sign(context.Background(), sk1))

			cfg1.Signature = cfg2.Signature
			require.NotNil(t, cfg1.Signature)

			return cfg1
		},
		peer1,
		false,
	}, {
		"valid-signature",
		func(t *testing.T) *pb.NetworkConfig {
			networkConfig := pb.NewNetworkConfig(pb.NetworkState_PROTECTED)
			networkConfig.Participants[peer1.Pretty()] = generatePeerAddrs(t, peer1)
			networkConfig.Participants[peer2.Pretty()] = generatePeerAddrs(t, peer2)

			require.NoError(t, networkConfig.Sign(context.Background(), sk1))
			require.NotNil(t, networkConfig.Signature)

			return networkConfig
		},
		peer1,
		true,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			networkConfig := tt.networkConfig(t)
			valid := networkConfig.ValidateSignature(context.Background(), tt.signerID)
			assert.Equal(t, tt.expected, valid)

			if networkConfig != nil {
				assert.NotNil(t, networkConfig.Signature)
			}
		})
	}
}

func TestNetworkConfig_Save(t *testing.T) {
	sk1 := test.GeneratePrivateKey(t)
	peer1 := test.GetPeerIDFromKey(t, sk1)

	sk2 := test.GeneratePrivateKey(t)
	peer2 := test.GetPeerIDFromKey(t, sk2)

	testCases := []struct {
		name                 string
		saveInitialConfig    func(*testing.T) string
		expectedSignerID     peer.ID
		expectedErr          error
		validateLoadedConfig func(*testing.T, *pb.NetworkConfig)
	}{{
		"rejects-missing-file",
		func(*testing.T) string { return "not/a/valid/path/config.json" },
		peer1,
		pb.ErrInvalidConfig,
		func(*testing.T, *pb.NetworkConfig) {},
	}, {
		"rejects-invalid-file",
		func(t *testing.T) string {
			dir, _ := ioutil.TempDir("", "alice")
			path := path.Join(dir, "config.json")
			require.NoError(t, ioutil.WriteFile(path, []byte("not a valid format"), 0644), "ioutil.WriteFile()")
			return path
		},
		peer1,
		pb.ErrInvalidConfig,
		func(*testing.T, *pb.NetworkConfig) {},
	}, {
		"rejects-invalid-network-state",
		func(t *testing.T) string {
			dir, _ := ioutil.TempDir("", "alice")
			path := path.Join(dir, "config.json")

			networkConfig := pb.NewNetworkConfig(42)
			err := networkConfig.SaveToFile(context.Background(), path, sk1)
			require.NoError(t, err, "networkConfig.SaveToFile()")

			return path
		},
		peer1,
		pb.ErrInvalidNetworkState,
		func(*testing.T, *pb.NetworkConfig) {},
	}, {
		"rejects-signer-mismatch",
		func(t *testing.T) string {
			dir, _ := ioutil.TempDir("", "alice")
			path := path.Join(dir, "config.json")

			networkConfig := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)
			err := networkConfig.SaveToFile(context.Background(), path, sk1)
			require.NoError(t, err, "networkConfig.SaveToFile()")

			return path
		},
		peer2,
		pb.ErrInvalidSignature,
		func(*testing.T, *pb.NetworkConfig) {},
	}, {
		"saves-and-loads",
		func(t *testing.T) string {
			dir, _ := ioutil.TempDir("", "alice")
			path := path.Join(dir, "config.json")

			networkConfig := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)
			networkConfig.Participants[peer1.Pretty()] = generatePeerAddrs(t, peer1)
			err := networkConfig.SaveToFile(context.Background(), path, sk1)
			require.NoError(t, err, "networkConfig.SaveToFile()")

			return path
		},
		peer1,
		nil,
		func(t *testing.T, networkConfig *pb.NetworkConfig) {
			assert.Equal(t, pb.NetworkState_BOOTSTRAP, networkConfig.NetworkState)
			assert.Len(t, networkConfig.Participants, 1)
			assert.Contains(t, networkConfig.Participants, peer1.Pretty())
			assert.Len(t, networkConfig.Participants[peer1.Pretty()].Addresses, 1)
		},
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			path := tt.saveInitialConfig(t)

			networkConfig := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)
			err := networkConfig.LoadFromFile(ctx, path, tt.expectedSignerID)
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				tt.validateLoadedConfig(t, networkConfig)
			}
		})
	}
}
