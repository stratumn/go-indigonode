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

/*
Package manager deals with managing services.

The aim of this package is to provide visibility and control over
long-running goroutines. A goroutine is attached to a service. The manager can
start and stop services. It also deals with service dependencies and pruning
unused services.
*/
package manager

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/manager"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

var (
	// ErrNotFound is returned when a service is not found for a given
	// service ID.
	ErrNotFound = errors.New("service is not registered")

	// ErrInvalidStatus is returned when a service or a dependency has an
	// invalid status.
	ErrInvalidStatus = errors.New("service has invalid status")

	// ErrNeeded is returned when a service cannot be stopped because it
	// is required by other running services.
	ErrNeeded = errors.New("service is needed by other services")

	// ErrCyclic is returned when a service has cyclic dependencies.
	ErrCyclic = errors.New("service has cyclic dependencies")

	// ErrMissingServiceID is returned when a method requiring a service ID
	// was called without providing one.
	ErrMissingServiceID = errors.New("a service ID is required")
)

// log is the logger for the manager.
var log = logging.Logger("manager")

// Manager manages the lifecycle and dependencies of services.
//
// It has a queue that only allows one of these functions to run at the same
// time:
//
//	- Start()
// 	- Stop()
// 	- Prune()
//	- StopAll()
type Manager struct {
	// processes is a map of processes for each registered service.
	processes map[string]*process

	// friends keeps tracks of services that are friendly with a service.
	friends map[string]map[string]Friendly

	// queue is used to queue tasks that need to run sequentially.
	queue chan func()
}

// New creates a new service manager.
func New() *Manager {
	defer log.EventBegin(context.Background(), "New").Done()

	mgr := &Manager{
		processes: map[string]*process{},
		friends:   map[string]map[string]Friendly{},
		queue:     make(chan func()),
	}

	// Register the manager as a service.
	mgr.Register(managerService{mgr})

	return mgr
}

// Register adds a service to the set of known services.
func (m *Manager) Register(serv Service) {
	id := serv.ID()
	defer log.EventBegin(context.Background(), "register", logging.Metadata{
		"service": id,
	}).Done()

	m.processes[id] = newProcess(serv)
	m.friends[id] = map[string]Friendly{}

	// Find already registred friends.
	for sid, ps := range m.processes {
		s := ps.Service()

		if friendly, ok := s.(Friendly); ok {
			likes := friendly.Likes()
			if likes == nil {
				continue
			}

			if _, ok := likes[id]; ok {
				m.friends[id][sid] = friendly
			}
		}
	}

	// Find services it likes.
	if friendly, ok := serv.(Friendly); ok {
		likes := friendly.Likes()
		for sid := range likes {
			if _, ok := m.friends[sid]; ok {
				m.friends[sid][id] = friendly
			}
		}
	}
}

// Work tells the manager to start and execute tasks in the queue.
func (m *Manager) Work(ctx context.Context) error {
	log.Event(ctx, "Work")

	for {
		select {
		case task := <-m.queue:
			task()
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		}
	}
}

// do puts a task at the end of the queue and blocks until executed.
func (m *Manager) do(task func()) {
	done := make(chan struct{})
	m.queue <- func() {
		task()
		close(done)
	}
	<-done
}

// do puts a task that can return an error at the end of the queue and
// blocks until executed.
func (m *Manager) doErr(task func() error) error {
	done := make(chan error, 1)
	m.queue <- func() {
		done <- task()
	}
	return <-done
}

// Start starts a registered service and all its dependencies.
//
// It blocks until all the services have started or after the first failure.
//
// After a successful start, the service status will be set to ServiceRunning.
// If it failed to start a dependency, the status will remain ServiceStopped,
// but an error will be returned. If all dependencies started, but the service
// failed to start, its status will be set to ServiceErrored, and an error will
// be returned.
//
// Starting a service makes it non-prunable, but the dependencies are prunable
// unless they were started with Start(). Calling Start() on an already running
// service will make it non-prunable.
func (m *Manager) Start(serviceID string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": serviceID,
	})
	event := log.EventBegin(ctx, "Start")
	defer event.Done()

	err := m.doErr(func() error {
		return m.doStart(ctx, serviceID)
	})
	if err != nil {
		event.SetError(err)
	}

	return err
}

