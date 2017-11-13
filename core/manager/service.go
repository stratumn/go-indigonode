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
	pb "github.com/stratumn/alice/grpc/manager"
	"google.golang.org/grpc"
)

// Service describes an Alice service.
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
	// Needs returns a list of service identifiers needed before this
	// service can start.
	Needs() map[string]struct{}
}

// Pluggable connects other services.
type Pluggable interface {
	Needy

	// Plug is given a map of exposed connected objects, giving the handler
	// a chance to save them. It must check that the types are correct, or
	// return an error.
	Plug(exposed map[string]interface{}) error
}

// Friendly can befriend other services, but doesn't depend on them.
type Friendly interface {
	// Likes returns a list of service identifiers this service can
	// befriend.
	Likes() map[string]struct{}

	// Befriend is called every time a service it likes just started
	// running or is about to stop. If it just started running, it is
	// passed the exposed object. If it is about to stop, nil is given.
	// It must check that the exposed type is valid before saving it.
	Befriend(serviceID string, exposed interface{})
}

// Exposer exposes something to other services.
type Exposer interface {
	// Expose exposes the service to other services. Services that depend
	// on this service will receive the return object in their Plug method.
	// Services that are friendly with this services will receive the
	// returned object in their Befriend method.
	Expose() interface{}
}

// Runner runs a function.
type Runner interface {
	// Run should start the service. It should block until the service is
	// done or the context is canceled. It should send a message to running
	// once it has started, and to stopping when it begins stopping.
	Run(ctx context.Context, running chan struct{}, stopping chan struct{}) error
}

// StatusCode represents the status of a service.
type StatusCode int8

// Service statuses.
const (
	Stopped StatusCode = iota
	Prestarting
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
//	github.com/stratumn/alice/core/*manager.Manager
func (s managerService) Expose() interface{} {
	return s.mgr
}

// AddToGRPCServer adds the service to a gRPC server.
func (s managerService) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterManagerServer(gs, (grpcServer)(s))
}
