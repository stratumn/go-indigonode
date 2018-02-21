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

package testutil

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protocol/coin/gossip"
	"github.com/stratumn/alice/core/protocol/coin/gossip/mockgossip"
)

// NewDummyGossip returns a gossip mock that does nothing.
func NewDummyGossip(t *testing.T) gossip.Gossip {
	mockCtrl := gomock.NewController(t)
	mockGossip := mockgossip.NewMockGossip(mockCtrl)

	mockGossip.EXPECT().ListenTx(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().ListenBlock(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().PublishTx(gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().PublishBlock(gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().Close().AnyTimes().Return(nil)

	return mockGossip
}
