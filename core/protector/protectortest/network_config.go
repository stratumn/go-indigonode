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

package protectortest

import (
	"context"
	"testing"

	"github.com/stratumn/go-indigonode/core/protector"
	"github.com/stratumn/go-indigonode/core/protector/pb"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
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
