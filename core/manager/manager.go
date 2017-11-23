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

This package is concurrently safe (TODO: formally verify that it is).
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

// state contains the state of a service.
type state struct {
	// Serv is the service attached to the state.
	Serv Service

	// Prunable is true if this process can be pruned. A service will be
	// pruned only if prunable is true and no active process depends on
	// this process.
	Prunable bool

	// Cancel tells the service to shut down.
	Cancel func()

	// Err is the error from when the service last stopped, if any.
	Err error

	// Status is the status of the service, such as Running.
	Status StatusCode

	// Refs is the set of currently active services that depend on this
	// service. It is used to decide if the service can be stopped.
	Refs map[string]struct{}

	// Friends keeps tracks of services that like this service (see the
	// Friendly interface).
	Friends map[string]Friendly

	// These keep track of channels that need to be notified when the
	// status changes.
	StartingChs []chan struct{}
	RunningChs  []chan struct{}
	StoppingChs []chan struct{}
	StoppedChs  []chan struct{}
	ErroredChs  []chan error

	// Queue is used to queue tasks that need to run sequentially at the
	// service level.
	Queue *Queue
}

// Manager manages the lifecycle and dependencies of services.
//
// You must call Work before any other function can return.
type Manager struct {
	// states keeps the states of every registered service.
	states map[string]*state

	// queue is used to queue tasks that need to run sequentially at the
	// manager level.
	queue *Queue
}

// New creates a new service manager.
func New() *Manager {
	defer log.EventBegin(context.Background(), "New").Done()

	mgr := &Manager{
		states: map[string]*state{},
		queue:  NewQueue(),
	}

	return mgr
}

// Work tells the manager to start and execute tasks in the queues.
//
// It blocks until the context is canceled.
func (m *Manager) Work(ctx context.Context) error {
	log.Event(ctx, "beginWork")
	defer log.Event(ctx, "endWork")

	queueCtx, cancelQueue := context.WithCancel(ctx)

	doneCh := make(chan error)
	go func() {
		doneCh <- m.queue.Work(queueCtx)
	}()

	<-ctx.Done()

	log.Event(ctx, "stoppingWork")

	m.stopStateQueues()
	cancelQueue()

	if err := <-doneCh; err != nil {
		return err
	}

	return errors.WithStack(ctx.Err())
}

// stopStateQueues stops all the state queues.
//
// It must be called before stopping the work context, otherwise it will block
// forever.
func (m *Manager) stopStateQueues() {
	m.queue.DoHi(func() {
		for _, servID := range sortedStateKeys(m.states) {
			m.states[servID].Queue.Stop()
		}
	})
}

// Register adds a service to the set of known services.
func (m *Manager) Register(serv Service) {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": serv.ID(),
	})

	defer log.EventBegin(ctx, "register").Done()

	m.queue.DoHi(func() {
		m.doRegister(ctx, serv)
	})
}

// doRegister must be executed in the manager queue.
//
// It launches the service state queue.
func (m *Manager) doRegister(ctx context.Context, serv Service) {
	defer log.EventBegin(ctx, "doRegister").Done()

	id := serv.ID()

	s := &state{
		Serv:    serv,
		Status:  Stopped,
		Refs:    map[string]struct{}{},
		Friends: map[string]Friendly{},
		Queue:   NewQueue(),
	}

	m.states[id] = s

	go func() {
		err := s.Queue.Work(ctx)
		if err != nil && errors.Cause(err) != context.Canceled {
			log.Event(ctx, "queueError", logging.Metadata{
				"error": err.Error(),
			})
		}
	}()

	m.addFriends(s)
	m.addToFriends(serv)
}

// RegisterService registers the manager itself as a service.
func (m *Manager) RegisterService() {
	m.Register(managerService{m})
}

// addFriends finds registered services that like the given service and adds
// them to its friend set.
//
// It must be executed in the manager queue.
func (m *Manager) addFriends(s *state) {
	servID := s.Serv.ID()

	for _, friendID := range sortedStateKeys(m.states) {
		friendState := m.states[friendID]

		if friendly, ok := friendState.Serv.(Friendly); ok {
			likes := friendly.Likes()
			if likes == nil {
				continue
			}

			if _, ok := likes[servID]; ok {
				s.Queue.DoHi(func() {
					s.Friends[friendID] = friendly
				})
			}
		}
	}
}

