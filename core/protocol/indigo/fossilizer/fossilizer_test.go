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

package fossilizer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/protocol/indigo/fossilizer"
	"github.com/stratumn/alice/core/protocol/indigo/fossilizer/mockbatchfossilizer"
	"github.com/stratumn/alice/core/protocol/indigo/fossilizer/mockfossilizer"
	indigofossilizer "github.com/stratumn/go-indigocore/fossilizer"
	"github.com/stretchr/testify/assert"
)

func TestNewFossilizer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	indigoFossilizer := mockfossilizer.NewMockAdapter(ctrl)
	f := fossilizer.New(indigoFossilizer)
	assert.NotNil(t, f)
}

func TestGetInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	indigoFossilizer := mockfossilizer.NewMockAdapter(ctrl)
	ctx := context.Background()

	f := fossilizer.New(indigoFossilizer)
	t.Run("successfully gets the data", func(t *testing.T) {
		info := "test"
		indigoFossilizer.EXPECT().GetInfo(ctx).Times(1).Return(info, nil)
		res, err := f.GetInfo(ctx)
		assert.NoError(t, err)
		assert.Equal(t, res, info)
	})

	t.Run("return an error accordingly", func(t *testing.T) {
		testErr := errors.New("test")
		indigoFossilizer.EXPECT().GetInfo(ctx).Times(1).Return(nil, testErr)
		res, err := f.GetInfo(ctx)
		assert.EqualError(t, err, testErr.Error())
		assert.Nil(t, res)
	})
}

func TestAddFossilizerEventChan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	indigoFossilizer := mockfossilizer.NewMockAdapter(ctrl)
	ctx := context.Background()

	f := fossilizer.New(indigoFossilizer)
	t.Run("calls the AddFossilizerEventChan method", func(t *testing.T) {
		eventCh := make(chan *indigofossilizer.Event, 1)
		indigoFossilizer.EXPECT().AddFossilizerEventChan(eventCh).Times(1)
		f.AddFossilizerEventChan(ctx, eventCh)
	})
}

func TestFossilize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	indigoFossilizer := mockfossilizer.NewMockAdapter(ctrl)
	ctx := context.Background()

	f := fossilizer.New(indigoFossilizer)
	t.Run("returns nil on success", func(t *testing.T) {
		indigoFossilizer.EXPECT().Fossilize(ctx, nil, nil).Times(1).Return(nil)
		assert.NoError(t, f.Fossilize(ctx, nil, nil))
	})

	t.Run("returns an error on failure", func(t *testing.T) {
		testErr := errors.New("test")
		indigoFossilizer.EXPECT().Fossilize(ctx, nil, nil).Times(1).Return(testErr)
		assert.EqualError(t, f.Fossilize(ctx, nil, nil), testErr.Error())
	})
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	t.Run("returns immediately when using a non-batch fossilizer", func(t *testing.T) {
		indigoFossilizer := mockfossilizer.NewMockAdapter(ctrl)
		f := fossilizer.New(indigoFossilizer)
		assert.NoError(t, f.Start(ctx))
	})

	t.Run("correctly starts a batch fossilizer", func(t *testing.T) {
		indigoBatchFossilizer := mockbatchfossilizer.NewMockAdapter(ctrl)
		bf := fossilizer.New(indigoBatchFossilizer)
		indigoBatchFossilizer.EXPECT().Start(ctx).Times(1).Return(nil)

		c := make(chan struct{}, 1)
		go func() {
			err := bf.Start(ctx)
			assert.NoError(t, err)
			c <- struct{}{}
		}()
		<-c
	})

	t.Run("returns from the goroutine on context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		indigoBatchFossilizer := mockbatchfossilizer.NewMockAdapter(ctrl)
		bf := fossilizer.New(indigoBatchFossilizer)
		indigoBatchFossilizer.EXPECT().Start(ctx).Times(1).DoAndReturn(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})

		c := make(chan struct{}, 1)
		go func() {
			err := bf.Start(ctx)
			assert.EqualError(t, err, context.Canceled.Error())
			c <- struct{}{}
		}()
		cancel()
		<-c
	})
}

func TestStarted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	t.Run("returns a channel returning immediately when using a non-batch fossilizer", func(t *testing.T) {
		indigoFossilizer := mockfossilizer.NewMockAdapter(ctrl)
		f := fossilizer.New(indigoFossilizer)
		startedChan := f.Started(ctx)
		select {
		case <-startedChan:
			break
		default:
			t.Error("Channel should already contain an event")
		}
	})

	t.Run("returns a channel that resolves only after the fossilizer has started", func(t *testing.T) {
		indigoBatchFossilizer := mockbatchfossilizer.NewMockAdapter(ctrl)
		bf := fossilizer.New(indigoBatchFossilizer)

		started := make(chan struct{}, 1)
		indigoBatchFossilizer.EXPECT().Started().Times(1).Return(started)

		go func() {
			// wait for 150ms before sending an event in the started channel.
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			<-ctx.Done()
			started <- struct{}{}
		}()

		startedChan := bf.Started(ctx)
		select {
		case <-startedChan:
			t.Error("Channel should not contain an event")
		default:
			break
		}

		// wait a little bit longer before checking if the channel contains an event.
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		<-ctx.Done()

		select {
		case <-startedChan:
			break
		default:
			t.Error("Channel should contain an event")

		}
	})
}
