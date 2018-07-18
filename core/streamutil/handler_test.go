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

package streamutil_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/core/monitoring"
	"github.com/stratumn/go-indigonode/core/streamutil"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stratumn/go-indigonode/test/mocks"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

func TestAutoCloseHandler(t *testing.T) {
	peerID := test.GeneratePeerID(t)

	t.Run("close-stream", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(peerID).Times(1)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn).Times(1)
		stream.EXPECT().Protocol().Return(protocol.ID("test"))

		handler := streamutil.WithAutoClose(
			"handler",
			"test",
			func(ctx context.Context, span *monitoring.Span, s inet.Stream, c streamutil.Codec) error {
				return nil
			},
		)

		stream.EXPECT().Close().Times(1)

		handler(stream)
	})
}
