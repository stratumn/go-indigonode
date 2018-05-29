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
	"github.com/stratumn/alice/test/mocks"
	"github.com/stretchr/testify/require"
)

func TestCoordinator_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	host := mocks.NewMockHost(ctrl)
	host.EXPECT().SetStreamHandler(bootstrap.PrivateWithCoordinatorProtocolID, gomock.Any()).Times(1)

	handler, err := bootstrap.NewCoordinatorHandler(
		host,
		protector.NewCoordinatorNetworkMode(),
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, handler)

	host.EXPECT().RemoveStreamHandler(bootstrap.PrivateWithCoordinatorProtocolID).Times(1)
	handler.Close(context.Background())
}