// addToFriends finds services that are liked by the given service and adds it
// to their friend set.
//
// It must be executed in the manager queue.
func (m *Manager) addToFriends(serv Service) {
	if friendly, ok := serv.(Friendly); ok {
		likes := friendly.Likes()
		for _, servID := range sortedSetKeys(likes) {
			if s, ok := m.states[servID]; ok {
				s.Queue.DoHi(func() {
					s.Friends[serv.ID()] = friendly
				})
			}
		}
	}
}

// safeState safely gets the state of a service.
func (m *Manager) safeState(servID string) (state *state, err error) {
	m.queue.DoHi(func() {
		var ok bool

		state, ok = m.states[servID]
		if !ok {
			err = errors.WithMessage(ErrNotFound, fmt.Sprintf(
				"could not find service %q",
				servID,
			))
		}
	})

	return
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
// (but not stoppable) unless they were started with Start(). Calling Start()
// on an already running service will make it non-prunable.
//
// It:
//
//	- safely obtains the dependencies of the service in topological order
//	- starts the dependencies and the service in order
//
func (m *Manager) Start(servID string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": servID,
	})

	event := log.EventBegin(ctx, "Start")
	defer event.Done()

	deps, err := m.Deps(servID)
	if err != nil {
		event.SetError(err)
		return err
	}

	event.Append(logging.Metadata{"deps": deps})

	err = m.startDeps(ctx, servID, deps)
	if err != nil {
		event.SetError(err)
	}

	return err
}

