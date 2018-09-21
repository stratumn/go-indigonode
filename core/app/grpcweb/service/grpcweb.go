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

// Package service wraps the grpcapi server to implement the gRPC-Web spec
package service

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/stratumn/go-indigonode/core/httputil"
)

var (
	// ErrNotGrpcServer is returned when the connected service is not a gRPC Server.
	ErrNotGrpcServer = errors.New("connected service is not a gRPC Server")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Service is the gRPC-Web service.
type Service struct {
	config *Config
	server *grpcweb.WrappedGrpcServer
}

// Config contains configuration options for the Clock service.
type Config struct {
	// Host is the name of the host service.
	Grpcapi string `toml:"grpcapi" comment:"The name of the grpcapi service."`

	// Address is the address to bind to.
	Address string `toml:"address" comment:"Address to bind to."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "grpcweb"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "GRPC-Web"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Wraps the grpcapi server to implement the gRPC-Web spec."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Grpcapi: "grpcapi",
		Address: "/ip4/127.0.0.1/tcp/8906",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	if _, err := ma.NewMultiaddr(conf.Address); err != nil {
		return errors.WithStack(err)
	}

	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Grpcapi] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	gs, ok := exposed[s.config.Grpcapi].(*grpc.Server)
	if !ok {
		return errors.Wrap(ErrNotGrpcServer, s.config.Grpcapi)
	}
	s.server = grpcweb.WrapServer(gs)

	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	var err error
	serverCtx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()

	serverDone := make(chan error, 1)

	go func() {
		serverDone <- httputil.StartServer(serverCtx, s.config.Address, s.server)
	}()

	running()

	select {
	case err = <-serverDone:
		stopping()

	case <-ctx.Done():
		stopping()
		cancelServer()
		err = <-serverDone
	}

	s.server = nil

	if err == nil {
		return errors.WithStack(ctx.Err())
	}

	return err
}
