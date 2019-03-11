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
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/app/event/grpc"

	logging "github.com/ipfs/go-log"
)

var (
	// ErrUnupportedTopic is returned when the topic cannot be parsed.
	ErrUnupportedTopic = errors.New("topic cannot be parsed")
)

// Emitter emits events.
// Listeners can be added and removed.
type Emitter interface {
	AddListener(topic string) (<-chan *grpc.Event, error)
	RemoveListener(<-chan *grpc.Event)
	GetListenersCount(topic string) int
	Emit(*grpc.Event)
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
	receiveChan chan *grpc.Event
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
func (e *ServerEmitter) AddListener(topic string) (<-chan *grpc.Event, error) {
	log.Event(context.Background(), "AddListener", logging.Metadata{
		"topic": topic,
	})

	g, err := glob.Compile(topic, '.')
	if err != nil {
		return nil, errors.WithMessage(err, topic)
	}

	receiveChan := make(chan *grpc.Event)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners[&listener{
		topic:       g,
		receiveChan: receiveChan,
	}] = struct{}{}

	return receiveChan, nil
}

// RemoveListener removes an event listener.
func (e *ServerEmitter) RemoveListener(listener <-chan *grpc.Event) {
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
func (e *ServerEmitter) Emit(ev *grpc.Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for l := range e.listeners {
		if l.topic.Match(ev.Topic) {
			e.pending.Add(1)

			go func(l chan *grpc.Event) {
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
