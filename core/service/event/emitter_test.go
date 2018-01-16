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
	"testing"
	"time"

	pb "github.com/stratumn/alice/grpc/event"
	"github.com/stretchr/testify/assert"
)

func TestEmitter(t *testing.T) {
	assert := assert.New(t)

	t.Run("Can add and remove listeners", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		assert.Equal(0, emitter.GetListenersCount(), "emitter.GetListenersCount()")

		l1 := emitter.AddListener()
		l2 := emitter.AddListener()
		assert.Equal(2, emitter.GetListenersCount(), "emitter.GetListenersCount()")

		emitter.RemoveListener(l2)
		assert.Equal(1, emitter.GetListenersCount(), "emitter.GetListenersCount()")

		emitter.RemoveListener(l1)
		assert.Equal(0, emitter.GetListenersCount(), "emitter.GetListenersCount()")
	})

	t.Run("Close closes all channels", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)

		emitter.AddListener()
		emitter.AddListener()
		emitter.Close()

		assert.Equal(0, emitter.GetListenersCount(), "emitter.GetListenersCount()")
	})

	t.Run("Emit to multiple listeners", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		defer emitter.Close()

		l1 := emitter.AddListener()
		l2 := emitter.AddListener()

		e := &pb.Event{Message: "Hey", Level: pb.Level_INFO}
		emitter.Emit(e)

		assertReceived := func(l <-chan *pb.Event) {
			select {
			case ee := <-l:
				assert.Equal(e, ee, "emitter.Emit()")
			case <-time.After(10 * time.Millisecond):
				assert.Fail("emitter.Emit() did not send event")
			}
		}

		assertReceived(l1)
		assertReceived(l2)
	})

	t.Run("Drop messages if listeners are too slow", func(t *testing.T) {
		emitter := NewEmitter(5 * time.Millisecond)
		defer emitter.Close()

		l := emitter.AddListener()

		e1 := &pb.Event{Message: "1", Level: pb.Level_INFO}
		e2 := &pb.Event{Message: "2", Level: pb.Level_INFO}
		emitter.Emit(e1)
		emitter.Emit(e2)

		// Receive first event.
		select {
		case <-l:
		case <-time.After(10 * time.Millisecond):
			assert.Fail("emitter.Emit() did not send event")
		}

		// Sleep until second event is dropped.
		<-time.After(10 * time.Millisecond)

		// The second event should have been dropped.
		select {
		case ee := <-l:
			assert.Nil(ee, "emitter.Emit()")
		case <-time.After(5 * time.Millisecond):
			break
		}
	})

	t.Run("Doesn't remove listeners while events are pending", func(t *testing.T) {
		emitter := NewEmitter(10 * time.Millisecond)
		defer emitter.Close()

		l := emitter.AddListener()
		e1 := &pb.Event{Message: "1", Level: pb.Level_INFO}
		emitter.Emit(e1)

		removeChan := make(chan struct{})

		select {
		case removeChan <- func() struct{} {
			emitter.RemoveListener(l)
			return struct{}{}
		}():
			assert.Fail("emitter.RemoveListener()")
		case <-time.After(5 * time.Millisecond):
			break
		}

		<-time.After(5 * time.Millisecond)

		// Event should have been dropped and listener removed.
		assert.Equal(0, emitter.GetListenersCount(), "emitter.GetListenersCount()")
	})
}
