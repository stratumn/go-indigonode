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

// Package grpcweb wraps the grpcapi server to implement the gRPC-Web spec
package grpcweb

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/netutil"
	"google.golang.org/grpc"

	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
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
	lis, err := netutil.Listen(s.config.Address)
	if err != nil {
		return err
	}

	serverCtx, cancelServer := context.WithCancel(ctx)
	defer cancelServer()

	serverDone := make(chan error, 1)

	go func() {
		serverDone <- s.startServer(serverCtx, lis)
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

// startServer starts exposing gRPC-Web API on an HTTP endpoint.
func (s *Service) startServer(ctx context.Context, lis net.Listener) error {
	srv := http.Server{Handler: s.server}

	done := make(chan error, 1)
	go func() {
		err := srv.Serve(lis)
		if err != nil && err != http.ErrServerClosed {
			done <- err
			return
		}

		close(done)
	}()

	shutdown := func() error {
		shutdownCtx, cancelShutdown := context.WithTimeout(
			context.Background(),
			time.Second/2,
		)
		defer cancelShutdown()

		return srv.Shutdown(shutdownCtx)
	}

	select {
	case err := <-done:
		return errors.WithStack(err)

	case <-ctx.Done():
		for {
			if err := shutdown(); err != nil {
				return errors.WithStack(err)
			}

			select {
			case err := <-done:
				if err != nil {
					return errors.WithStack(err)
				}
			case <-time.After(time.Second / 2):
				// Serve will not stop if we call Shutdown
				// before Serve, so in case this happens we
				// try again (for instance during tests.).
				continue
			}

			return errors.WithStack(ctx.Err())
		}
	}
}
