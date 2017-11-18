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

A service must at least implement the Service interface. It may also implement
these additional interfaces:

	- Needy if it depends on other services
	- Pluggable if it needs to access the exposed type of its dependencies
	- Friendly if it can use but does not need some other services
	- Exposer if it exposes a type to other services
	- Runner if it runs a long running function
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
//
// Before any of these tasks can run the work queue must be started by calling
// Work().
type Manager struct {
	// processes is a map of processes for each registered service.
	processes map[string]*process

	// friends keeps tracks of services that are friendly with a service.
	friends map[string]map[string]Friendly

	// queue is used to queue tasks that need to run sequentially.
	queue chan func()
}

// New creates a new service manager.
//
// It also registers the manager service which can be used to control the
// manager itself.
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

	m.addFriends(id)
	m.addToFriends(serv)
}

// addFriends finds registered services that like the given service ID and adds
// them to its friend set.
func (m *Manager) addFriends(id string) {
	for servID, ps := range m.processes {
		s := ps.Service()

		if friendly, ok := s.(Friendly); ok {
			likes := friendly.Likes()
			if likes == nil {
				continue
			}

			if _, ok := likes[id]; ok {
				m.friends[id][servID] = friendly
			}
		}
	}
}

// addToFriends finds services are liked by the given service and adds it to
// their friend set.
func (m *Manager) addToFriends(serv Service) {
	if friendly, ok := serv.(Friendly); ok {
		likes := friendly.Likes()
		for servID := range likes {
			if _, ok := m.friends[servID]; ok {
				m.friends[servID][serv.ID()] = friendly
			}
		}
	}
}

// Work tells the manager to start and execute tasks in the queue.
//
// It blocks until the context is canceled.
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

// Start starts a registered service and all its dependencies in topological
// order.
//
// It blocks until all the services have started or after the first failure.
//
// After a successful start, the service status will be set to Running. If it
// failed to start a dependency, the status will remain Stopped, but an error
// will be returned. If all dependencies started, but the service failed to
// start, its status will be set to Errored, and an error will be returned.
//
// Starting a service makes it non-prunable, but the dependencies are prunable
// unless they were started with Start(). Calling Start() on an already running
// service will make it non-prunable.
func (m *Manager) Start(servID string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": servID,
	})
	event := log.EventBegin(ctx, "Start")
	defer event.Done()

	err := m.doErr(func() error {
		return m.doStart(ctx, servID)
	})
	if err != nil {
		event.SetError(err)
	}

	return err
}

func (m *Manager) doStart(ctx context.Context, servID string) error {
	event := log.EventBegin(ctx, "doStart")
	defer event.Done()

	deps, err := m.Deps(servID)
	if err != nil {
		event.SetError(err)
		return err
	}

	event.Append(logging.Metadata{"deps": deps})

	for _, depID := range deps {
		if err := m.startDepOf(ctx, servID, depID); err != nil {
			event.Append(logging.Metadata{
				"dependency": depID,
				"error":      err.Error(),
			})
			return err
		}
	}

	return nil
}

// startDepOf starts a dependency of a service if it isn't already running. It
// returns once the process is running or if it stopped unexpectedly before
// getting to the running state.
func (m *Manager) startDepOf(ctx context.Context, servID, depID string) error {
	ps := m.processes[depID]
	ps.WaitForStableState()
	status := ps.Status()

	switch status {
	case Running:
		// Make it unprunable if it is the service that was requested.
		ps.SetPrunable(ps.Prunable() && servID != depID)
		return nil
	case Stopped:
		// Keep going
	case Errored:
		// Keep going
	default:
		return errors.WithStack(ErrInvalidStatus)
	}

	ps.SetPrunable(servID != depID)

	return m.waitForStart(ctx, ps)
}

// waitForStart waits for a service to either start or exit with an error.
func (m *Manager) waitForStart(ctx context.Context, ps *process) error {
	done := make(chan error, 1)
	go func() {
		done <- m.exec(ctx, ps.Service().ID())
	}()

	select {
	case <-ps.Running():
		return nil
	case err := <-done:
		if err == nil {
			err = ErrInvalidStatus
		}
		return err
	}
}