// startDeps starts all the dependencies of a service (including the service)
// in order.
//
// For each dependency it:
//
//	- safely obtains its state
//	- starts it in its state queue
//	- waits for it to either start or exit (in which case it fails)
//
func (m *Manager) startDeps(ctx context.Context, servID string, deps []string) error {
	for _, depID := range deps {
		s, err := m.safeState(depID)
		if err != nil {
			return err
		}

		err = s.Queue.DoError(func() error {
			return m.startDepOf(ctx, servID, depID)
		})

		if err != nil {
			return err
		}

		select {
		case <-m.safeDoRunning(s):

		case <-m.safeDoStopping(s):
			err = errors.WithMessage(ErrInvalidStatus, fmt.Sprintf(
				"service %q is stopping unexpectedly",
				s.Serv.ID(),
			))
			s.Queue.DoHi(func() {
				m.doSetErrored(s, err)
			})

		case <-m.safeDoStopped(s):
			err = errors.WithMessage(ErrInvalidStatus, fmt.Sprintf(
				"service %q quit unexpectedly",
				s.Serv.ID(),
			))
			s.Queue.DoHi(func() {
				m.doSetErrored(s, err)
			})

		case err = <-m.safeDoErrored(s):
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// startDepOf starts a dependency of a service if it isn't already running.
//
// It:
//
//	- makes sure the dependency is either Running, Stopped, or Errored
//	- updates the prunable state
//	- sets the status to Starting
//	- adds references to services it needs
//	- plugs the needed services
//	- starts a goroutine to run the service
//
// It must be executed in the queue of the dependency state queue.
func (m *Manager) startDepOf(ctx context.Context, servID, depID string) error {
	s := m.states[depID]

	switch s.Status {
	case Running:
		// Make it unprunable if it is the service that was requested.
		s.Prunable = s.Prunable && depID != servID
		return nil

	case Stopped:
		// Keep going

	case Errored:
		// Keep going

	default:
		return errors.WithMessage(ErrInvalidStatus, fmt.Sprintf(
			"cannot start service %q needed by %q because its status is %q",
			depID,
			servID,
			s.Status,
		))
	}

	s.Prunable = depID != servID

	err := m.queue.DoErrorHi(func() error {
		m.doSetStarting(s)
		m.addRefs(s)
		return m.plugNeeds(s)
	})
	if err != nil {
		m.doSetErrored(s, err)
		return err
	}

	go func() {
		err := m.exec(ctx, s)
		if err != nil && errors.Cause(err) != context.Canceled {
			log.Event(ctx, "execFailed", logging.Metadata{
				"error": err.Error(),
			})
		}
	}()

	return nil
}

// exec starts a service but not its dependencies.
//
// It:
//
//	- creates channels to observe the service
//	- launches the Run function in a goroutine
//	- starts observing the service
//
// It blocks until the service exits.
func (m *Manager) exec(ctx context.Context, s *state) error {
	ctx = logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": s.Serv.ID(),
	})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.Cancel = cancel

	runningCh := make(chan struct{})
	stoppingCh := make(chan struct{})
	doneCh := make(chan error)

	go func() {
		doneCh <- m.run(ctx, s, runningCh, stoppingCh)
	}()

	return m.observe(ctx, s, runningCh, stoppingCh, doneCh)
}

// run runs the Run function of a service if it has one or a default one.
func (m *Manager) run(ctx context.Context, s *state, runningCh, stoppingCh chan struct{}) error {
	running := func() {
		runningCh <- struct{}{}
	}

	stopping := func() {
		stoppingCh <- struct{}{}
	}

	if runner, ok := s.Serv.(Runner); ok {
		return runner.Run(ctx, running, stopping)
	}

	running()
	<-ctx.Done()
	stopping()

	return ctx.Err()
}

// observe observers the lifecycle of a service and updates the status
// accordingly.
func (m *Manager) observe(
	ctx context.Context,
	s *state,
	runningCh, stoppingCh chan struct{},
	doneCh chan error,
) error {
	servID := s.Serv.ID()

	for {
		select {
		case <-runningCh:
			log.Event(ctx, "running", logging.Metadata{
				"dependency": servID,
			})

			m.setRunning(s)

		case <-stoppingCh:
			log.Event(ctx, "stopping", logging.Metadata{
				"dependency": servID,
			})

			m.setStopping(s)

		case err := <-doneCh:
			if err != nil && errors.Cause(err) != context.Canceled {
				log.Event(ctx, "errored", logging.Metadata{
					"error": err.Error(),
				})
				m.setErrored(s, err)
			} else {
				log.Event(ctx, "stopped")
				m.setStopped(s)
			}

			return err
		}
	}
}

// doSetStarting must be executed in the state queue.
func (m *Manager) doSetStarting(s *state) {
	s.Status = Starting

	for _, ch := range s.StartingChs {
		close(ch)
	}

	s.StartingChs = nil
}

// setRunning sets the status of a service state to Running.
//
// It runs a high-priority task in the state queue.
func (m *Manager) setRunning(s *state) {
	s.Queue.DoHi(func() {
		m.doSetRunning(s)
	})
}

// doSetRunning must be executed in the state queue.
func (m *Manager) doSetRunning(s *state) {
	s.Status = Running
	m.befriend(s)

	for _, ch := range s.RunningChs {
		close(ch)
	}

	s.RunningChs = nil
}

// setStopping sets the status of a service state to Stopping.
//
// It runs a high-priority task in the state queue.
func (m *Manager) setStopping(s *state) {
	s.Queue.DoHi(func() {
		m.doSetStopping(s)
	})
}

// doSetStopping must be executed in the state queue.
func (m *Manager) doSetStopping(s *state) {
	s.Status = Stopping
	m.unfriend(s)

	for _, ch := range s.StoppingChs {
		close(ch)
	}

	s.StoppingChs = nil
}

// setStopped sets the status of a service state to Stopped.
func (m *Manager) setStopped(s *state) {
	m.queue.DoHi(func() {
		s.Queue.DoHi(func() {
			m.doSetStopped(s)
		})
	})
}

// doSetStopped must be executed in both the manager and state queues.
func (m *Manager) doSetStopped(s *state) {
	s.Status = Stopped
	s.Prunable = false
	s.Refs = map[string]struct{}{}
	m.removeRefs(s)
	s.Err = nil

	for _, ch := range s.StoppedChs {
		close(ch)
	}

	s.StoppedChs = nil
}

// setErrored sets the status of a service state to Errored.
func (m *Manager) setErrored(s *state, err error) {
	m.queue.DoHi(func() {
		s.Queue.DoHi(func() {
			m.doSetErrored(s, err)
		})
	})
}

// doSetErrored must be executed in both the manager and state queues.
func (m *Manager) doSetErrored(s *state, err error) {
	s.Status = Errored
	s.Prunable = true
	s.Refs = map[string]struct{}{}
	m.removeRefs(s)
	s.Err = err

	for _, ch := range s.ErroredChs {
		ch <- err
		close(ch)
	}

	s.ErroredChs = nil
}

// addRefs add a reference to the given service to all the services it needs.
//
// It must be executed in the manager queue.
func (m *Manager) addRefs(s *state) {
	if needy, ok := s.Serv.(Needy); ok {
		for _, depID := range sortedSetKeys(needy.Needs()) {
			dep := m.states[depID]
			dep.Queue.DoHi(func() {
				dep.Refs[s.Serv.ID()] = struct{}{}
			})
		}
	}
}

// removeRefs removes the reference to the given service from all the services
// it needs.
//
// It must be executed in the manager queue.
func (m *Manager) removeRefs(s *state) {
	if needy, ok := s.Serv.(Needy); ok {
		for _, depID := range sortedSetKeys(needy.Needs()) {
			dep := m.states[depID]
			dep.Queue.DoHi(func() {
				delete(dep.Refs, s.Serv.ID())
			})
		}
	}
}

// plugNeeds finds all the services the given service needs and calls the plug
// method of the service.
//
// It must be executed in the manager queue.
func (m *Manager) plugNeeds(s *state) error {
	if pluggable, ok := s.Serv.(Pluggable); ok {
		outlets := map[string]interface{}{}
		for _, depID := range sortedSetKeys(pluggable.Needs()) {
			dep := m.states[depID]
			if exposer, ok := dep.Serv.(Exposer); ok {
				outlets[depID] = exposer.Expose()
			} else {
				outlets[depID] = nil
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
//
// It must be executed in the state queue.
func (m *Manager) befriend(s *state) {
	exposer, ok := s.Serv.(Exposer)

	for _, friendlyID := range sortedFriendlyKeys(s.Friends) {
		friendly := s.Friends[friendlyID]
		if ok {
			friendly.Befriend(s.Serv.ID(), exposer.Expose())
		} else {
			friendly.Befriend(s.Serv.ID(), nil)
		}
	}
}

// unfriend calls the befriend methods with nil for the given service ID of all
// services that like the service.
//
// It must be executed in the state queue.
func (m *Manager) unfriend(s *state) {
	for _, friendlyID := range sortedFriendlyKeys(s.Friends) {
		friendly := s.Friends[friendlyID]
		friendly.Befriend(s.Serv.ID(), nil)
	}
}

// Stop stops a service. Stopping a service will fail if it is a dependency of
// other running services or it isn't running.
//
// It:
//
//	- safely obtains the state of the service
//	- stops the service
//	- waits for the service to stop or exit with an error
func (m *Manager) Stop(servID string) error {
	ctx := logging.ContextWithLoggable(context.Background(), logging.Metadata{
		"service": servID,
	})

	event := log.EventBegin(ctx, "Stop")
	defer event.Done()

	s, err := m.safeState(servID)
	if err != nil {
		event.SetError(err)
		return err
	}

	err = m.queue.DoError(func() error {
		return m.doStop(ctx, s)
	})
	if err != nil {
		event.SetError(err)
		return err
	}

	select {
	case <-m.safeDoStopped(s):
		return nil
	case err := <-m.safeDoErrored(s):
		return err
	}
}

// doStop must be executed in the manager queue.
//
// It:
//
//	- makes sure the service is not needed by other services
//	- checks that the service is Running
//	- cancels the Run function
func (m *Manager) doStop(ctx context.Context, s *state) error {
	event := log.EventBegin(ctx, "doStop", logging.Metadata{
		"service": s.Serv.ID(),
	})
	defer event.Done()

	for range s.Refs {
		err := errors.WithStack(ErrNeeded)
		event.SetError(err)
		return err
	}

	if status := s.Status; status != Running {
		err := errors.WithMessage(ErrInvalidStatus, fmt.Sprintf(
			"service %q is already running",
			s.Serv.ID(),
		))
		event.SetError(err)
		return err
	}

	go s.Cancel()

	return errors.WithStack(ctx.Err())
}

// Prune stops all active services that are prunable and don't have services
// that depend on it.
//
// It:
//
//	- finds a service that can be pruned
//	- stops the service
//	- waits for it to stop or exit with an error
//	- repeats until no service can be pruned
func (m *Manager) Prune() {
	ctx := context.Background()
	event := log.EventBegin(ctx, "Prune")
	defer event.Done()

	var servIDs []string
	m.queue.DoHi(func() {
		servIDs = sortedStateKeys(m.states)
	})

	for {
		done := true

		for _, servID := range servIDs {
			s, err := m.safeState(servID)
			if err != nil {
				log.Event(ctx, "safeStateError", logging.Metadata{
					"error": err.Error(),
				})
				continue
			}

			pruned := m.pruneIfPrunable(ctx, s)

			if pruned {
				select {
				case <-m.safeDoStopped(s):
				case <-m.safeDoErrored(s):
				}
			}

			done = done && !pruned
		}

		if done {
			return
		}
	}
}

// pruneIfPrunable prunes a service if it is prunable.
//
// It:
//
//	- makes sure the service is prunable and running
//	- makes sure it's not needed by another service
//	- stops the service
//	- returns whether it stopped a service
func (m *Manager) pruneIfPrunable(ctx context.Context, s *state) (pruned bool) {
	s.Queue.Do(func() {
		if !s.Prunable || s.Status != Running {
			return
		}

		for range s.Refs {
			return
		}

		err := m.doStop(ctx, s)
		if err != nil {
			log.Event(ctx, "pruneError", logging.Metadata{
				"error": err.Error(),
			})
			return
		}

		pruned = true
	})

	return
}

// StopAll stops all active services in an order which respects dependencies.
//
// It:
//
//	- finds a service that can be stopped
//	- stops the service
//	- waits for it to stop or exit with an error
//	- repeats until no service can be stopped
func (m *Manager) StopAll() {
	ctx := context.Background()
	event := log.EventBegin(ctx, "Prune")
	defer event.Done()

	var servIDs []string
	m.queue.DoHi(func() {
		servIDs = sortedStateKeys(m.states)
	})

	for {
		done := true

		for _, servID := range servIDs {
			s, err := m.safeState(servID)
			if err != nil {
				log.Event(ctx, "safeStateError", logging.Metadata{
					"error": err.Error(),
				})
				continue
			}

			stopped := m.stopIfStoppable(ctx, s)

			if stopped {
				select {
				case <-m.safeDoStopped(s):
				case <-m.safeDoErrored(s):
				}
			}

			done = done && !stopped
		}

		if done {
			return
		}
	}
}

// stopIfStoppable stops a service if it can be stopped.
//
// It:
//
//	- makes sure the service is not stopped or exited with an error
//	- returns true if the service is already stopping
//	- makes sure it's not needed by another service
//	- stops the service
//	- returns whether it stopped a service
func (m *Manager) stopIfStoppable(ctx context.Context, s *state) (stopped bool) {
	s.Queue.Do(func() {
		status := s.Status

		if status == Stopped || status == Errored {
			return
		}

		if status == Stopping {
			stopped = true
			return
		}

		for range s.Refs {
			return
		}

		err := m.doStop(ctx, s)
		if err != nil {
			log.Event(ctx, "stopError", logging.Metadata{
				"error": err.Error(),
			})
		}

		stopped = true
	})

	return
}

// List returns all the registered service IDs.
func (m *Manager) List() (servIDs []string) {
	m.queue.DoHi(func() {
		servIDs = sortedStateKeys(m.states)
	})

	return
}

// Find returns a service.
func (m *Manager) Find(servID string) (serv Service, err error) {
	m.queue.DoHi(func() {
		serv, err = m.doFind(servID)
	})

	return
}

// doFind must be executed in the manager queue.
func (m *Manager) doFind(servID string) (Service, error) {
	s, ok := m.states[servID]
	if !ok {
		return nil, errors.WithMessage(ErrNotFound, fmt.Sprintf(
			"could not find service %q",
			servID,
		))
	}

	return s.Serv, nil
}

// Proto returns a Protobuf struct for a service.
func (m *Manager) Proto(servID string) (*pb.Service, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return nil, err
	}

	var proto *pb.Service

	m.queue.DoHi(func() {
		proto, err = m.doProto(s)
	})

	return proto, err
}

// doProto must be executed in the manager queue.
func (m *Manager) doProto(s *state) (*pb.Service, error) {
	msg := pb.Service{
		Id:        s.Serv.ID(),
		Status:    pb.Service_Status(s.Status),
		Stoppable: m.doStoppable(s),
		Prunable:  s.Prunable,
		Name:      s.Serv.Name(),
		Desc:      s.Serv.Desc(),
	}

	if needy, ok := s.Serv.(Needy); ok {
		msg.Needs = sortedSetKeys(needy.Needs())
	}

	return &msg, nil
}

// Status returns the status of a service.
func (m *Manager) Status(servID string) (StatusCode, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return Stopped, err
	}

	return s.Status, nil
}

// Stoppable returns true if a service is stoppable (meaning it isn't a
// dependency of a running service.
func (m *Manager) Stoppable(servID string) (bool, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return false, err
	}

	stoppable := true

	s.Queue.DoHi(func() {
		for range s.Refs {
			stoppable = false
			return
		}
	})

	return stoppable, err
}

// doProto must be executed in the state queue.
func (m *Manager) doStoppable(s *state) bool {
	for range s.Refs {
		return false
	}

	return true
}

// Prunable returns true if a service is prunable (meaning it wasn't started
// explicitly).
func (m *Manager) Prunable(servID string) (bool, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return false, err
	}

	return s.Prunable, nil
}

// Expose returns the object returned by the Expose method of the service.
//
// It returns nil if the service doesn't implement the Exposer interface or if
// it exposed nil.
func (m *Manager) Expose(servID string) (interface{}, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return false, err
	}

	if exposer, ok := s.Serv.(Exposer); ok {
		return exposer.Expose(), nil
	}

	return nil, nil
}

