# Extend Alice By Writing Your Own Service

Great care went into making Alice extendable with minimum hassle. The core of
Alice handles service dependencies and configuration files. Extending the API
is a little more work but it's still very reasonable, and the CLI uses
reflection to automatically add new commands, so you don't have to worry about
that part.

Adding new services to Alice is easy. A sevice needs to implement at least the
three methods of the `Service` interface (`ID`, `Name`, and `Desc`). You can
start from this template, which implements all the most commonly used
interfaces:

```go
package myservice

import (
	"context"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// log is the logger for the service.
var log = logging.Logger("myservice")

// Service is the My Service service.
type Service struct {
	config *Config
}

// Config contains configuration options for the My Service service.
type Config struct {
	// Example is an example setting.
	Example string `toml:"example" comment:"Example setting."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "myservice"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "My Service"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "My own little service."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	// Set the default configuration settings of your service here.
	return Config{
		Example: "Default value",
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	// If your service depends on other services, specify it here.
	// For example, if you need the "manager" service:
	//
	//  return map[string]struct{}{
	//		"manager": struct{}{}
	//  }
	return nil
}

// Plug sets the connected services.
func (s *Service) Plug(services map[string]interface{}) error {
	// If your service depends on other services, it will be given what those
	// services exposed here. For example:
	//
	// var ok
	// s.mgr, ok = services["manager"].(*manager.Manager)
	// if !ok {
	//		return errors.New("invalid manager type")
	// }
	return nil
}

// Expose exposes the service to other services.
func (s *Service) Expose() interface{} {
	// If you want to expose something to other services, this is the place to
	// do it. For example:
	//
	// return s.mgr
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {

	// Initialize the service before sending an empty struct to the running
	// channel.

	running <- struct{}{}

	// Start any long running process with a goroutine here.

	// Handle exit conditions.
	select {
	case <-ctx.Done():
    // ...
	}

	stopping <- struct{}{}

	// Stop the service after sending an empty struct the the stopping channel.

	return errors.WithStack(ctx.Err())
}
```

It should be self-explanatory if you are experienced with the Go programming
language.

To build `alice` with your service included, all you have to do is register
your package somewhere. For instance the core services are registered in
`core/service.go`:

```go
// Copyright © 2017 Stratumn SAS
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

package core

import (
	"github.com/stratumn/alice/core/manager"
	"github.com/stratumn/alice/core/service/bootstrap"
	"github.com/stratumn/alice/core/service/connmgr"
	"github.com/stratumn/alice/core/service/grpcapi"
	"github.com/stratumn/alice/core/service/host"
	"github.com/stratumn/alice/core/service/identify"
	"github.com/stratumn/alice/core/service/kaddht"
	"github.com/stratumn/alice/core/service/metrics"
	"github.com/stratumn/alice/core/service/mssmux"
	"github.com/stratumn/alice/core/service/natmgr"
	"github.com/stratumn/alice/core/service/ping"
	"github.com/stratumn/alice/core/service/pruner"
	"github.com/stratumn/alice/core/service/relay"
	"github.com/stratumn/alice/core/service/signal"
	"github.com/stratumn/alice/core/service/swarm"
	"github.com/stratumn/alice/core/service/yamux"
)

// services contains all the services.
var services = []manager.Service{
	&grpcapi.Service{},
	&pruner.Service{},
	&signal.Service{},
	&yamux.Service{},
	&mssmux.Service{},
	&swarm.Service{},
	&connmgr.Service{},
	&host.Service{},
	&natmgr.Service{},
	&metrics.Service{},
	&relay.Service{},
	&identify.Service{},
	&bootstrap.Service{},
	&kaddht.Service{},
	&ping.Service{},
}
```

After registering your service, you can build `alice` using `go build`. To add
the configuration settings of your service to an existing `alice.core.toml`
file, run `alice init --recreate` (TODO: this doesn work very well).

You should now be able to start your service from the CLI using
`manager-start myservice`.

## Extending The API/CLI

The extend the API, you must first implement a [gRPC](https://grpc.io) service.

[TODO: Quick How To]

Then your service can implement the `Registrable` interface:

```go
// Registrable represents something that can add itself to the gRPC server, so
// that other services can add functions to the API.
type Registrable interface {
	AddToGRPCServer(*grpc.Server)
}
```

All you need to do is add a `AddToGRPCServer(*grpc.Server)` function to your
service. For instance:

```go
// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterMyServiceServer(gs, myGRPCServer{s})
}
```

The `grpcapi` service automatically looks for services that expose a
`Registrable` and adds them to the gRPC server.

You don't need to manually add commands to the CLI. Using `reflection`, the CLI
inspects the services available on the gRPC server, and automatically creates
commands for them when it connects to the API. Keep in mind that currently it
is only able to reflect very basic gRPC methods.

## Interfaces Compatible With The Service Manager

The service manager understands the following interfaces:

```go
// Copyright © 2017 Stratumn SAS
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
```

## Available Services

| ID | NAME | DESC | EXPOSES |
| --- | --- | --- | --- |
| api | API Services | Starts API services. |  |
| boot | Boot Services | Starts boot services. | struct{}{} |
| bootstrap | Bootstrap | Periodically connects to known peers. |  |
| connmgr | Connection Manager | Manages connections to peers. | github.com/libp2p/\*go-libp2p-connmgr.BasicConnMgr |
| grpcapi | gRPC API | Starts a gRPC API server. |  |
| host | Host | Starts a P2P host. | github.com/stratumn/alice/core/service/\*host.Host |
| identify | Identify | Identifies peers. | github.com/libp2p/go-libp2p/p2p/protocols/\*identify.IDService |
| kaddht | Kademlia DHT | Manages a Kademlia distributed hash table. | github.com/libp2p/\*go-libp2p-kad-dht.IpfsDHT |
| manager | Service Manager | Manages services. | github.com/stratumn/alice/core/\*manager.Manager |
| metrics | Metrics | Collects metrics. | github.com/stratumn/alice/core/service/\*metrics.Metrics |
| mssmux | Stream Muxer Router | Routes protocols to stream muxers. | github.com/libp2p/go-stream-muxer.Transport |
| natmgr | NAT Manager | Manages NAT port mappings. | github.com/libp2p/go-libp2p/p2p/host/basic.NATManager |
| network | Network Services | Starts network services. |  |
| p2p | P2P Services | Starts P2P services. |  |
| ping | Ping | Handles ping requests and responses. | github.com/libp2p/go-libp2p/p2p/protocols/\*ping.PingService |
| pruner | Service Pruner | Prunes unused services. |  |
| relay | Relay | Enables the P2P circuit relay transport. |  |
| signal | Signal Handler | Handles exit signals. |  |
| swarm | Swarm | Connects to peers. | github.com/libp2p/\*go-libp2p-swarm.Swarm |
| system | System Services | Starts system services. |  |
| yamux | Yamux | Multiplexes streams using Yamux. | github.com/libp2p/go-stream-muxer.Transport |
