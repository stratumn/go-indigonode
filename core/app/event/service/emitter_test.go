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
	"testing"
	"time"

	"github.com/stratumn/go-indigonode/core/app/event/grpc"
	"github.com/stretchr/testify/assert"
)

func TestEmitter(t *testing.T) {
	assert := assert.New(t)

	t.Run("Can add and remove listeners", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		assert.Equal(0, emitter.GetListenersCount("topic"), "emitter.GetListenersCount()")

		l1, err := emitter.AddListener("topic")
		assert.NoError(err, "emitter.AddListener(topic)")

		l2, err := emitter.AddListener("topic")
		assert.NoError(err, "emitter.AddListener(topic)")

		assert.Equal(2, emitter.GetListenersCount("topic"), "emitter.GetListenersCount()")

		emitter.RemoveListener(l2)
		assert.Equal(1, emitter.GetListenersCount("topic"), "emitter.GetListenersCount()")

		emitter.RemoveListener(l1)
		assert.Equal(0, emitter.GetListenersCount("topic"), "emitter.GetListenersCount()")
	})

	t.Run("Close closes all channels", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)

		emitter.AddListener("topic")
		emitter.AddListener("topic")
		emitter.Close()

		assert.Equal(0, emitter.GetListenersCount("topic"), "emitter.GetListenersCount()")
	})

	t.Run("Emit to multiple listeners", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		defer emitter.Close()

		l1, err := emitter.AddListener("topic")
		assert.NoError(err, "emitter.AddListener(topic)")

		l2, err := emitter.AddListener("topic")
		assert.NoError(err, "emitter.AddListener(topic)")

		e := &grpc.Event{Message: "Hey", Level: grpc.Level_INFO, Topic: "topic"}
		emitter.Emit(e)

		assertReceived := func(l <-chan *grpc.Event) {
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

	t.Run("Emit to listeners with the correct topic", func(t *testing.T) {
		emitter := NewEmitter(DefaultTimeout)
		defer emitter.Close()

		type topicTestCase struct {
			name        string
			topic       string
			shouldMatch bool
		}

		testCases := []topicTestCase{
			{
				"With the precise topic",
				"topic.precise",
				true,
			}, {
				"With a matching topic",
				"topic.*",
				true,
			}, {
				"With a non match topic",
				"other.*",
				false,
			}, {
				"With a different topic",
				"other.different",
				false,
			}, {
				"With a global topic",
				"**",
				true,
			},
		}

		for _, t := range testCases {
			l, err := emitter.AddListener(t.topic)
			assert.NoError(err, "emitter.AddListener(%s)", t.topic)

			e := &grpc.Event{Message: "Hey", Level: grpc.Level_INFO, Topic: "topic.precise"}
			emitter.Emit(e)

			assertReceived := func(l <-chan *grpc.Event, expect bool) {
				select {
				case ee := <-l:
					if !expect {
						assert.Fail("emitter.Emit() should not have received event", t.name)
					}
					assert.Equal(e, ee, "emitter.Emit()")
				case <-time.After(10 * time.Millisecond):
					if expect {
						assert.Fail("emitter.Emit() did not send event", t.name)
					}
				}
			}

			assertReceived(l, t.shouldMatch)
		}

	})

	t.Run("Drop messages if listeners are too slow", func(t *testing.T) {
		emitter := NewEmitter(5 * time.Millisecond)
		defer emitter.Close()

		l, err := emitter.AddListener("topic")
		assert.NoError(err, "emitter.AddListener(topic)")

		e1 := &grpc.Event{Message: "1", Level: grpc.Level_INFO, Topic: "topic"}
		e2 := &grpc.Event{Message: "2", Level: grpc.Level_INFO, Topic: "topic"}
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

		l, err := emitter.AddListener("topic")
		assert.NoError(err, "emitter.AddListener(topic)")

		e1 := &grpc.Event{Message: "1", Level: grpc.Level_INFO, Topic: "topic"}
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
		assert.Equal(0, emitter.GetListenersCount("topic"), "emitter.GetListenersCount()")
	})
}
