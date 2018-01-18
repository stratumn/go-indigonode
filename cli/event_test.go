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
	"github.com/golang/protobuf/proto"
	"github.com/stratumn/alice/cli"
	"github.com/stratumn/alice/cli/mockcli"
	pbchat "github.com/stratumn/alice/grpc/chat"
	pbevent "github.com/stratumn/alice/grpc/event"
	"github.com/stretchr/testify/assert"
)

// IdleServerStream is a cli.ServerStream that doesn't push any message.
type IdleServerStream struct{}

func (_ IdleServerStream) RecvMsg() (proto.Message, error) {
	<-time.After(10 * time.Second)
	return nil, nil
}

// ClosedServerStream is a cli.ServerStream that closes the connection
// on the first attempt to receive.
type ClosedServerStream struct {
	mu     sync.RWMutex
	called bool
}

func (s *ClosedServerStream) RecvMsg() (proto.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.called = true
	return nil, io.EOF
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
	mu       sync.RWMutex
	messages []proto.Message
}

func (s *BoundedServerStream) RecvMsg() (proto.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.messages) == 0 {
		return nil, io.EOF
	}

	msg := s.messages[0]
	s.messages = s.messages[1:]

	return msg, nil
}

func (s *BoundedServerStream) WaitUntilClosed(t *testing.T) {
	waitUntil(t, func() bool {
		s.mu.RLock()
		defer s.mu.RUnlock()

		return len(s.messages) == 0
	})
}

func testConsoleRPCEventListener(t *testing.T, w io.Writer, mockServerStream cli.ServerStream) (cli.EventListener, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)

	mockSig := mockcli.NewMockSignaler(mockCtrl)
	mockSig.EXPECT().Signal(syscall.SIGWINCH).AnyTimes()

	mockStub := mockcli.NewMockServerStreamStub(mockCtrl)
	mockStub.EXPECT().Name().Return("test").AnyTimes()
	mockStub.EXPECT().Invoke(gomock.Any()).Return(mockServerStream, nil).AnyTimes()

	console := cli.NewConsole(w, false)

	el := cli.NewConsoleStubEventListener(console, mockStub, "test", mockSig)

	return el, mockCtrl
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

func TestConsoleRPCEventListener_GetService(t *testing.T) {
	el := cli.NewConsoleStubEventListener(nil, nil, "event", nil)

	assert.Equal(t, "event", el.Service(), "el.Service()")
}

func TestConsoleRPCEventListener_Start(t *testing.T) {
	el, mockCtrl := testConsoleRPCEventListener(t, ioutil.Discard, &IdleServerStream{})
	defer mockCtrl.Finish()

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
			messages: []proto.Message{
				&pbevent.Event{Message: "hello"},
				&pbevent.Event{Message: "world"},
			},
		}

		el, mockCtrl := testConsoleRPCEventListener(t, buf, mockServerStream)
		defer mockCtrl.Finish()

		ctx, cancel := context.WithCancel(context.Background())
		start(t, el, ctx)
		defer cancel()

		mockServerStream.WaitUntilClosed(t)
		waitUntilDisconnected(t, el)

		assert.Contains(buf.String(), "hello")
		assert.Contains(buf.String(), "world")
	})

	t.Run("Ignores invalid events", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		mockServerStream := &BoundedServerStream{
			messages: []proto.Message{
				&pbevent.Event{Message: "hello"},
				&pbchat.ChatMessage{Message: "chat"},
				&pbevent.Event{Message: "world"},
			},
		}

		el, mockCtrl := testConsoleRPCEventListener(t, buf, mockServerStream)
		defer mockCtrl.Finish()

		ctx, cancel := context.WithCancel(context.Background())
		start(t, el, ctx)
		defer cancel()

		mockServerStream.WaitUntilClosed(t)
		waitUntilDisconnected(t, el)

		assert.Contains(buf.String(), "hello")
		assert.Contains(buf.String(), "world")
		assert.NotContains(buf.String(), "chat")
	})
}

func TestConsoleRPCEventListener_Stop(t *testing.T) {
	assert := assert.New(t)

	mockServerStream := &ClosedServerStream{}
	el, mockCtrl := testConsoleRPCEventListener(t, ioutil.Discard, mockServerStream)
	defer mockCtrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	start(t, el, ctx)

	mockServerStream.WaitUntilCalled(t)
	waitUntilDisconnected(t, el)
	cancel()

	assert.False(el.Connected(), "el.Connected()")
}
