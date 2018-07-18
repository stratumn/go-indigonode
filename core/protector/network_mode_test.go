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

package protector_test

import (
	"testing"

	"github.com/stratumn/go-indigonode/core/protector"
	"github.com/stratumn/go-indigonode/test"
	"github.com/stretchr/testify/assert"
)

func TestNewCoordinatorNetworkMode(t *testing.T) {
	mode := protector.NewCoordinatorNetworkMode()
	assert.Equal(t, protector.PrivateWithCoordinatorMode, mode.ProtectionMode)
	assert.True(t, mode.IsCoordinator)
}

func TestNewCoordinatedMode(t *testing.T) {
	peerID := test.GeneratePeerID(t)
	addr1 := test.GeneratePeerMultiaddr(t, peerID)
	addr2 := test.GeneratePeerMultiaddr(t, peerID)

	testCases := []struct {
		name             string
		coordinatorID    string
		coordinatorAddrs []string
		err              error
	}{{
		"invalid-coordinator-id",
		"b4tm4n",
		[]string{addr1.String()},
		protector.ErrInvalidCoordinatorID,
	}, {
		"missing-coordinator-addrs",
		peerID.Pretty(),
		nil,
		protector.ErrMissingCoordinatorAddr,
	}, {
		"invalid-coordinator-addr",
		peerID.Pretty(),
		[]string{"/not/a/multiaddr"},
		protector.ErrInvalidCoordinatorAddr,
	}, {
		"valid-coordinator",
		peerID.Pretty(),
		[]string{addr1.String(), addr2.String()},
		nil,
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := protector.NewCoordinatedNetworkMode(
				tt.coordinatorID,
				tt.coordinatorAddrs,
			)

			if tt.err == nil {
				assert.Nil(t, err)
				assert.Equal(t, protector.PrivateWithCoordinatorMode, mode.ProtectionMode)
				assert.False(t, mode.IsCoordinator)
				assert.Equal(t, tt.coordinatorID, mode.CoordinatorID.Pretty())
				assert.Equal(t, len(tt.coordinatorAddrs), len(mode.CoordinatorAddrs))
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