func (m *Manager) doStart(ctx context.Context, serviceID string) error {
	event := log.EventBegin(ctx, "doStart")
	defer event.Done()

	// Find all services required by the service.
	deps, err := m.Deps(serviceID)
	if err != nil {
		event.SetError(err)
		return err
	}

	event.Append(logging.Metadata{"deps": deps})

	// Start the process for the service and its dependencies.
	for _, sid := range deps {
		ps := m.processes[sid]
		status := ps.Status()

		// Deal with temporary states.
		switch status {
		case Stopping:
			// Ignore error, it will be logged. Start it again.
			log.Event(ctx, "prewaitStop", logging.Metadata{"dependency": sid})
			<-ps.Stopped()
		case Starting:
			log.Event(ctx, "prewaitStart", logging.Metadata{"dependency": sid})
			select {
			case <-ps.Running():
				ps.SetPrunable(ps.Prunable() && sid != serviceID)
				continue
			case err := <-ps.Stopped():
				if err != nil && errors.Cause(err) != context.Canceled {
					event.Append(logging.Metadata{"dependency": sid})
					event.SetError(err)
					return err
				}
			}
		}

		status = ps.Status()

		// Note: it is possible that the status changed if the process
		// stops, that case is currently not handled so it is not
		// perfectly concurrently safe. The worse case scenario is that
		// we think a process is running so we don't start it.

		switch status {
		case Running:
			// Make it unprunable if it is the service that was
			// requested.
			ps.SetPrunable(ps.Prunable() && sid != serviceID)
			continue
		case Stopped:
			// Keep going
		default:
			err := errors.WithStack(ErrInvalidStatus)
			event.Append(logging.Metadata{"dependency": sid})
			event.SetError(err)
			return err
		}

		ps.SetStatus(Prestarting, nil)
		ps.SetPrunable(sid != serviceID)
		ps.ClearRefs()

		// Start it.
		failed := make(chan error, 1)
		go func() {
			if err := m.exec(ctx, sid); err != nil && ps.Status() != Errored {
				failed <- err
			}
		}()

		// Wait for status change.
		select {
		case <-ps.Running():
			continue
		case err := <-failed:
			ps.SetStatus(Errored, err)
			event.Append(logging.Metadata{"dependency": sid})
			event.SetError(err)
			return err
		case err := <-ps.Stopped():
			ps.SetStatus(Errored, err)
			event.Append(logging.Metadata{"dependency": sid})
			event.SetError(err)
			return err
		}
	}

	return nil
}

// exec starts a service but not its dependencies.
func (m *Manager) exec(ctx context.Context, serviceID string) error {
	ctx = logging.ContextWithLoggable(ctx, logging.Metadata{
		"service": serviceID,
	})

	ps := m.processes[serviceID]
	ps.SetStatus(Starting, nil)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ps.SetCancel(cancel)

	serv := ps.Service()

	// Plug connected services.
	if pluggable, ok := serv.(Pluggable); ok {
		outlets := map[string]interface{}{}
		for sid := range pluggable.Needs() {
			if exposer, ok := m.processes[sid].Service().(Exposer); ok {
				outlets[sid] = exposer.Expose()
			} else {
				outlets[sid] = nil
			}
		}

		if err := pluggable.Plug(outlets); err != nil {
			log.Event(ctx, "plugError", logging.Metadata{
				"error": err.Error(),
			})
			return err
		}
	}

	// Handle refs.
	if needy, ok := serv.(Needy); ok {
		for sid := range needy.Needs() {
			m.processes[sid].AddRef(serviceID)
		}
	}

	running := make(chan struct{})
	stopping := make(chan struct{})
	done := make(chan error)

	go func() {
		done <- ps.Run(ctx, running, stopping)
	}()

	for {
		select {
		case <-running:
			log.Event(ctx, "running", logging.Metadata{
				"dependency": serviceID,
			})
			ps.SetStatus(Running, nil)

			// Befriend friendly services.
			exposer, ok := serv.(Exposer)
			for _, friendly := range m.friends[serviceID] {
				if ok {
					friendly.Befriend(serviceID, exposer.Expose())
				} else {
					friendly.Befriend(serviceID, nil)
				}
			}

		case <-stopping:
			log.Event(ctx, "stopping", logging.Metadata{
				"dependency": serviceID,
			})
			ps.SetStatus(Stopping, nil)

			// Unfriend friendly services.
			for _, friendly := range m.friends[serviceID] {
				friendly.Befriend(serviceID, nil)
			}

		case err := <-done:
			if err == nil || errors.Cause(err) == context.Canceled {
				log.Event(ctx, "stopped", logging.Metadata{
					"dependency": serviceID,
				})
			} else {
				log.Event(ctx, "errored", logging.Metadata{
					"error":      err.Error(),
					"dependency": serviceID,
				})
			}

			if needy, ok := serv.(Needy); ok {
				for sid := range needy.Needs() {
					m.processes[sid].RemoveRef(serviceID)
				}
			}
			if err == nil || errors.Cause(err) == context.Canceled {
				err = nil
				ps.SetStatus(Stopped, nil)
			} else {
				ps.SetStatus(Stopped, err)
			}
			return err
		}
	}

}

