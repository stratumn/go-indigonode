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

// Package testservice defines types to help test services.
package testservice

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/manager"
)

// Expose call the run function of the service and returns the exposed object
// after the service is running.
func Expose(ctx context.Context, t *testing.T, serv manager.Exposer, timeout time.Duration) interface{} {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	exposedCh := make(chan interface{}, 1)
	doneCh := make(chan error, 1)

	if runner, ok := serv.(manager.Runner); ok {
		go func() {
			err := runner.Run(ctx, func() {
				exposedCh <- serv.Expose()
			}, func() {})

			if err != nil && errors.Cause(err) != context.Canceled {
				doneCh <- err
				return
			}

			doneCh <- nil
		}()
	} else {
		exposedCh <- serv.Expose()
		doneCh <- nil
	}

	var exposed interface{}

	select {
	case exposed = <-exposedCh:
	case <-time.After(timeout):
		t.Fatal("service did not expose anything in time")
	}

	cancel()

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("service errored: %s", err)

		}
	case <-time.After(timeout):
		t.Fatal("service did not exit in time")
	}

	return exposed
}

// TestRun makes sure the Run function of a service calls the running and
// stopping functions properly, and that it doesn't return an error.
func TestRun(ctx context.Context, t *testing.T, serv manager.Runner, timeout time.Duration) {
	TestRunning(ctx, t, serv, timeout, nil)
}

// TestRunning runs a function while the service is running.
func TestRunning(ctx context.Context, t *testing.T, serv manager.Runner, timeout time.Duration, fn func()) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	runningCh := make(chan struct{}, 1)
	stoppingCh := make(chan struct{}, 1)
	doneCh := make(chan error, 1)

	go func() {
		err := serv.Run(ctx, func() {
			close(runningCh)
		}, func() {
			close(stoppingCh)
		})

		if err != nil && errors.Cause(err) != context.Canceled {
			doneCh <- err
			return
		}

		doneCh <- nil
	}()

	select {
	case <-runningCh:
	case <-time.After(timeout):
		t.Fatal("service did not call running in time")
	}

	if fn != nil {
		fnCh := make(chan struct{}, 1)
		go func() {
			fn()
			close(fnCh)
		}()

		select {
		case <-fnCh:
		case <-time.After(timeout):
			t.Fatal("function did not complete in time")
		}
	}

	cancel()

	select {
	case <-stoppingCh:
	case <-time.After(timeout):
		t.Fatal("service did not call stopping in time")
	}

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("service errored: %s", err)

		}
	case <-time.After(timeout):
		t.Fatal("service did not exit in time")
	}
}
