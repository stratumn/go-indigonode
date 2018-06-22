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

package protectortest

import (
	"context"
	"testing"

	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protector/pb"
	"github.com/stratumn/alice/test"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// NewTestNetworkConfig creates a network config in the given state
// with the given peers allowed.
func NewTestNetworkConfig(t *testing.T, networkState pb.NetworkState, peers ...peer.ID) protector.NetworkConfig {
	ctx := context.Background()
	cfg, err := protector.NewInMemoryConfig(ctx, pb.NewNetworkConfig(networkState))
	require.NoError(t, err, "protector.NewInMemoryConfig()")

	for _, peerID := range peers {
		err := cfg.AddPeer(ctx, peerID, test.GeneratePeerMultiaddrs(t, peerID))
		require.NoError(t, err, "cfg.AddPeer()")
	}

	return cfg
}
