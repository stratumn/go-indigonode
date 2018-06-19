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

package streamutil_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/streamutil"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/alice/test/mocks"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
)

var testLogger = logging.Logger("test")

func TestAutoCloseHandler(t *testing.T) {
	peerID := test.GeneratePeerID(t)

	t.Run("close-stream", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(peerID).Times(1)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn).Times(1)

		handler := streamutil.WithAutoClose(
			testLogger,
			"test",
			func(ctx context.Context, event *logging.EventInProgress, s inet.Stream, c streamutil.Codec) error {
				return nil
			},
		)

		stream.EXPECT().Close().Times(1)

		handler(stream)
	})

	t.Run("log-handler-error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mocks.NewMockConn(ctrl)
		conn.EXPECT().RemotePeer().Return(peerID).Times(1)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Return(conn).Times(1)
		stream.EXPECT().Close().Times(1)

		event := testLogger.EventBegin(context.Background(), "test")
		logger := mocks.NewMockEventLogger(ctrl)
		logger.EXPECT().EventBegin(gomock.Any(), "test", gomock.Any()).Return(event).Times(1)

		handler := streamutil.WithAutoClose(
			logger,
			"test",
			func(ctx context.Context, event *logging.EventInProgress, s inet.Stream, c streamutil.Codec) error {
				return nil
			},
		)

		handler(stream)
	})
}
