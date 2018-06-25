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
	"github.com/pkg/errors"
	"github.com/stratumn/go-indigonode/core/streamutil"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stratumn/go-indigonode/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
)

func TestNewStream(t *testing.T) {
	ctx := context.Background()
	provider := streamutil.NewStreamProvider()

	peerID := test.GeneratePeerID(t)
	protocolID := protocol.ID("/indigo/node/test/streamutil/v1.0.0")

	t.Run("reject-missing-peer-id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)

		s, err := provider.NewStream(ctx, host)
		assert.EqualError(t, err, streamutil.ErrMissingPeerID.Error())
		assert.Nil(t, s)
	})

	t.Run("reject-missing-protocol", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)

		s, err := provider.NewStream(ctx, host, streamutil.OptPeerID(peerID))
		assert.EqualError(t, err, streamutil.ErrMissingProtocolIDs.Error())
		assert.Nil(t, s)
	})

	t.Run("return-host-error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		host := mocks.NewMockHost(ctrl)
		host.EXPECT().NewStream(gomock.Any(), peerID, protocolID).Times(1).Return(nil, errors.New("too bad"))

		s, err := provider.NewStream(
			ctx,
			host,
			streamutil.OptPeerID(peerID),
			streamutil.OptProtocolIDs(protocolID),
		)
		assert.EqualError(t, err, "too bad")
		assert.Nil(t, s)
	})

	t.Run("close-stream", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mocks.NewMockConn(ctrl)

		stream := mocks.NewMockStream(ctrl)
		stream.EXPECT().Conn().Times(1).Return(conn)

		host := mocks.NewMockHost(ctrl)
		host.EXPECT().NewStream(gomock.Any(), peerID, protocolID).Times(1).Return(stream, nil)

		s, err := provider.NewStream(
			ctx,
			host,
			streamutil.OptPeerID(peerID),
			streamutil.OptProtocolIDs(protocolID),
		)
		require.NoError(t, err)
		require.NotNil(t, s.Codec())
		require.Equal(t, conn, s.Conn())

		stream.EXPECT().Close().Times(1)
		s.Close()
	})
}
