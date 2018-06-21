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

package service

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

		e := &pb.Event{Message: "Hey", Level: pb.Level_INFO, Topic: "topic"}
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

			e := &pb.Event{Message: "Hey", Level: pb.Level_INFO, Topic: "topic.precise"}
			emitter.Emit(e)

			assertReceived := func(l <-chan *pb.Event, expect bool) {
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

		e1 := &pb.Event{Message: "1", Level: pb.Level_INFO, Topic: "topic"}
		e2 := &pb.Event{Message: "2", Level: pb.Level_INFO, Topic: "topic"}
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

		e1 := &pb.Event{Message: "1", Level: pb.Level_INFO, Topic: "topic"}
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