// exec starts a service but not its dependencies.
//
// It blocks until the service exits.
func (m *Manager) exec(ctx context.Context, servID string) error {
	ctx = logging.ContextWithLoggable(ctx, logging.Metadata{
		"service": servID,
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ps := m.processes[servID]
	ps.SetStatus(Starting, nil)
	ps.SetCancel(cancel)
	serv := ps.Service()
	m.addRefs(serv)

	if err := m.plugNeeds(serv); err != nil {
		log.Event(ctx, "plugError", logging.Metadata{
			"error": err.Error(),
		})
	}

	running := make(chan struct{})
	stopping := make(chan struct{})
	done := make(chan error)

	go func() {
		done <- ps.Run(ctx, running, stopping)
	}()

	return m.observe(ctx, ps, running, stopping, done)
}

// observe handles the status changes of a process and updates the states
// accordingly.
func (m *Manager) observe(
	ctx context.Context,
	ps *process,
	running, stopping chan struct{},
	done chan error,
) error {
	serv := ps.Service()
	servID := serv.ID()

	for {
		select {
		case <-running:
			log.Event(ctx, "running", logging.Metadata{
				"dependency": servID,
			})

			ps.SetStatus(Running, nil)
			m.befriend(serv)

		case <-stopping:
			log.Event(ctx, "stopping", logging.Metadata{
				"dependency": servID,
			})

			ps.SetStatus(Stopping, nil)
			m.unfriend(servID)

		case err := <-done:
			if err != nil && errors.Cause(err) != context.Canceled {
				log.Event(ctx, "errored", logging.Metadata{
					"error":      err.Error(),
					"dependency": servID,
				})
				ps.SetStatus(Errored, err)
			} else {
				log.Event(ctx, "stopped", logging.Metadata{
					"dependency": servID,
				})
				ps.SetStatus(Stopped, nil)
			}

			m.removeRefs(serv)
			ps.SetPrunable(true)
			ps.ClearRefs()

			return err
		}
	}
}

// addRefs add a reference to the given service to all the services it needs.
func (m *Manager) addRefs(serv Service) {
	if needy, ok := serv.(Needy); ok {
		for depID := range needy.Needs() {
			m.processes[depID].AddRef(serv.ID())
		}
	}
}

// removeRefs removes the reference to the given service from all the services
// it needs.
func (m *Manager) removeRefs(serv Service) {
	if needy, ok := serv.(Needy); ok {
		for depID := range needy.Needs() {
			m.processes[depID].RemoveRef(serv.ID())
		}
	}
}

// plugNeeds finds all the services the given service needs and calls the plug
// method of the service.
func (m *Manager) plugNeeds(serv Service) error {
	if pluggable, ok := serv.(Pluggable); ok {
		outlets := map[string]interface{}{}
		for servID := range pluggable.Needs() {
			if exposer, ok := m.processes[servID].Service().(Exposer); ok {
				outlets[servID] = exposer.Expose()
			} else {
				outlets[servID] = nil
			}
		}

		if err := pluggable.Plug(outlets); err != nil {
			return err
		}
	}

	return nil
}

// befriend calls the befriend methods with the given service of all services
// that like the service.
func (m *Manager) befriend(serv Service) {
	servID := serv.ID()

	exposer, ok := serv.(Exposer)

	for _, friendlyID := range sortedFriendlyKeys(m.friends[servID]) {
		friendly := m.friends[servID][friendlyID]
		if ok {
			friendly.Befriend(servID, exposer.Expose())
		} else {
			friendly.Befriend(servID, nil)
		}
	}
}

// unfriend calls the befriend methods with nil for the given service ID of all
// services that like the service.
func (m *Manager) unfriend(servID string) {
	for _, friendlyID := range sortedFriendlyKeys(m.friends[servID]) {
		friendly := m.friends[servID][friendlyID]
		friendly.Befriend(servID, nil)
	}
}

// Stop stops a service. Stopping a service will fail if it is a dependency of
// other running services or it isn't running.
func (m *Manager) Stop(servID string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": servID,
	})
	event := log.EventBegin(ctx, "Stop")
	defer event.Done()

	return m.doErr(func() error {
		return m.doStop(ctx, servID)
	})
}

func (m *Manager) doStop(ctx context.Context, servID string) error {
	event := log.EventBegin(ctx, "doStop")
	defer event.Done()

	ps, ok := m.processes[servID]
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

		for _, servID := range sortedProcessKeys(m.processes) {
			ps := m.processes[servID]

			if !m.canBePruned(ps) {
				continue
			}

			stopEvent := log.EventBegin(ctx, "stopProcess", logging.Metadata{
				"service": servID,
			})

			ps.Cancel()
			<-ps.Stopped()
			stopEvent.Done()

			done = false
		}

		if done {
			return
		}
	}
}

// canBePruned returns whether a process can be pruned.
func (m *Manager) canBePruned(ps *process) bool {
	if !ps.Prunable() || ps.Status() != Running {
		return false
	}

	for range ps.Refs() {
		return false
	}

	return true
}

// StopAll stops all active services in an order which respects dependencies.
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

		for _, servID := range sortedProcessKeys(m.processes) {
			ps := m.processes[servID]

			if !m.canBeStopped(ps) {
				continue
			}

			stopEvent := log.EventBegin(ctx, "stopProcess", logging.Metadata{
				"service": servID,
			})

			ps.Cancel()
			<-ps.Stopped()
			stopEvent.Done()

			done = false
		}

		if done {
			return
		}
	}
}

