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
type ClosedServerStream struct{}

func (_ ClosedServerStream) RecvMsg() (proto.Message, error) {
	return nil, io.EOF
}

// BoundedServerStream returns a configured number of messages and then
// closes the connection.
type BoundedServerStream struct {
	messages []proto.Message
}

func (s *BoundedServerStream) RecvMsg() (proto.Message, error) {
	if len(s.messages) == 0 {
		return nil, io.EOF
	}

	msg := s.messages[0]
	s.messages = s.messages[1:]

	return msg, nil
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

func waitUntilDisconnected(t *testing.T, el cli.EventListener) {
	disconnectChan := make(chan struct{})
	go func() {
		for {
			if !el.Connected() {
				disconnectChan <- struct{}{}
				return
			} else {
				runtime.Gosched()
			}
		}
	}()

	select {
	case <-disconnectChan:
		assert.False(t, el.Connected(), "el.Connected()")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "el.Connected() didn't disconnect in time")
	}
}

func TestConsoleRPCEventListener_GetService(t *testing.T) {
	el := cli.NewConsoleStubEventListener(nil, nil, "event", nil)

	assert.Equal(t, "event", el.Service(), "el.Service()")
}

func TestConsoleRPCEventListener_Start(t *testing.T) {
	assert := assert.New(t)

	el, mockCtrl := testConsoleRPCEventListener(t, ioutil.Discard, &IdleServerStream{})
	defer mockCtrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	err := el.Start(ctx)
	defer cancel()

	assert.NoError(err, "el.Start()")
	assert.True(el.Connected(), "el.Connected()")

	err = el.Start(context.Background())
	assert.NoError(err, "el.Start()")
	assert.True(el.Connected(), "el.Connected()")
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
		err := el.Start(ctx)
		defer cancel()

		assert.NoError(err, "el.Start()")
		assert.True(el.Connected(), "el.Connected()")

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
		err := el.Start(ctx)
		defer cancel()

		assert.NoError(err, "el.Start()")
		assert.True(el.Connected(), "el.Connected()")

		waitUntilDisconnected(t, el)

		assert.Contains(buf.String(), "hello")
		assert.Contains(buf.String(), "world")
		assert.NotContains(buf.String(), "chat")
	})
}

func TestConsoleRPCEventListener_Stop(t *testing.T) {
	assert := assert.New(t)

	el, mockCtrl := testConsoleRPCEventListener(t, ioutil.Discard, &ClosedServerStream{})
	defer mockCtrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(el.Start(ctx), "el.Start()")

	waitUntilDisconnected(t, el)
	cancel()

	assert.False(el.Connected(), "el.Connected()")
}
