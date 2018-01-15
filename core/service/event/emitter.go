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

//go:generate mockgen -package mockevent -destination mockevent/mockemitter.go github.com/stratumn/alice/core/service/event Emitter

package event

import (
	"context"
	"sync"
	"time"

	pb "github.com/stratumn/alice/grpc/event"
)

// Emitter emits events.
// Listeners can be added and removed.
type Emitter interface {
	AddListener(context.Context) <-chan *pb.Event
	RemoveListener(context.Context, <-chan *pb.Event)
	GetListenersCount() int
	Emit(*pb.Event)
	Close()
}

// DefaultTimeout is the recommended timeout before dropping messages.
var DefaultTimeout = 100 * time.Millisecond

// ServerEmitter is a simple Emitter.
type ServerEmitter struct {
	timeout time.Duration

	mu        sync.RWMutex
	pending   sync.WaitGroup
	listeners []chan *pb.Event
}

// NewEmitter creates a new event emitter.
func NewEmitter(timeout time.Duration) Emitter {
	return &ServerEmitter{
		timeout: timeout,
	}
}

// AddListener adds an event listener.
// It returns the channel on which events will be pushed.
func (e *ServerEmitter) AddListener(ctx context.Context) <-chan *pb.Event {
	log.Event(ctx, "AddListener")

	receiveChan := make(chan *pb.Event)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners = append(e.listeners, receiveChan)
	return receiveChan
}

// RemoveListener removes an event listener.
func (e *ServerEmitter) RemoveListener(ctx context.Context, listener <-chan *pb.Event) {
	log.Event(ctx, "RemoveListener")

	e.mu.Lock()
	defer e.mu.Unlock()

	// We need to wait for pending messages to be delivered or dropped.
	// If needed, we could be smarter and have a sync.WaitGroup per channel.
	e.pending.Wait()

	for i, l := range e.listeners {
		if l == listener {
			e.listeners[i] = e.listeners[len(e.listeners)-1]
			e.listeners = e.listeners[:len(e.listeners)-1]
			break
		}
	}
}

// GetListenersCount returns the number of active listeners.
func (e *ServerEmitter) GetListenersCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.listeners)
}

// Emit emits an event to connected listeners.
func (e *ServerEmitter) Emit(ev *pb.Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, listener := range e.listeners {
		e.pending.Add(1)

		go func(l chan *pb.Event) {
			select {
			case l <- ev:
			case <-time.After(e.timeout):
				break
			}

			e.pending.Done()
		}(listener)
	}
}

// Close closes open channels and internal handles.
func (e *ServerEmitter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// We need to wait for pending messages to be delivered or dropped.
	e.pending.Wait()

	for _, listener := range e.listeners {
		close(listener)
	}

	e.listeners = nil
}
