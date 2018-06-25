# Extend Indigo Node By Writing Your Own Service

Great care went into making Indigo Node extendable with minimum hassle. The core of
Indigo Node handles service dependencies and configuration files. Extending the API
is a little more work but it's still very reasonable, and the CLI uses
reflection to automatically add new commands, so you don't have to worry about
that part.

Adding new services to Indigo Node is easy. A service needs to implement at least the
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
	// If your service depends on other services, specify them here.
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
func (s *Service) Run(ctx context.Context, running, stopping func()) error {

	// Initialize the service before calling running().

	running()

	// Start any long running process with a goroutine here.

	// Handle exit conditions.
	select {
	case <-ctx.Done():
	// ...
	}

	stopping()

	// Stop the service after calling stopping().

	return errors.WithStack(ctx.Err())
}
```

It should be self-explanatory if you are experienced with the Go programming
language.

To build `indigo-node` with your service included, all you have to do is register
your package somewhere. For instance the core services are registered in
`core/service.go`:

```go
// Copyright © 2017-2018 Stratumn SAS
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
	"github.com/stratumn/go-indigonode/core/manager"
	bootstrap "github.com/stratumn/go-indigonode/core/app/bootstrap/service"
	connmgr "github.com/stratumn/go-indigonode/core/app/connmgr/service"
	grpcapi "github.com/stratumn/go-indigonode/core/app/grpcapi/service"
	host "github.com/stratumn/go-indigonode/core/app/host/service"
	identify "github.com/stratumn/go-indigonode/core/app/identify/service"
	kaddht "github.com/stratumn/go-indigonode/core/app/kaddht/service"
	metrics "github.com/stratumn/go-indigonode/core/app/metrics/service"
	mssmux "github.com/stratumn/go-indigonode/core/app/mssmux/service"
	natmgr "github.com/stratumn/go-indigonode/core/app/natmgr/service"
	ping "github.com/stratumn/go-indigonode/core/app/ping/service"
	pruner "github.com/stratumn/go-indigonode/core/app/pruner/service"
	relay "github.com/stratumn/go-indigonode/core/app/relay/service"
	signal "github.com/stratumn/go-indigonode/core/app/signal/service"
	swarm "github.com/stratumn/go-indigonode/core/app/swarm/service"
	yamux "github.com/stratumn/go-indigonode/core/app/yamux/service"
)

// BuiltinServices returns all the builtin services.
func BuiltinServices() []manager.Service {
	return []manager.Service{
		&bootstrap.Service{},
		&clock.Service{},
		&connmgr.Service{},
		&grpcapi.Service{},
		&host.Service{},
		&identify.Service{},
		&kaddht.Service{},
		&metrics.Service{},
		&mssmux.Service{},
		&natmgr.Service{},
		&ping.Service{},
		&pruner.Service{},
		&relay.Service{},
		&signal.Service{},
		&swarm.Service{},
		&yamux.Service{},
	}
}
```

If your service has configuration options, you should add a migration to add
them to the config file. The migrations for the core Indigo Node modules are in
`core/migrate.go`. You can append one for your service:

```go
var migrations = []cfg.MigrateHandler{

	// Previous migrations...

	func(tree *cfg.Tree) error {
		return tree.Set("myservice.option", "value")
	},
}
```

After registering your service, you can build `indigo-node` using `go build`.

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

The `grpcapi` service automatically looks for services that implement
`Registrable` and registers them on the gRPC server.

You don't need to manually add commands to the CLI. Using `reflection`, the CLI
inspects the services available on the gRPC server, and automatically creates
commands for them when it connects to the API. Keep in mind that currently it
is only able to reflect very basic gRPC methods.

## Interfaces Compatible With The Service Manager

The service manager understands the following interfaces:

```go
// Copyright © 2017-2018 Stratumn SAS
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

// Service describes an Indigo Node service.
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
```

## Available Services

| ID        | NAME                | DESC                                       | EXPOSES                                                          |
| --------- | ------------------- | ------------------------------------------ | ---------------------------------------------------------------- |
| api       | API Services        | Starts API services.                       |                                                                  |
| boot      | Boot Services       | Starts boot services.                      |                                                                  |
| bootstrap | Bootstrap           | Periodically connects to known peers.      | struct{}{}                                                       |
| connmgr   | Connection Manager  | Manages connections to peers.              | github.com/libp2p/\*go-libp2p-connmgr.BasicConnMgr               |
| grpcapi   | gRPC API            | Starts a gRPC API server.                  |                                                                  |
| host      | Host                | Starts a P2P host.                         | github.com/stratumn/go-indigonode/core/\*p2p.Host                |
| identify  | Identify            | Identifies peers.                          | github.com/libp2p/go-libp2p/p2p/protocols/\*identify.IDService   |
| kaddht    | Kademlia DHT        | Manages a Kademlia distributed hash table. | github.com/libp2p/\*go-libp2p-kad-dht.IpfsDHT                    |
| manager   | Service Manager     | Manages services.                          | github.com/stratumn/go-indigonode/core/\*manager.Manager         |
| metrics   | Metrics             | Collects metrics.                          | github.com/stratumn/go-indigonode/core/service/\*metrics.Metrics |
| mssmux    | Stream Muxer Router | Routes protocols to stream muxers.         | github.com/libp2p/go-stream-muxer.Transport                      |
| natmgr    | NAT Manager         | Manages NAT port mappings.                 | github.com/libp2p/go-libp2p/p2p/host/basic.NATManager            |
| network   | Network Services    | Starts network services.                   |                                                                  |
| p2p       | P2P Services        | Starts P2P services.                       |                                                                  |
| ping      | Ping                | Handles ping requests and responses.       | github.com/libp2p/go-libp2p/p2p/protocols/\*ping.PingService     |
| pruner    | Service Pruner      | Prunes unused services.                    |                                                                  |
| relay     | Relay               | Enables the P2P circuit relay transport.   |                                                                  |
| signal    | Signal Handler      | Handles exit signals.                      |                                                                  |
| swarm     | Swarm               | Connects to peers.                         | github.com/libp2p/\*go-libp2p-swarm.Swarm                        |
| system    | System Services     | Starts system services.                    |                                                                  |
| yamux     | Yamux               | Multiplexes streams using Yamux.           | github.com/libp2p/go-stream-muxer.Transport                      |
