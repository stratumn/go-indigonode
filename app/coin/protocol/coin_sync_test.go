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

package protocol

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-node/app/coin/pb"
	"github.com/stratumn/go-node/app/coin/protocol/engine/mockengine"
	"github.com/stratumn/go-node/app/coin/protocol/processor/mockprocessor"
	"github.com/stratumn/go-node/app/coin/protocol/synchronizer/mocksynchronizer"
	ctestutil "github.com/stratumn/go-node/app/coin/protocol/testutil"
	"github.com/stratumn/go-node/app/coin/protocol/validator/mockvalidator"
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
