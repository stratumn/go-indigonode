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
