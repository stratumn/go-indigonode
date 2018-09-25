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
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/streamutil"
	"github.com/stratumn/go-node/test"
	"github.com/stratumn/go-node/test/mocks"
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
