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
