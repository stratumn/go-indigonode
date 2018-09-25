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
	pb "github.com/stratumn/go-node/core/manager/grpc"
	"google.golang.org/grpc"
)

// Service describes a Stratumn Node service.
type Service interface {
	// ID returns a unique identifier.
	ID() string

	// Name returns a user friendly name.
	Name() string

	// Desc returns a short description of what the service does.
	Desc() string
}

// Needy depends on other services.
type Needy interface {
	// Needs returns a set of service identifiers needed before this
	// service can start.
	Needs() map[string]struct{}
}

// Pluggable connects other services.
type Pluggable interface {
	Needy

	// Plug is given a map of exposed connected objects, giving the handler
	// a chance to use them. It must check that the types are correct, or
	// return an error.
	Plug(exposed map[string]interface{}) error
}

// Friendly can befriend other services, but doesn't depend on them.
type Friendly interface {
	// Likes returns a set of service identifiers this service can
	// befriend.
	Likes() map[string]struct{}

	// Befriend is called every time a service it likes just started
	// running or is about to stop. If it just started running, it is
	// passed the exposed object. If it is about to stop, nil is given.
	// It must check that the exposed type is valid before using it.
	Befriend(serviceID string, exposed interface{})
}

// Exposer exposes a type to other services.
type Exposer interface {
	// Expose exposes a type to other services. Services that depend on
	// this service will receive the returned object in their Plug method
	// if they have one. Services that are friendly with this services will
	// receive the returned object in their Befriend method.
	Expose() interface{}
}

// Runner runs a function.
type Runner interface {
	// Run should start the service. It should block until the service is
	// done or the context is canceled. It should call running() once it
	// has started, and stopping() when it begins stopping.
	Run(ctx context.Context, running, stopping func()) error
}

// StatusCode represents the status of a service.
type StatusCode int8

// Service statuses.
const (
	Stopped StatusCode = iota
	Starting
	Running
	Stopping
	Errored
)

// String returns a string representation of a service status.
func (s StatusCode) String() string {
	switch s {
	case Stopped:
		return "stopped"
	case Starting:
		return "starting"
	case Running:
		return "running"
	case Stopping:
		return "stopping"
	case Errored:
		return "errored"
	}

	panic(errors.Errorf("invalid status code: %d", s))
}

// managerService exposes the manager itself as a service.
type managerService struct {
	mgr *Manager
}

// ID returns the unique identifier of the service.
func (s managerService) ID() string {
	return "manager"
}

// Name return the user friendly name of the service.
func (s managerService) Name() string {
	return "Service Manager"
}

// Desc return a short description of what the service does.
func (s managerService) Desc() string {
	return "Manages services."
}

// Expose exposes the manager to other services.
//
// It exposes the type:
//	github.com/stratumn/go-node/core/*manager.Manager
func (s managerService) Expose() interface{} {
	return s.mgr
}

// AddToGRPCServer adds the service to a gRPC server.
func (s managerService) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterManagerServer(gs, grpcServer{mgr: s.mgr})
}