// Stop stops a service. Stopping a service will fail if it is a dependency of
// other running services.
func (m *Manager) Stop(serviceID string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": serviceID,
	})
	event := log.EventBegin(ctx, "Stop")
	defer event.Done()

	return m.doErr(func() error {
		return m.doStop(ctx, serviceID)
	})
}

func (m *Manager) doStop(ctx context.Context, serviceID string) error {
	event := log.EventBegin(ctx, "doStop")
	defer event.Done()

	ps, ok := m.processes[serviceID]
	if !ok {
		err := errors.WithStack(ErrNotFound)
		event.SetError(err)
		return err
	}

	for range ps.Refs() {
		err := errors.WithStack(ErrNeeded)
		event.SetError(err)
		return err
	}

	if status := ps.Status(); status != Running {
		err := errors.WithStack(ErrInvalidStatus)
		event.SetError(err)
		return err
	}

	ps.Cancel()

	return <-ps.Stopped()
}

// Prune stops all active services that are prunable and don't have services
// that depend on it.
func (m *Manager) Prune() {
	defer log.EventBegin(context.Background(), "Prune").Done()
	m.do(m.doPrune)
}

func (m *Manager) doPrune() {
	ctx := context.Background()
	event := log.EventBegin(ctx, "doPrune")
	defer event.Done()

	// Loop until there are no more prunable services.
	for {
		done := true

		// Find prunable processes in deterministic order.
	SERVICE_PRUNE_SERVICE_LOOP:
		for _, sid := range sortedProcessKeys(m.processes) {
			ps := m.processes[sid]

			if !ps.Prunable() || ps.Status() != Running {
				continue
			}

			for range ps.Refs() {
				continue SERVICE_PRUNE_SERVICE_LOOP
			}

			stopEvent := log.EventBegin(ctx, "stopProcess", logging.Metadata{
				"service": sid,
			})
			ps.Cancel()
			<-ps.Stopped()
			stopEvent.Done()

			done = false
		}

		if done {
			break
		}
	}
}

// StopAll stops all active processes in an order which respects dependencies.
func (m *Manager) StopAll() {
	defer log.EventBegin(context.Background(), "StopAll").Done()
	m.do(m.doStopAll)
}

func (m *Manager) doStopAll() {
	ctx := context.Background()
	event := log.EventBegin(ctx, "doStopAll")
	defer event.Done()

	// Loop until there are no more running processes.
	for {
		done := true

		// Find active services with no active dependencies in
		// deterministic order.
	SERVICE_STOP_ALL_SERVICE_LOOP:
		for _, sid := range sortedProcessKeys(m.processes) {
			ps := m.processes[sid]
			status := ps.Status()

			if status == Stopped || status == Errored {
				continue
			}

			if status == Stopping {
				<-ps.Stopped()
				continue
			}

			done = false

			for range ps.Refs() {
				continue SERVICE_STOP_ALL_SERVICE_LOOP
			}

			stopEvent := log.EventBegin(ctx, "stopProcess", logging.Metadata{
				"service": sid,
			})
			ps.Cancel()
			<-ps.Stopped()
			stopEvent.Done()

		}

		if done {
			break
		}
	}
}

// List returns all the registered service IDs.
func (m *Manager) List() []string {
	return sortedProcessKeys(m.processes)
}

// Find returns a service.
func (m *Manager) Find(serviceID string) (Service, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, serviceID)
	}

	serv := ps.Service()

	return serv, nil
}

// Proto returns a Protobuf struct for a service.
func (m *Manager) Proto(serviceID string) (*pb.Service, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, serviceID)
	}

	serv := ps.Service()

	msg := pb.Service{
		Id:        serviceID,
		Status:    pb.Service_Status(ps.Status()),
		Stoppable: ps.Stoppable(),
		Prunable:  ps.Prunable(),
		Name:      serv.Name(),
		Desc:      serv.Desc(),
	}

	if needy, ok := serv.(Needy); ok {
		msg.Needs = sortedSetKeys(needy.Needs())
	}

	return &msg, nil
}

