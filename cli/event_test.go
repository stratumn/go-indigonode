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

package cli_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
	pbevent "github.com/stratumn/alice/core/app/event/grpc"
	"github.com/stratumn/alice/core/app/event/grpc/mockgrpc"
	"github.com/stretchr/testify/assert"
)

// ClosedServerStream is a cli.ServerStream that closes the connection
// on the first attempt to receive.
type ClosedServerStream struct {
	mu     sync.RWMutex
	called bool
}

func (s *ClosedServerStream) RecvMsg(interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.called = true
	return io.EOF
}

func (s *ClosedServerStream) WaitUntilCalled(t *testing.T) {
	waitUntil(t, func() bool {
		s.mu.RLock()
		defer s.mu.RUnlock()

		return s.called
	})
}

// BoundedServerStream returns a configured number of messages and then
// closes the connection.
type BoundedServerStream struct {
	mu     sync.RWMutex
	events []*pbevent.Event
}

func (s *BoundedServerStream) RecvMsg(m interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.events) == 0 {
		return io.EOF
	}

	e := s.events[0]

	ev, _ := m.(*pbevent.Event)
	ev.Message = e.Message
	ev.Level = e.Level

	s.events = s.events[1:]
	return nil
}

func (s *BoundedServerStream) WaitUntilClosed(t *testing.T) {
	waitUntil(t, func() bool {
		s.mu.RLock()
		defer s.mu.RUnlock()

		return len(s.events) == 0
	})
}

func testConsoleRPCEventListener(t *testing.T, w io.Writer) (cli.EventListener, *gomock.Controller, *mockgrpc.MockEmitter_ListenClient) {
	mockCtrl := gomock.NewController(t)

	mockSig := mockcli.NewMockSignaler(mockCtrl)
	mockSig.EXPECT().Signal(syscall.SIGWINCH).AnyTimes()

	cons := cli.NewConsole(w, false)

	elc := mockgrpc.NewMockEmitter_ListenClient(mockCtrl)

	client := mockgrpc.NewMockEmitterClient(mockCtrl)
	client.EXPECT().Listen(gomock.Any(), gomock.Any()).Return(elc, nil).AnyTimes()

	el := cli.NewConsoleClientEventListener(cons, client, mockSig)

	return el, mockCtrl, elc
}

func waitUntil(t *testing.T, cond func() bool) {
	condChan := make(chan struct{})
	go func() {
		for {
			if cond() {
				condChan <- struct{}{}
				return
			} else {
				runtime.Gosched()
			}
		}
	}()

	select {
	case <-condChan:
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "waitUntil")
	}
}

func waitUntilConnected(t *testing.T, el cli.EventListener) {
	waitUntil(t, func() bool { return el.Connected() })
}

func waitUntilDisconnected(t *testing.T, el cli.EventListener) {
	waitUntil(t, func() bool { return !el.Connected() })
}

func start(t *testing.T, el cli.EventListener, ctx context.Context) {
	go func() {
		err := el.Start(ctx)
		assert.NoError(t, err, "el.Start()")
	}()
}

func TestConsoleRPCEventListener_Start(t *testing.T) {
	el, mockCtrl, elc := testConsoleRPCEventListener(t, ioutil.Discard)
	defer mockCtrl.Finish()

	elc.EXPECT().RecvMsg(gomock.Any()).Return(nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	start(t, el, ctx)
	defer cancel()

	waitUntilConnected(t, el)

	start(t, el, ctx)

	// Wait a bit to give time to disconnect previous connection.
	<-time.After(10 * time.Millisecond)

	waitUntilConnected(t, el)
}

func TestConsoleRPCEventListener_Print(t *testing.T) {
	assert := assert.New(t)

	t.Run("Prints incoming events", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		mockServerStream := &BoundedServerStream{
			events: []*pbevent.Event{
				{Message: "hello", Level: pbevent.Level_INFO},
				{Message: "world", Level: pbevent.Level_INFO},
			},
		}

		el, mockCtrl, elc := testConsoleRPCEventListener(t, buf)
		defer mockCtrl.Finish()

		elc.EXPECT().RecvMsg(gomock.Any()).DoAndReturn(mockServerStream.RecvMsg).Times(3)

		ctx, cancel := context.WithCancel(context.Background())
		start(t, el, ctx)
		defer cancel()

		mockServerStream.WaitUntilClosed(t)
		waitUntilDisconnected(t, el)

		assert.Contains(buf.String(), "hello")
		assert.Contains(buf.String(), "world")
	})
}

func TestConsoleRPCEventListener_Stop(t *testing.T) {
	assert := assert.New(t)

	mockServerStream := &ClosedServerStream{}
	el, mockCtrl, elc := testConsoleRPCEventListener(t, ioutil.Discard)
	defer mockCtrl.Finish()

	elc.EXPECT().RecvMsg(gomock.Any()).DoAndReturn(mockServerStream.RecvMsg).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	start(t, el, ctx)

	mockServerStream.WaitUntilCalled(t)
	waitUntilDisconnected(t, el)
	cancel()

	assert.False(el.Connected(), "el.Connected()")
}
