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

package testutil

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/gossip"
	"github.com/stratumn/go-indigonode/app/coin/protocol/gossip/mockgossip"
)

// NewDummyGossip returns a gossip mock that does nothing.
func NewDummyGossip(t *testing.T) gossip.Gossip {
	mockCtrl := gomock.NewController(t)
	mockGossip := mockgossip.NewMockGossip(mockCtrl)

	mockGossip.EXPECT().ListenTx(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().ListenBlock(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().PublishTx(gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().PublishBlock(gomock.Any()).AnyTimes().Return(nil)
	mockGossip.EXPECT().AddBlockListener().AnyTimes().Return(make(chan *pb.Header))
	mockGossip.EXPECT().Close().AnyTimes().Return(nil)

	return mockGossip
}
