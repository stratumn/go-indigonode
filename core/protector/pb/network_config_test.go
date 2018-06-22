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

package pb_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/stratumn/alice/core/protector/pb"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

func generatePeerAddrs(t *testing.T, peerID peer.ID) *pb.PeerAddrs {
	if peerID == "" {
		return &pb.PeerAddrs{
			Addresses: []string{test.GenerateMultiaddr(t).String()},
		}
	}

	return &pb.PeerAddrs{
		Addresses: []string{test.GeneratePeerMultiaddr(t, peerID).String()},
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

func TestNetworkConfig_ValidateContent(t *testing.T) {
	peerID := test.GeneratePeerID(t)

	testCases := []struct {
		name          string
		networkConfig *pb.NetworkConfig
		err           error
	}{{
		"invalid-network-state",
		&pb.NetworkConfig{NetworkState: 42},
		pb.ErrInvalidNetworkState,
	}, {
		"invalid-peer-id",
		&pb.NetworkConfig{
			Participants: map[string]*pb.PeerAddrs{
				"b4tm4n": generatePeerAddrs(t, ""),
			},
		},
		pb.ErrInvalidPeerID,
	}, {
		"missing-peer-address",
		&pb.NetworkConfig{
			Participants: map[string]*pb.PeerAddrs{
				peerID.Pretty(): nil,
			},
		},
		pb.ErrMissingPeerAddrs,
	}, {
		"invalid-peer-address",
		&pb.NetworkConfig{
			Participants: map[string]*pb.PeerAddrs{
				peerID.Pretty(): &pb.PeerAddrs{
					Addresses: []string{"not/a/multi/addr"},
				},
			},
		},
		pb.ErrInvalidPeerAddr,
	}, {
		"valid-config",
		pb.NewNetworkConfig(pb.NetworkState_PROTECTED),
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.networkConfig.ValidateContent(context.Background())
			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}

func TestNetworkConfig_Load(t *testing.T) {
	dir, _ := ioutil.TempDir("", "alice")

	signerKey := test.GeneratePrivateKey(t)
	signerID := test.GetPeerIDFromKey(t, signerKey)

	testCases := []struct {
		name       string
		createFile func(*testing.T) string
		validate   func(*testing.T, *pb.NetworkConfig)
		err        error
	}{{
		"file-missing",
		func(t *testing.T) string {
			return "does-not-exist.json"
		},
		nil,
		pb.ErrInvalidConfig,
	}, {
		"invalid-file-format",
		func(t *testing.T) string {
			err := ioutil.WriteFile(path.Join(dir, "invalid.json"), []byte("not hotdog"), 0644)
			require.NoError(t, err, "ioutil.WriteFile()")
			return "invalid.json"
		},
		nil,
		pb.ErrInvalidConfig,
	}, {
		"invalid-peer-id",
		func(t *testing.T) string {
			networkConfig := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)
			networkConfig.Participants["b4tm4n"] = generatePeerAddrs(t, signerID)

			configBytes, err := json.Marshal(networkConfig)
			require.NoError(t, err, "json.Marshal()")

			err = ioutil.WriteFile(path.Join(dir, "invalid-peer-id.json"), configBytes, 0644)
			require.NoError(t, err, "ioutil.WriteFile()")

			return "invalid-peer-id.json"
		},
		nil,
		pb.ErrInvalidPeerID,
	}, {
		"invalid-network-state",
		func(t *testing.T) string {
			networkConfig := &pb.NetworkConfig{NetworkState: 42}

			configBytes, err := json.Marshal(networkConfig)
			require.NoError(t, err, "json.Marshal()")

			err = ioutil.WriteFile(path.Join(dir, "invalid-network-state.json"), configBytes, 0644)
			require.NoError(t, err, "ioutil.WriteFile()")

			return "invalid-network-state.json"
		},
		nil,
		pb.ErrInvalidNetworkState,
	}, {
		"invalid-config-signature",
		func(t *testing.T) string {
			networkConfig := pb.NewNetworkConfig(pb.NetworkState_BOOTSTRAP)

			unknownKey := test.GeneratePrivateKey(t)
			err := networkConfig.Sign(context.Background(), unknownKey)
			require.NoError(t, err, "networkConfig.Sign()")

			configBytes, err := json.Marshal(networkConfig)
			require.NoError(t, err, "json.Marshal()")

			err = ioutil.WriteFile(path.Join(dir, "invalid-signature.json"), configBytes, 0644)
			require.NoError(t, err, "ioutil.WriteFile()")

			return "invalid-signature.json"
		},
		nil,
		pb.ErrInvalidSignature,
	}, {
		"valid-config",
		func(t *testing.T) string {
			networkConfig := pb.NewNetworkConfig(pb.NetworkState_PROTECTED)
			networkConfig.Participants[signerID.Pretty()] = generatePeerAddrs(t, signerID)

			err := networkConfig.Sign(context.Background(), signerKey)
			require.NoError(t, err, "networkConfig.Sign()")

			configBytes, err := json.Marshal(networkConfig)
			require.NoError(t, err, "json.Marshal()")

			err = ioutil.WriteFile(path.Join(dir, "config.json"), configBytes, 0644)
			require.NoError(t, err, "ioutil.WriteFile()")

			return "config.json"
		},
		func(t *testing.T, networkConfig *pb.NetworkConfig) {
			assert.Equal(t, pb.NetworkState_PROTECTED, networkConfig.NetworkState)
			assert.Len(t, networkConfig.Participants, 1)
			assert.Contains(t, networkConfig.Participants, signerID.Pretty())
		},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			configPath := path.Join(dir, tt.createFile(t))
			networkConfig := &pb.NetworkConfig{}
			err := networkConfig.LoadFromFile(ctx, configPath, signerID)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				tt.validate(t, networkConfig)
			}
		})
	}
}
