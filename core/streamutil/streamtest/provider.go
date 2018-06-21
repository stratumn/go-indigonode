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

package streamtest

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/streamutil"
	"github.com/stratumn/alice/core/streamutil/mockstream"
	"github.com/stretchr/testify/assert"

	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

// ExpectStreamPeerAndProtocol configures a mock provider to
// expect a stream to the given peer with the given protocol.
// It then returns the given stream and error.
func ExpectStreamPeerAndProtocol(
	t *testing.T,
	p *mockstream.MockProvider,
	peerID peer.ID,
	pid protocol.ID,
	stream streamutil.Stream,
	err error,
) {
	p.EXPECT().NewStream(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ ihost.Host, opts ...streamutil.StreamOption) (streamutil.Stream, error) {
			streamOpts := &streamutil.StreamOptions{}
			for _, opt := range opts {
				opt(streamOpts)
			}

			assert.Equal(t, peerID, streamOpts.PeerID)
			assert.Len(t, streamOpts.PIDs, 1)
			assert.Equal(t, pid, streamOpts.PIDs[0])
			assert.NotNil(t, streamOpts.Event)

			return stream, err
		},
	)
}
