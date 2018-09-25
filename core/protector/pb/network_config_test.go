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

package pb_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	json "github.com/gibson042/canonicaljson-go"
	"github.com/gogo/protobuf/types"
	"github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
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
		"missing-timestamp",
		func(t *testing.T) *pb.NetworkConfig {
			networkConfig := pb.NewNetworkConfig(pb.NetworkState_PROTECTED)
			networkConfig.Participants[peer1.Pretty()] = generatePeerAddrs(t, peer1)
			networkConfig.Participants[peer2.Pretty()] = generatePeerAddrs(t, peer2)

			require.NoError(t, networkConfig.Sign(context.Background(), sk1))
			require.NotNil(t, networkConfig.Signature)

			networkConfig.LastUpdated = nil

			return networkConfig
		},
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

			// LastUpdated shouldn't be set by default.
			// Otherwise a new node risks rejecting valid configurations
			// that were signed before the node was created.
			// LastUpdated should come from a signed configuration (set by Sign()).
			require.Nil(t, networkConfig.LastUpdated)

			require.NoError(t, networkConfig.Sign(context.Background(), sk1))
			require.NotNil(t, networkConfig.Signature)
			require.NotNil(t, networkConfig.LastUpdated)

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
		&pb.NetworkConfig{
			NetworkState: pb.NetworkState_PROTECTED,
			LastUpdated:  types.TimestampNow(),
			Participants: map[string]*pb.PeerAddrs{
				peerID.Pretty(): &pb.PeerAddrs{
					Addresses: []string{test.GeneratePeerMultiaddr(t, peerID).String()},
				},
			},
		},
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
	dir, _ := ioutil.TempDir("", "indigo-node")

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
