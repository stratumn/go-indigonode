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

package streamtest

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/core/streamutil"
	"github.com/stratumn/go-indigonode/core/streamutil/mockstream"
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

			return stream, err
		},
	)
}
