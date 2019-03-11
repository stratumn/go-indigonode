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

package protocol_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol"
	"github.com/stratumn/go-node/core/protector"
	"github.com/stratumn/go-node/test"
	"github.com/stratumn/go-node/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiformats/go-multiaddr"
)

func TestBootstrapNew(t *testing.T) {
	peerID := test.GeneratePeerID(t)
	peerAddr := test.GeneratePeerMultiaddr(t, peerID)

	testCases := []struct {
		name         string
		networkMode  *protector.NetworkMode
		expectedType interface{}
		expectedErr  error
	}{{
		"invalid-protection-mode",
		&protector.NetworkMode{ProtectionMode: "quantum"},
		nil,
		protocol.ErrInvalidProtectionMode,
	}, {
		"public-network",
		&protector.NetworkMode{},
		&protocol.PublicNetworkHandler{},
		nil,
	}, {
		"private-coordinator-node",
		&protector.NetworkMode{
			ProtectionMode: protector.PrivateWithCoordinatorMode,
			IsCoordinator:  true,
		},
		&protocol.CoordinatorHandler{},
		nil,
	}, {
		"private-with-coordinator",
		&protector.NetworkMode{
			ProtectionMode:   protector.PrivateWithCoordinatorMode,
			CoordinatorID:    peerID,
			CoordinatorAddrs: []multiaddr.Multiaddr{peerAddr},
		},
		&protocol.CoordinatedHandler{},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			host.EXPECT().SetStreamHandler(gomock.Any(), gomock.Any()).AnyTimes()

			h, err := protocol.New(host, nil, tt.networkMode, nil, nil)

			if tt.expectedErr == nil {
				require.NoError(t, err)
				require.NotNil(t, h)
				assert.IsType(t, tt.expectedType, h)
			} else {
				assert.EqualError(t, err, tt.expectedErr.Error())
			}
		})
	}
}
