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

//go:generate mockgen -package mockcli -destination mockcli/mocksignaler.go github.com/stratumn/alice/cli Signaler
//go:generate mockgen -package mockcli -destination mockcli/mockserverstream.go github.com/stratumn/alice/cli ServerStream
//go:generate mockgen -package mockcli -destination mockcli/mockserverstreamstub.go github.com/stratumn/alice/cli ServerStreamStub

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/pkg/errors"
	pbevent "github.com/stratumn/alice/grpc/event"

	"google.golang.org/grpc"
)

// Signaler is an interface to send signals to the OS.
type Signaler interface {
	// Signal sends a signal to the OS.
	Signal(os.Signal) error
}

// ServerStream is an interface to receive messages from a streaming server.
type ServerStream interface {
	// RecvMsg receives a protobuf message.
	RecvMsg() (proto.Message, error)
}

// ServerStreamStub is an interface to connect to a streaming server.
type ServerStreamStub interface {
	// Invoke connects to a streaming server.
	Invoke(context.Context) (ServerStream, error)

	// Name returns the name of the invoked endpoint.
	Name() string
}

// rpcServerStreamStub implements ServerStreamStub.
// It connects to a gRPC streaming server for a specific method.
type rpcServerStreamStub struct {
	stub       grpcdynamic.Stub
	methodDesc *desc.MethodDescriptor
}

// newRPCServerStreamStub returns a ServerStreamStub that connects to a
// specific gRPC streaming server endpoint.
func newRPCServerStreamStub(conn *grpc.ClientConn, methodDesc *desc.MethodDescriptor) ServerStreamStub {
	eventRegistry := dynamic.NewKnownTypeRegistryWithDefaults()
	eventRegistry.AddKnownType((*pbevent.Event)(nil))
	mf := dynamic.NewMessageFactoryWithKnownTypeRegistry(eventRegistry)
	stub := grpcdynamic.NewStubWithMessageFactory(conn, mf)

	return &rpcServerStreamStub{
		stub:       stub,
		methodDesc: methodDesc,
	}
}

// Name returns the name of the invoked endpoint.
func (s *rpcServerStreamStub) Name() string {
	return s.methodDesc.GetName()
}

// Invoke connects to a streaming server.
func (s *rpcServerStreamStub) Invoke(ctx context.Context) (ServerStream, error) {
	req := dynamic.NewMessage(s.methodDesc.GetInputType())
	return s.stub.InvokeRpcServerStream(ctx, s.methodDesc, req)
}

// ConsoleRPCEventListener implements the EventListener interface.
// It connects to the RPC server's Listen endpoint and prints events to
// the console.
type ConsoleRPCEventListener struct {
	service string

	cons *Console
	sig  Signaler

	stub ServerStreamStub

	mu        sync.RWMutex
	connected bool
	close     context.CancelFunc
}

// NewConsoleRPCEventListener creates a new ConsoleRPCEventListener
// connected to a given console and RPC Listen endpoint.
func NewConsoleRPCEventListener(cons *Console, conn *grpc.ClientConn, service string, methodDesc *desc.MethodDescriptor) EventListener {
	stub := newRPCServerStreamStub(conn, methodDesc)

	p, err := os.FindProcess(syscall.Getpid())
	if err != nil {
		cons.Debugf("Could not get pid: %s.\n", err.Error())
	}

	return NewConsoleStubEventListener(cons, stub, service, p)
}

// NewConsoleStubEventListener creates a new ConsoleRPCEventListener for a given
// ServerStreamStub.
func NewConsoleStubEventListener(cons *Console, stub ServerStreamStub, service string, sig Signaler) EventListener {
	return &ConsoleRPCEventListener{
		cons:    cons,
		sig:     sig,
		service: service,
		stub:    stub,
	}
}

// Service returns the name of the streaming service.
func (el *ConsoleRPCEventListener) Service() string {
	return el.service
}

// Start connects to the corresponding event emitter and continuously
// listens for events and displays them.
// Start only allows a single connection to the RPC server.
// It will close the previous connection before starting a new one.
func (el *ConsoleRPCEventListener) Start(ctx context.Context) error {
	ss, err := el.connect(ctx)
	if err != nil {
		return err
	}

	defer func() {
		el.mu.Lock()
		el.connected = false
		el.close = nil
		el.mu.Unlock()
	}()

	for {
		res, err := ss.RecvMsg()
		if err == io.EOF || err == context.Canceled || err != nil && strings.Contains(err.Error(), "Canceled") {
			el.cons.Debugln("Event listener closed.")
			return nil
		}
		if err != nil {
			return err
		}

		ev, ok := res.(*pbevent.Event)
		if !ok {
			el.cons.Debugf("Could not parse event: %v.\n", res)
			continue
		}

		el.print(ev)
	}
}

// connect connects to the corresponding event emitter, closing a previous
// connection if there is already one.
func (el *ConsoleRPCEventListener) connect(ctx context.Context) (ServerStream, error) {
	el.mu.Lock()
	defer el.mu.Unlock()

	if el.connected && el.close != nil {
		el.cons.Debugf("Closing previous connection to %s.%s.\n", el.service, el.stub.Name())
		el.close()
	}

	listenCtx, cancel := context.WithCancel(ctx)
	el.close = cancel

	el.cons.Debugf("Connecting to event emitter: %s.%s.\n", el.service, el.stub.Name())
	ss, err := el.stub.Invoke(listenCtx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	el.connected = true

	return ss, nil
}

// print prints the event to the console.
func (el *ConsoleRPCEventListener) print(ev *pbevent.Event) {
	// The prefix is to get rid of the "Alice>" prompt at the
	// beginning of the received event.
	msg := fmt.Sprintf("\x0d%s", ev.Message)

	switch ev.Level {
	case pbevent.Level_DEBUG:
		el.cons.Debugln(msg)
	case pbevent.Level_INFO:
		el.cons.Infoln(msg)
	case pbevent.Level_WARNING:
		el.cons.Warningln(msg)
	case pbevent.Level_ERROR:
		el.cons.Errorln(msg)
	default:
		el.cons.Debugf("Unknown event level: %d.\n", ev.Level)
	}

	// This is necessary to re-render the prompt ("Alice>"").
	// See https://github.com/c-bata/go-prompt/issues/18.
	if err := el.sig.Signal(syscall.SIGWINCH); err != nil {
		el.cons.Debugln("Couldn't send signal to the OS to re-render. Ignoring.")
	}
}

// Connected returns true if the listener is connected to the event
// emitter and ready to receive events.
func (el *ConsoleRPCEventListener) Connected() bool {
	el.mu.RLock()
	defer el.mu.RUnlock()

	return el.connected
}
