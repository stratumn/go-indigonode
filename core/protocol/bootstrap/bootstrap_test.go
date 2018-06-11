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

package bootstrap_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stratumn/alice/test"
	"github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
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
		bootstrap.ErrInvalidProtectionMode,
	}, {
		"public-network",
		&protector.NetworkMode{},
		&bootstrap.PublicNetworkHandler{},
		nil,
	}, {
		"private-coordinator-node",
		&protector.NetworkMode{
			ProtectionMode: protector.PrivateWithCoordinatorMode,
			IsCoordinator:  true,
		},
		&bootstrap.CoordinatorHandler{},
		nil,
	}, {
		"private-with-coordinator",
		&protector.NetworkMode{
			ProtectionMode:   protector.PrivateWithCoordinatorMode,
			CoordinatorID:    peerID,
			CoordinatorAddrs: []multiaddr.Multiaddr{peerAddr},
		},
		&bootstrap.CoordinatedHandler{},
		protector.ErrConnectionRefused,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			host := mocks.NewMockHost(ctrl)
			host.EXPECT().SetStreamHandler(gomock.Any(), gomock.Any()).AnyTimes()
			host.EXPECT().Peerstore().AnyTimes().Return(peerstore.NewPeerstore())
			host.EXPECT().Connect(gomock.Any(), gomock.Any()).AnyTimes().Return(protector.ErrConnectionRefused)

			h, err := bootstrap.New(ctx, host, tt.networkMode, nil, nil)

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
