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

package protocol

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/app/coin/pb"
	"github.com/stratumn/go-indigonode/app/coin/protocol/engine/mockengine"
	"github.com/stratumn/go-indigonode/app/coin/protocol/processor/mockprocessor"
	"github.com/stratumn/go-indigonode/app/coin/protocol/synchronizer/mocksynchronizer"
	ctestutil "github.com/stratumn/go-indigonode/app/coin/protocol/testutil"
	"github.com/stratumn/go-indigonode/app/coin/protocol/validator/mockvalidator"
	"github.com/stretchr/testify/assert"
)

func TestSynchronize(t *testing.T) {
	genBlock := &pb.Block{Header: &pb.Header{Nonce: 42, Version: 32}}
	ch := &ctestutil.SimpleChain{}
	ch.AddBlock(genBlock)
	ch.SetHead(genBlock)

	// Run coin.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mockprocessor.NewMockProcessor(ctrl)
	val := mockvalidator.NewMockValidator(ctrl)
	eng := mockengine.NewMockPoW(ctrl)
	sync := mocksynchronizer.NewMockSynchronizer(ctrl)

	tests := map[string]func(*testing.T, *Coin){
		"append-blocks-in-order": func(t *testing.T, coin *Coin) {
			block1 := &pb.Block{Header: &pb.Header{PreviousHash: []byte("plap")}}
			block2 := &pb.Block{Header: &pb.Header{PreviousHash: []byte("zou")}}

			hash := []byte("plap")
			resCh := make(chan *pb.Block)
			errCh := make(chan error)

			gomock.InOrder(
				sync.EXPECT().Synchronize(gomock.Any(), hash, coin.chain).Return(resCh, errCh).Times(1),
				val.EXPECT().ValidateBlock(block1, gomock.Any()).Return(nil).Times(1),
				eng.EXPECT().VerifyHeader(gomock.Any(), block1.Header).Return(nil).Times(1),
				proc.EXPECT().Process(gomock.Any(), block1, gomock.Any(), gomock.Any()).Return(nil).Times(1),
				val.EXPECT().ValidateBlock(block2, gomock.Any()).Return(nil).Times(1),
				eng.EXPECT().VerifyHeader(gomock.Any(), block2.Header).Return(nil).Times(1),
				proc.EXPECT().Process(gomock.Any(), block2, gomock.Any(), gomock.Any()).Return(nil).Times(1),
			)

			go func() {
				resCh <- block1
				resCh <- block2
				close(resCh)
			}()

			err := coin.synchronize(context.Background(), hash)
			assert.NoError(t, err, "synchronize()")
		},
		"sync-fail": func(t *testing.T, coin *Coin) {
			block := &pb.Block{Header: &pb.Header{PreviousHash: []byte("plap")}}

			hash := []byte("plap")
			resCh := make(chan *pb.Block)
			errCh := make(chan error)

			gomock.InOrder(
				sync.EXPECT().Synchronize(gomock.Any(), hash, coin.chain).Return(resCh, errCh).Times(1),
				val.EXPECT().ValidateBlock(block, gomock.Any()).Return(nil).Times(1),
				eng.EXPECT().VerifyHeader(gomock.Any(), block.Header).Return(nil).Times(1),
				proc.EXPECT().Process(gomock.Any(), block, gomock.Any(), gomock.Any()).Return(nil).Times(1),
			)

			gandalf := errors.New("you shall not pass")

			go func() {
				resCh <- block
				errCh <- gandalf
			}()

			err := coin.synchronize(context.Background(), hash)
			assert.EqualError(t, err, gandalf.Error(), "synchronize()")
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			coin := NewCoinBuilder(t).
				WithChain(ch).
				WithProcessor(proc).
				WithValidator(val).
				WithEngine(eng).
				WithSynchronizer(sync).
				WithGenesisBlock(genBlock).
				Build(t)

			err := coin.Run(context.Background())
			assert.NoError(t, err, "coin.Run()")

			test(t, coin)
		})
	}
}