// canBeStopped returns whether a process can be stopped.
//
// It also waits for the service to stop if it is stopping.
func (m *Manager) canBeStopped(ps *process) bool {
	status := ps.Status()

	if status == Stopped || status == Errored {
		return false
	}

	if status == Stopping {
		<-ps.Stopped()
		return false
	}

	for range ps.Refs() {
		return false
	}

	return true
}

// List returns all the registered service IDs.
func (m *Manager) List() []string {
	return sortedProcessKeys(m.processes)
}

// Find returns a service.
func (m *Manager) Find(servID string) (Service, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, servID)
	}

	serv := ps.Service()

	return serv, nil
}

// Proto returns a Protobuf struct for a service.
func (m *Manager) Proto(servID string) (*pb.Service, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, servID)
	}

	serv := ps.Service()

	msg := pb.Service{
		Id:        servID,
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
func (m *Manager) Status(servID string) (StatusCode, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return Stopped, errors.Wrap(ErrNotFound, servID)
	}

	return ps.Status(), nil
}

// Stoppable returns true if a service is stoppable (meaning it isn't a
// dependency of a running service.
func (m *Manager) Stoppable(servID string) (bool, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return false, errors.Wrap(ErrNotFound, servID)
	}

	return ps.Stoppable(), nil
}

// Prunable returns true if a service is prunable (meaning it wasn't started
// explicitly).
func (m *Manager) Prunable(servID string) (bool, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return false, errors.Wrap(ErrNotFound, servID)
	}

	return ps.Prunable(), nil
}

// Expose returns the object returned by the Expose method of the service.
//
// It returns nil if the service doesn't implement the Exposer interface or if
// it exposed nil.
func (m *Manager) Expose(servID string) (interface{}, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, servID)
	}

	if exposer, ok := ps.Service().(Exposer); ok {
		return exposer.Expose(), nil
	}

	return nil, nil
}

// Running returns a channel that will be notified once a service has started.
//
// If the services is already running, the channel receives immediately.
func (m *Manager) Running(servID string) (<-chan struct{}, error) {
	ps, ok := m.processes[servID]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, servID)
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
// The order of dependencies is deterministic.
func (m *Manager) Deps(servID string) ([]string, error) {
	return m.depSort(servID, &map[string]bool{})
}

// depSort does topological sorting using depth first search.
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

// Fgraph prints a unicode dependency graph of a service to a writer.
func (m *Manager) Fgraph(w io.Writer, servID string, prefix string) error {
	service, err := m.Find(servID)
	if err != nil {
		return err
	}

	fmt.Fprint(w, servID)

	if needy, ok := service.(Needy); ok && len(needy.Needs()) > 0 {
		deps := sortedSetKeys(needy.Needs())
		spaces := strings.Repeat(" ", len([]rune(servID)))

		return m.fgraphDeps(w, deps, prefix, spaces)
	}

	fmt.Fprintln(w)

	return nil
}

// fgraphDeps renders the dependencies of a service.
func (m *Manager) fgraphDeps(w io.Writer, deps []string, prefix, spaces string) error {
	l := len(deps)

	if err := m.fgraphFirstDep(w, deps[0], l < 2, prefix, spaces); err != nil {
		return err
	}

	if l > 2 {
		if err := m.fgraphMidDeps(w, deps, prefix, spaces); err != nil {
			return err
		}
	}

	if l > 1 {
		if err := m.fgraphLastDep(w, deps[l-1], prefix, spaces); err != nil {
			return err
		}
	}

	// Don't need to print extra newline.
	return nil
}

// fgraphFirstDep renders the first dependency of a service.
func (m *Manager) fgraphFirstDep(
	w io.Writer,
	depID string,
	unique bool,
	prefix, spaces string,
) error {
	if unique {
		fmt.Fprint(w, "─")
		return m.Fgraph(w, depID, prefix+spaces+" ")
	}

	fmt.Fprint(w, "┬")
	return m.Fgraph(w, depID, prefix+spaces+"│")
}

// fgraphMidDeps renders the middle dependencies of a serivce, meaning not the
// first nor the last one.
func (m *Manager) fgraphMidDeps(
	w io.Writer,
	deps []string,
	prefix, spaces string,
) error {
	for _, depID := range deps[1 : len(deps)-1] {
		fmt.Fprintln(w, prefix+spaces+"│")
		fmt.Fprint(w, prefix+spaces+"├")

		if err := m.Fgraph(w, depID, prefix+spaces+"│"); err != nil {
			return err
		}
	}

	return nil
}

// fgraphLastDep renders the last dependency of a service.
func (m *Manager) fgraphLastDep(
	w io.Writer,
	depID string,
	prefix, spaces string,
) error {
	fmt.Fprintln(w, prefix+spaces+"│")
	fmt.Fprint(w, prefix+spaces+"└")

	return m.Fgraph(w, depID, prefix+spaces+" ")
}