// Starting returns a channel that will be closed once a service is starting.
//
// If the services is already starting, the channel closes immediately.
func (m *Manager) Starting(servID string) (<-chan struct{}, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return nil, err
	}

	return m.safeDoStarting(s), nil
}

func (m *Manager) safeDoStarting(s *state) <-chan struct{} {
	var ch <-chan struct{}
	s.Queue.DoHi(func() {
		ch = m.doStarting(s)
	})

	return ch
}

// doStarting must be executed in the state queue.
func (m *Manager) doStarting(s *state) <-chan struct{} {
	ch := make(chan struct{})

	if s.Status == Starting {
		close(ch)
	} else {
		s.StartingChs = append(s.StartingChs, ch)
	}

	return ch
}

// Running returns a channel that will be closed once a service is running.
//
// If the services is already running, the channel closes immediately.
func (m *Manager) Running(servID string) (<-chan struct{}, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return nil, err
	}

	return m.safeDoRunning(s), nil
}

func (m *Manager) safeDoRunning(s *state) <-chan struct{} {
	var ch <-chan struct{}
	s.Queue.DoHi(func() {
		ch = m.doRunning(s)
	})

	return ch
}

// doRunning must be executed in the state queue.
func (m *Manager) doRunning(s *state) <-chan struct{} {
	ch := make(chan struct{})

	if s.Status == Running {
		close(ch)
	} else {
		s.RunningChs = append(s.RunningChs, ch)
	}

	return ch
}

