// Copyright © 2017  Stratumn SAS
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

package manager

import (
	"context"

	"github.com/pkg/errors"
)

// Queue executes functions sequentially.
type Queue struct {
	// ch is the normal queue.
	ch chan func()

	// chHi is the high priority queue.
	hiCh chan func()

	// cancel cancels the work context.
	cancel func()
}

// NewQueue creates a new queue.
func NewQueue() *Queue {
	return &Queue{
		ch:   make(chan func()),
		hiCh: make(chan func()),
	}
}

// Work tells the queue to start executing tasks.
//
// It blocks until the context is canceled.
func (q *Queue) Work(ctx context.Context) error {
	ctx, q.cancel = context.WithCancel(ctx)
	defer q.cancel()

	log.Event(ctx, "Queue.Work")

	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		case task := <-q.hiCh:
			task()
			continue
		default:
		}

		select {
		case task := <-q.hiCh:
			task()
			continue
		case task := <-q.ch:
			task()
		}

	}
}

// Stop stops the queue.
func (q *Queue) Stop() {
	if q.cancel != nil {
		q.cancel()
	}
}

// Do puts a task at the end of the queue and blocks until executed.
func (q *Queue) Do(task func()) {
	done := make(chan struct{})
	q.ch <- func() {
		task()
		close(done)
	}
	<-done
}

// DoError puts a task that can return an error at the end of the queue and
// blocks until executed.
func (q *Queue) DoError(task func() error) error {
	done := make(chan error, 1)
	q.ch <- func() {
		done <- task()
	}
	return <-done
}

// DoHi puts a task at the end of the hi-priority queue and blocks until
// executed.
func (q *Queue) DoHi(task func()) {
	done := make(chan struct{})
	q.hiCh <- func() {
		task()
		close(done)
	}
	<-done
}

// DoErrorHi puts a task that can return an error at the end of the hi-priority
// queue and blocks until executed.
func (q *Queue) DoErrorHi(task func() error) error {
	done := make(chan error, 1)
	q.hiCh <- func() {
		done <- task()
	}
	return <-done
}