// Status returns the status of a service.
func (m *Manager) Status(serviceID string) (StatusCode, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return Stopped, errors.Wrap(ErrNotFound, serviceID)
	}

	return ps.Status(), nil
}

// Stoppable returns true if a service is stoppable (meaning it has no
// refs).
func (m *Manager) Stoppable(serviceID string) (bool, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return false, errors.Wrap(ErrNotFound, serviceID)
	}

	return ps.Stoppable(), nil
}

// Prunable returns true if a service is prunable.
func (m *Manager) Prunable(serviceID string) (bool, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return false, errors.Wrap(ErrNotFound, serviceID)
	}

	return ps.Prunable(), nil
}

// Expose returns the object returned by the expose method of the service.
//
// It returns nil if the service doesn't implement the Exposer interface or if
// it exposed nil.
func (m *Manager) Expose(serviceID string) (interface{}, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, serviceID)
	}

	if exposer, ok := ps.Service().(Exposer); ok {
		return exposer.Expose(), nil
	}

	return nil, nil
}

// Running returns a channel that will be notified once a service has started.
//
// If the services is already running, the channel fires immediately.
func (m *Manager) Running(serviceID string) (chan struct{}, error) {
	ps, ok := m.processes[serviceID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, serviceID)
	}

	return ps.Running(), nil
}

// Deps finds and sorts in topological order the services required by the
// given service (meaning an order in which all the dependencies can be
// started.
//
// The returned slice contains the given service ID (in last position).
//
// It returns an error if the dependency graph is cyclic (meaning there isn't
// an order in which to properly start dependencies).
//
// The dependencies are computed deterministically.
func (m *Manager) Deps(serviceID string) ([]string, error) {
	return m.depSort(serviceID, &map[string]bool{})
}

// depSort does topological sorting using DFS.
func (m *Manager) depSort(id string, marks *map[string]bool) ([]string, error) {
	var deps []string

	// Check if we've already been here.
	if mark, ok := (*marks)[id]; ok {
		if !mark {
			// We have a temporary mark, so there is a cycle.
			return nil, errors.Wrap(ErrCyclic, id)
		}

		// We have a permanent mark, so move on to next needed service.
		return nil, nil
	}

	ps, ok := m.processes[id]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, id)
	}

	// Mark current service as temporary.
	(*marks)[id] = false

	// Visit required services in deterministic order.
	if needy, ok := ps.Service().(Needy); ok {
		for _, needed := range sortedSetKeys(needy.Needs()) {
			subdeps, err := m.depSort(needed, marks)
			if err != nil {
				return nil, err
			}

			// Add subdependencies.
			deps = append(deps, subdeps...)
		}
	}

	// Mark current service as permanent.
	(*marks)[id] = true

	// Add this service as a dependency.
	deps = append(deps, id)

	return deps, nil
}

// Fgraph prints the dependency graph of a service to a writer.
func (m *Manager) Fgraph(w io.Writer, serviceID string, prefix string) error {
	service, err := m.Find(serviceID)
	if err != nil {
		return err
	}

	fmt.Fprint(w, service.ID())

	if needy, ok := service.(Needy); ok && len(needy.Needs()) > 0 {
		needs := sortedSetKeys(needy.Needs())
		l := len(needs)
		spaces := strings.Repeat(" ", len([]rune(service.ID())))

		if l > 1 {
			fmt.Fprint(w, "┬")
			err = m.Fgraph(w, needs[0], prefix+spaces+"│")
			if err != nil {
				return err
			}
		} else {
			fmt.Fprint(w, "─")
			err = m.Fgraph(w, needs[0], prefix+spaces+" ")
			if err != nil {
				return err
			}
		}

		if l > 2 {
			for _, depID := range needs[1 : l-1] {
				fmt.Fprintln(w, prefix+spaces+"│")
				fmt.Fprint(w, prefix+spaces+"├")
				err = m.Fgraph(w, depID, prefix+spaces+"│")
				if err != nil {
					return err
				}
			}
		}

		if l > 1 {
			last := needs[l-1]
			fmt.Fprintln(w, prefix+spaces+"│")
			fmt.Fprint(w, prefix+spaces+"└")
			err = m.Fgraph(w, last, prefix+spaces+" ")
			if err != nil {
				return err
			}
		}

		// Don't need to print extra newline.
		return nil
	}

	fmt.Fprintln(w)

	return nil
}
