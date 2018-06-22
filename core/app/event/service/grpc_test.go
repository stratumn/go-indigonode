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

package service

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/core/app/event/grpc"
	"github.com/stratumn/alice/core/app/event/grpc/mockgrpc"
	"github.com/stratumn/alice/core/app/event/service/mockservice"
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
