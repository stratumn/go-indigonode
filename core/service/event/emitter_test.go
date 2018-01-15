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

package event

import (
	"context"
	"testing"
	"time"

	pb "github.com/stratumn/alice/grpc/event"
	"github.com/stretchr/testify/assert"
)

func TestEmitter(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	t.Run("Can add and remove listeners", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		assert.Equal(0, emitter.GetListenersCount(), "There should be no listener initially")

		l1 := emitter.AddListener(ctx)
		l2 := emitter.AddListener(ctx)
		assert.Equal(2, emitter.GetListenersCount(), "Listeners count")

		emitter.RemoveListener(ctx, l2)
		assert.Equal(1, emitter.GetListenersCount(), "Listeners count")

		emitter.RemoveListener(ctx, l1)
		assert.Equal(0, emitter.GetListenersCount(), "Listeners count")
	})

	t.Run("Close closes all channels", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)

		emitter.AddListener(ctx)
		emitter.AddListener(ctx)
		emitter.Close()

		assert.Equal(0, emitter.GetListenersCount(), "All listeners should be disconnected")
	})

	t.Run("Emit to multiple listeners", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		defer emitter.Close()

		l1 := emitter.AddListener(ctx)
		l2 := emitter.AddListener(ctx)

		e := &pb.Event{Message: "Hey", Level: pb.Level_INFO}
		emitter.Emit(e)

		assertReceived := func(l <-chan *pb.Event) {
			select {
			case ee := <-l:
				assert.Equal(e, ee, "Invalid event received")
			case <-time.After(10 * time.Millisecond):
				assert.Fail("Event not received in time window")
			}
		}

		assertReceived(l1)
		assertReceived(l2)
	})

	t.Run("Drop messages if listeners are too slow", func(t *testing.T) {
		emitter := NewEmitter(5 * time.Millisecond)
		defer emitter.Close()

		l := emitter.AddListener(ctx)

		e1 := &pb.Event{Message: "1", Level: pb.Level_INFO}
		e2 := &pb.Event{Message: "2", Level: pb.Level_INFO}
		emitter.Emit(e1)
		emitter.Emit(e2)

		// Receive first event.
		select {
		case <-l:
		case <-time.After(10 * time.Millisecond):
			assert.Fail("Event not received in time window")
		}

		// Sleep until second event is dropped.
		select {
		case <-time.After(10 * time.Millisecond):
			break
		}

		// The second event should have been dropped.
		select {
		case ee := <-l:
			assert.Nil(ee, "Unexpected event received")
		case <-time.After(5 * time.Millisecond):
			break
		}
	})

	t.Run("Doesn't remove listeners while events are pending", func(t *testing.T) {
		emitter := NewEmitter(10 * time.Millisecond)
		defer emitter.Close()

		l := emitter.AddListener(ctx)
		e1 := &pb.Event{Message: "1", Level: pb.Level_INFO}
		emitter.Emit(e1)

		removeChan := make(chan struct{})

		select {
		case removeChan <- func() struct{} {
			emitter.RemoveListener(ctx, l)
			return struct{}{}
		}():
			assert.Fail("RemoveListener should fail while an event is pending")
		case <-time.After(5 * time.Millisecond):
			break
		}

		select {
		case <-time.After(5 * time.Millisecond):
			break
		}

		// Event should have been dropped and listener removed.
		assert.Equal(0, emitter.GetListenersCount(), "Listener should have been removed")
	})
}