// Stopping returns a channel that will be closed once a service is stopping.
//
// If the services is already stopping, the channel closes immediately.
func (m *Manager) Stopping(servID string) (<-chan struct{}, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return nil, err
	}

	return m.safeDoStopping(s), nil
}

func (m *Manager) safeDoStopping(s *state) <-chan struct{} {
	var ch <-chan struct{}
	s.Queue.DoHi(func() {
		ch = m.doStopping(s)
	})

	return ch
}

// doStopping must be executed in the state queue.
func (m *Manager) doStopping(s *state) <-chan struct{} {
	ch := make(chan struct{})

	if s.Status == Stopping {
		close(ch)
	} else {
		s.StoppingChs = append(s.StoppingChs, ch)
	}

	return ch
}

// Stopped returns a channel that will be closed once a service stopped.
//
// If the services is already stopped, the channel receives immediately.
func (m *Manager) Stopped(servID string) (<-chan struct{}, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return nil, err
	}

	return m.safeDoStopped(s), nil
}

func (m *Manager) safeDoStopped(s *state) <-chan struct{} {
	var ch <-chan struct{}
	s.Queue.DoHi(func() {
		ch = m.doStopped(s)
	})

	return ch
}

// doStopped must be executed in the state queue.
func (m *Manager) doStopped(s *state) <-chan struct{} {
	ch := make(chan struct{})

	if s.Status == Stopped {
		close(ch)
	} else {
		s.StoppedChs = append(s.StoppedChs, ch)
	}

	return ch
}

