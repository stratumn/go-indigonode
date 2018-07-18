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

package service

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/go-indigonode/core/app/event/grpc"
	"github.com/stratumn/go-indigonode/core/app/event/grpc/mockgrpc"
	"github.com/stratumn/go-indigonode/core/app/event/service/mockservice"
	"github.com/stretchr/testify/assert"
)

func testGRPCServer(ctx context.Context, t *testing.T, eventEmitter Emitter) grpcServer {
	return grpcServer{func() Emitter { return eventEmitter }}
}

func TestGRPCServer_Listen_Add_Remove_Listeners(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	ctrl := gomock.NewController(t)
	defer func() {
		cancel()
		ctrl.Finish()
	}()

	mockEmitter := mockservice.NewMockEmitter(ctrl)
	srv := testGRPCServer(ctx, t, mockEmitter)
	ss := mockgrpc.NewMockEmitter_ListenServer(ctrl)

	ss.EXPECT().Context().AnyTimes().Return(ctx)
	addListener := mockEmitter.EXPECT().AddListener("topic").Times(1)
	mockEmitter.EXPECT().RemoveListener(gomock.Any()).After(addListener).Times(1)

	assert.NoError(t, srv.Listen(&grpc.ListenReq{Topic: "topic"}, ss), "srv.Listen()")
}

func TestGRPCServer_Listen_Send_Events(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	ctrl := gomock.NewController(t)
	defer func() {
		cancel()
		ctrl.Finish()
	}()

	emitter := NewEmitter(DefaultTimeout)
	srv := testGRPCServer(ctx, t, emitter)
	ss := mockgrpc.NewMockEmitter_ListenServer(ctrl)

	ss.EXPECT().Context().AnyTimes().Return(ctx)

	errChan := make(chan error)
	go func() {
		err := srv.Listen(&grpc.ListenReq{Topic: "topic"}, ss)
		errChan <- err
	}()

	// We wait for the server to register a listener before emitting.
	for {
		if emitter.GetListenersCount("topic") == 1 {
			break
		} else {
			runtime.Gosched()
		}
	}

	e := &grpc.Event{
		Message: "Hello",
		Level:   grpc.Level_INFO,
		Topic:   "topic",
	}
	ss.EXPECT().Send(e)
	emitter.Emit(e)

	assert.NoError(t, <-errChan, "srv.Listen()")
}
