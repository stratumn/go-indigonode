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
	"github.com/pkg/errors"
	pbevent "github.com/stratumn/go-node/core/app/event/grpc"

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

// ConsoleRPCEventListener implements the EventListener interface.
// It connects to the RPC server's Listen endpoint and prints events to
// the console.
type ConsoleRPCEventListener struct {
	cons *Console
	sig  Signaler

	client pbevent.EmitterClient

	mu        sync.RWMutex
	connected bool
	close     context.CancelFunc
}

// NewConsoleRPCEventListener creates a new ConsoleRPCEventListener
// connected to a given console and RPC Listen endpoint.
func NewConsoleRPCEventListener(cons *Console, conn *grpc.ClientConn) EventListener {
	client := pbevent.NewEmitterClient(conn)

	p, err := os.FindProcess(syscall.Getpid())
	if err != nil {
		cons.Debugf("Could not get pid: %s.\n", err.Error())
	}

	return NewConsoleClientEventListener(cons, client, p)
}

// NewConsoleClientEventListener creates a new ConsoleRPCEventListener for a given
// emitter client.
func NewConsoleClientEventListener(cons *Console, client pbevent.EmitterClient, sig Signaler) EventListener {
	return &ConsoleRPCEventListener{
		cons:   cons,
		sig:    sig,
		client: client,
	}
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
		ev := &pbevent.Event{}
		err := ss.RecvMsg(ev)
		if err == io.EOF || err == context.Canceled || err != nil && strings.Contains(err.Error(), "Canceled") {
			el.cons.Debugln("Event listener closed.")
			return nil
		}
		if err != nil {
			return err
		}

		el.print(ev)
	}
}

// connect connects to the corresponding event emitter, closing a previous
// connection if there is already one.
func (el *ConsoleRPCEventListener) connect(ctx context.Context) (pbevent.Emitter_ListenClient, error) {
	el.mu.Lock()
	defer el.mu.Unlock()

	if el.connected && el.close != nil {
		el.cons.Debugf("Closing previous connection to event service.\n")
		el.close()
	}

	listenCtx, cancel := context.WithCancel(ctx)
	el.close = cancel

	el.cons.Debugf("Connecting to event service.\n")
	lc, err := el.client.Listen(listenCtx, &pbevent.ListenReq{Topic: "**"})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	el.connected = true

	return lc, nil
}

// print prints the event to the console.
func (el *ConsoleRPCEventListener) print(ev *pbevent.Event) {
	// The prefix is to get rid of the "node>" prompt at the
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

	if err := renderPrompt(el.sig); err != nil {
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
