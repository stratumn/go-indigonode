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

package protocol

import (
	"context"
	"testing"
	"time"

	"github.com/stratumn/go-node/core/p2p"
	"github.com/stretchr/testify/require"

	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	swarmtesting "gx/ipfs/QmeDpqUwwdye8ABKVMPXKuWwPVURFdqTqssbTUB39E2Nwd/go-libp2p-swarm/testing"
)

func TestClock_RemoteTime(t *testing.T) {
	ctx := context.Background()
	h1 := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	h2 := p2p.NewHost(ctx, swarmtesting.GenSwarm(t, ctx))
	defer h1.Close()
	defer h2.Close()

	// connect h1 to h2
	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	require.NoError(t, h1.Connect(ctx, h2pi), "Connect()")

	clockH2 := NewClock(h2, 10*time.Second)
	h2.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
		clockH2.StreamHandler(ctx, stream)
	})

	clockH1 := &Clock{host: h1}
	remoteTime, err := clockH1.RemoteTime(ctx, h2.ID())

	require.NoError(t, err, "RemoteTime()")
	require.NotNil(t, remoteTime)
	require.WithinDuration(t, time.Now().UTC(), *remoteTime, time.Second)
}
