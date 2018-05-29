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

	"github.com/stratumn/alice/core/protector"
	"github.com/stratumn/alice/core/protocol/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapNew(t *testing.T) {
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
	}}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			h, err := bootstrap.New(ctx, nil, tt.networkMode, nil)

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
