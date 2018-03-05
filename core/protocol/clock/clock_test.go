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

package clock

import (
	"context"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
	testutil "gx/ipfs/QmYVR3C8DWPHdHxvLtNFYfjsXgaRAdh6hPMNH3KiwCgu4o/go-libp2p-netutil"

	"testing"
	"time"

	"github.com/stratumn/alice/core/p2p"
	"github.com/stretchr/testify/require"
)

func TestClock_RemoteTime(t *testing.T) {
	ctx := context.Background()
	h1 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
	h2 := p2p.NewHost(ctx, testutil.GenSwarmNetwork(t, ctx))
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
