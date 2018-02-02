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

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"

	pb "github.com/stratumn/alice/grpc/event"
)

var (
	// ErrNotSupportedTopic is returned when the topic cannot be parsed.
	ErrNotSupportedTopic = errors.New("topic cannot be parsed")
)

// Emitter emits events.
// Listeners can be added and removed.
type Emitter interface {
	AddListener(topic string) (<-chan *pb.Event, error)
	RemoveListener(<-chan *pb.Event)
	GetListenersCount(topic string) int
	Emit(*pb.Event)
	Close()
}

// DefaultTimeout is the recommended timeout before dropping messages.
var DefaultTimeout = 100 * time.Millisecond

// ServerEmitter is a simple Emitter.
type ServerEmitter struct {
	timeout time.Duration

	mu        sync.RWMutex
	listeners map[*listener]struct{}

	pending sync.WaitGroup
}

type listener struct {
	topic       glob.Glob
	receiveChan chan *pb.Event
}

// NewEmitter creates a new event emitter.
func NewEmitter(timeout time.Duration) Emitter {
	return &ServerEmitter{
		timeout:   timeout,
		listeners: make(map[*listener]struct{}),
	}
}

// AddListener adds an event listener.
// It returns the channel on which events will be pushed.
func (e *ServerEmitter) AddListener(topic string) (<-chan *pb.Event, error) {
	log.Event(context.Background(), "AddListener", logging.Metadata{
		"topic": topic,
	})

	g, err := glob.Compile(topic, '.')
	if err != nil {
		return nil, errors.WithMessage(errors.WithStack(err), topic)
	}

	receiveChan := make(chan *pb.Event)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners[&listener{g,
		receiveChan}] = struct{}{}

	return receiveChan, nil
}

// RemoveListener removes an event listener.
func (e *ServerEmitter) RemoveListener(listener <-chan *pb.Event) {
	log.Event(context.Background(), "RemoveListener")

	e.mu.Lock()
	defer e.mu.Unlock()

	// We need to wait for pending messages to be delivered or dropped.
	// If needed, we could be smarter and have a sync.WaitGroup per channel.
	e.pending.Wait()

	for l := range e.listeners {
		if l.receiveChan == listener {
			delete(e.listeners, l)
			break
		}
	}
}

// GetListenersCount returns the number of active listeners on a topic.
func (e *ServerEmitter) GetListenersCount(topic string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	count := 0

	for l := range e.listeners {
		if l.topic.Match(topic) {
			count++
		}
	}

	return count
}

// Emit emits an event to connected listeners.
func (e *ServerEmitter) Emit(ev *pb.Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for l := range e.listeners {
		if l.topic.Match(ev.Topic) {
			e.pending.Add(1)

			go func(l chan *pb.Event) {
				select {
				case l <- ev:
				case <-time.After(e.timeout):
					break
				}

				e.pending.Done()
			}(l.receiveChan)
		}
	}
}

// Close closes open channels and internal handles.
func (e *ServerEmitter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// We need to wait for pending messages to be delivered or dropped.
	e.pending.Wait()

	for l := range e.listeners {
		close(l.receiveChan)
	}

	e.listeners = nil
}
