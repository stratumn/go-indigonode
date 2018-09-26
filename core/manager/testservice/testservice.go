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

// Package testservice defines types to help test services.
package testservice

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/go-node/core/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CheckStrings checks that the strings returned by a service follow the
// convention.
func CheckStrings(t *testing.T, serv manager.Service) {
	id, name, desc := serv.ID(), serv.Name(), serv.Desc()

	assert.NotEqual(t, "", id, "serv.ID()")
	assert.NotEqual(t, "", name, "serv.Name()")
	assert.NotEqual(t, "", desc, "serv.Desc()")

	assert.Equal(t, strings.ToLower(id), id, "serv.ID() should be lowercase")
	assert.Equal(t, strings.Title(name), name, "serv.Name() should be a title with words beginning with an uppercase")

	if len(desc) > 0 {
		runes := []rune(desc)
		first, last := string(runes[0]), runes[len(runes)-1]

		assert.Equal(t, strings.ToUpper(first), first, "serv.Desc() should be a sentence that begins with an uppercase")
		assert.Equal(t, '.', last, "serv.Desc() should be a sentence that ends with a period")
	}
}

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
		require.Fail(t, "service did not expose anything in time")
	}

	cancel()

	select {
	case err := <-doneCh:
		require.NoError(t, err, "service errored")
	case <-time.After(timeout):
		require.Fail(t, "service did not exit in time")
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
		require.Fail(t, "service did not call running in time")
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
			require.Fail(t, "function did not complete in time")
		}
	}

	cancel()

	select {
	case <-stoppingCh:
	case <-time.After(timeout):
		require.Fail(t, "service did not call stopping in time")
	}

	select {
	case err := <-doneCh:
		require.NoError(t, err, "service errored")
	case <-time.After(timeout):
		require.Fail(t, "service did not exit in time")
	}
}
