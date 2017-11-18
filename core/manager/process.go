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
	"sync"

	"github.com/pkg/errors"
)

// process represents a running service managed by a service manager.
type process struct {
	// service is the service attached to the process.
	service Service

	// prunable is true this process can be pruned. A process will be
	// pruned only if prunable is true and no active process depends on
	// this process.
	prunable bool

	// Cancel tells the process to shut down.
	cancel func()

	// Err is the error from when the process last stopped, if any.
	err error

	// mu is used to avoid data races, but the current implementation is
	// not perfectly concurrently safe. It protects the remaining
	// variables.
	mu sync.RWMutex

	// status is the status of the process, such as Running.
	status StatusCode

	// refs is the set of currently active processes that depend on this
	// process. It is used to decide if the service can be stopped.
	refs map[string]struct{}

	// These keep track of channels that need to receive when the status
	// changes.
	running []chan struct{}
	stopped []chan error
}

// newProcess creates a new process for a service.
func newProcess(service Service) *process {
	return &process{
		service: service,
		refs:    map[string]struct{}{},
		status:  Stopped,
	}
}

// Service returns the service of the process.
func (ps *process) Service() Service {
	return ps.service
}

// Run calls the Run function of the service if it has one.
//
// If it doesn't have a Run function, it blocks until the context is canceled.
func (ps *process) Run(ctx context.Context, running chan struct{}, stopping chan struct{}) error {
	if runner, ok := ps.service.(Runner); ok {
		return runner.Run(ctx, running, stopping)
	}

	running <- struct{}{}
	<-ctx.Done()
	stopping <- struct{}{}

	return errors.WithStack(ctx.Err())
}

// Status returns the status of the process.
func (ps *process) Status() StatusCode {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.status
}

// SetStatus sets the status of the process.
//
// It will send to any channel that observes the new status.
func (ps *process) SetStatus(status StatusCode, err error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	oldStatus := ps.status
	ps.status = status

	if status == oldStatus {
		return
	}

	switch status {
	case Running:
		for _, ch := range ps.running {
			ch <- struct{}{}
		}
		ps.running = nil
	case Stopped:
		for _, ch := range ps.stopped {
			ch <- err
		}
		ps.stopped = nil
		ps.err = err
	}
}

// Prunable returns whether the process is prunable.
func (ps *process) Prunable() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.prunable
}

// SetPrunable sets whether the process is prunable.
func (ps *process) SetPrunable(prunable bool) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.prunable = prunable
}

// Stoppable returns whether the process can be stopped.
func (ps *process) Stoppable() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for range ps.refs {
		return false
	}

	return true
}

// Refs returns the services referencing this service.
func (ps *process) Refs() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return sortedSetKeys(ps.refs)
}

// AddRef adds a service referencing this process.
func (ps *process) AddRef(serviceID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.refs[serviceID] = struct{}{}
}

// RemoveRef removes a service referencing this process.
func (ps *process) RemoveRef(serviceID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.refs, serviceID)
}

// ClearRefs removes all services referencing this process.
func (ps *process) ClearRefs() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.refs = map[string]struct{}{}
}

// Cancel tells the process to stop.
func (ps *process) Cancel() {
	ps.cancel()
}

// SetCancel sets the cancel func.
func (ps *process) SetCancel(cancel func()) {
	ps.cancel = cancel
}

// Running returns a channel to be notified once the process is running.
func (ps *process) Running() <-chan struct{} {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan struct{}, 1)

	if ps.status == Running {
		ch <- struct{}{}
	} else {
		ps.running = append(ps.running, ch)
	}

	return ch
}

// Stopped returns a channel to be notified once the process is stopped.
func (ps *process) Stopped() <-chan error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan error, 1)

	if ps.status == Stopped {
		ch <- ps.err
	} else {
		ps.stopped = append(ps.stopped, ch)
	}

	return ch
}

// waitForStableState wait until the process is either running, stopped, or
// errored.
func (ps *process) WaitForStableState() {
	switch ps.status {
	case Starting:
		select {
		case <-ps.Running():
		case <-ps.Stopped():
		}
	case Stopping:
		<-ps.Stopped()
	}
}
