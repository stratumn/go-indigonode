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
	"github.com/stratumn/alice/app/coin/protocol/p2p"
	"github.com/stratumn/alice/app/coin/protocol/p2p/mockp2p"
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