// Errored returns a channel that will be notified of the error and closed once
// a service errored.
//
// If the services is already errored, the channel receives the error and
// closes immediately.
func (m *Manager) Errored(servID string) (<-chan error, error) {
	s, err := m.safeState(servID)
	if err != nil {
		return nil, err
	}

	return m.safeDoErrored(s), nil
}

func (m *Manager) safeDoErrored(s *state) <-chan error {
	var ch <-chan error
	s.Queue.DoHi(func() {
		ch = m.doErrored(s)
	})

	return ch
}

// doErrored must be executed in the state queue.
func (m *Manager) doErrored(s *state) <-chan error {
	ch := make(chan error, 1)

	if s.Status == Errored {
		ch <- s.Err
		close(ch)
	} else {
		s.ErroredChs = append(s.ErroredChs, ch)
	}

	return ch
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
func (m *Manager) Deps(servID string) (deps []string, err error) {
	m.queue.DoHi(func() {
		deps, err = m.doDeps(servID)
	})

	return
}

func (m *Manager) doDeps(servID string) ([]string, error) {
	return m.depSort(servID, &map[string]bool{})
}

// depSort does topological sorting using depth first search.
//
// It must be executed in the manager queue.
func (m *Manager) depSort(servID string, marks *map[string]bool) ([]string, error) {
	var deps []string

	// Check if we've already been here.
	if mark, ok := (*marks)[servID]; ok {
		if !mark {
			return nil, errors.WithMessage(ErrCyclic, fmt.Sprintf(
				"cyclic dependency found in %q",
				servID,
			))
		}

		// We have a permanent mark, so move on to next needed service.
		return nil, nil
	}

	s, ok := m.states[servID]
	if !ok {
		return nil, errors.WithMessage(ErrNotFound, fmt.Sprintf(
			"could not find service %q",
			servID,
		))
	}

	// Mark current service as temporary.
	(*marks)[servID] = false

	// Visit required services in deterministic order.
	if needy, ok := s.Serv.(Needy); ok {
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
	(*marks)[servID] = true

	// Add this service as a dependency.
	deps = append(deps, servID)

	return deps, nil
}

// Fgraph prints a unicode dependency graph of a service to a writer.
func (m *Manager) Fgraph(w io.Writer, servID string, prefix string) error {
	return m.queue.DoErrorHi(func() error {
		return m.doFgraph(w, servID, prefix)
	})
}

// doGraph must be executed in the manager queue.
func (m *Manager) doFgraph(w io.Writer, servID string, prefix string) error {
	serv, err := m.doFind(servID)
	if err != nil {
		return err
	}

	fmt.Fprint(w, servID)

	if needy, ok := serv.(Needy); ok && len(needy.Needs()) > 0 {
		deps := sortedSetKeys(needy.Needs())
		spaces := strings.Repeat(" ", len([]rune(servID)))

		return m.fgraphDeps(w, deps, prefix, spaces)
	}

	fmt.Fprintln(w)

	return nil
}

// fgraphDeps renders the dependencies of a service.
//
// It must be executed in the manager queue.
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
//
// It must be executed in the manager queue.
func (m *Manager) fgraphFirstDep(
	w io.Writer,
	depID string,
	unique bool,
	prefix, spaces string,
) error {
	if unique {
		fmt.Fprint(w, "─")
		return m.doFgraph(w, depID, prefix+spaces+" ")
	}

	fmt.Fprint(w, "┬")
	return m.doFgraph(w, depID, prefix+spaces+"│")
}

// fgraphMidDeps renders the middle dependencies of a serivce, meaning not the
// first nor the last one.
//
// It must be executed in the manager queue.
func (m *Manager) fgraphMidDeps(
	w io.Writer,
	deps []string,
	prefix, spaces string,
) error {
	for _, depID := range deps[1 : len(deps)-1] {
		fmt.Fprintln(w, prefix+spaces+"│")
		fmt.Fprint(w, prefix+spaces+"├")

		if err := m.doFgraph(w, depID, prefix+spaces+"│"); err != nil {
			return err
		}
	}

	return nil
}

// fgraphLastDep renders the last dependency of a service.
//
// It must be executed in the manager queue.
func (m *Manager) fgraphLastDep(
	w io.Writer,
	depID string,
	prefix, spaces string,
) error {
	fmt.Fprintln(w, prefix+spaces+"│")
	fmt.Fprint(w, prefix+spaces+"└")

	return m.doFgraph(w, depID, prefix+spaces+" ")
}
