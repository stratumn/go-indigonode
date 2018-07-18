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
	"github.com/stratumn/go-indigonode/app/coin/protocol/p2p"
	"github.com/stratumn/go-indigonode/app/coin/protocol/p2p/mockp2p"
)

// NewDummyP2P returns a p2p mock that does nothing.
func NewDummyP2P(t *testing.T) p2p.P2P {
	mockCtrl := gomock.NewController(t)
	mockP2P := mockp2p.NewMockP2P(mockCtrl)

	mockP2P.EXPECT().RequestBlockByHash(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	mockP2P.EXPECT().RequestBlocksByNumber(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	mockP2P.EXPECT().RequestHeaderByHash(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
	mockP2P.EXPECT().RequestHeadersByNumber(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	mockP2P.EXPECT().RespondBlockByHash(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockP2P.EXPECT().RespondBlocksByNumber(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockP2P.EXPECT().RespondHeaderByHash(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	mockP2P.EXPECT().RespondHeadersByNumber(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

	return mockP2P
}
