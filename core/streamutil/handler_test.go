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
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/core/monitoring"
	"github.com/stratumn/go-node/core/streamutil"
	"github.com/stratumn/go-node/test"
	"github.com/stratumn/go-node/test/mocks"

	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

func TestAutoCloseHandler(t *testing.T) {
	peerID := test.GeneratePeerID(t)

	t.Run("close-stream", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(peerID)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn)
		stream.EXPECT().Protocol().Return(protocol.ID("test"))

		handler := streamutil.WithAutoClose(
			"handler",
			"test",
			func(ctx context.Context, span *monitoring.Span, s inet.Stream, c streamutil.Codec) error {
				return nil
			},
		)

		// FullClose closes the stream and sets a deadline before trying to
		// read an EOF.
		// See https://github.com/libp2p/go-libp2p/blob/master/NEWS.md#zombie-streams
		stream.EXPECT().Close()
		stream.EXPECT().SetDeadline(gomock.Any())
		stream.EXPECT().Read(gomock.Any()).Return(0, io.EOF)

		handler(stream)
	})
}
