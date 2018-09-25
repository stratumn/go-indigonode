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

// Package service defines types for the clock service.
package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/app/clock/grpc"
	"github.com/stratumn/go-node/app/clock/protocol"
	"google.golang.org/grpc"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	pstore "gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Host represents a Stratumn Node host.
type Host = ihost.Host

// Service is the Clock service.
type Service struct {
	config  *Config
	host    Host
	timeout time.Duration
	clock   *protocol.Clock
}

// Config contains configuration options for the Clock service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// WriteTimeout sets how long to wait before closing the stream when
	// writing the time to a peer.
	WriteTimeout string `toml:"write_timeout" comment:"How long to wait before closing the stream when writing the time to a peer."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "clock"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Clock"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Returns the time of a node."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Host:         "host",
		WriteTimeout: "10s",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	timeout, err := time.ParseDuration(conf.WriteTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	s.timeout = timeout
	s.config = &conf

	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Host] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.host, ok = exposed[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	return nil
}

// Expose exposes the clock service to other services.
//
// It exposes the type:
//	github.com/stratumn/go-node/core/service/*clock.Clock
func (s *Service) Expose() interface{} {
	return s.clock
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.clock = protocol.NewClock(s.host, s.timeout)

	// Wrap the stream handler with the context.
	handler := func(stream inet.Stream) {
		s.clock.StreamHandler(ctx, stream)
	}

	s.host.SetStreamHandler(s.clock.ProtocolID(), handler)

	running()
	<-ctx.Done()
	stopping()

	// Stop accepting streams.
	s.host.RemoveStreamHandler(s.clock.ProtocolID())

	// Wait for all the streams to close.
	s.clock.Wait()

	s.clock = nil

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterClockServer(gs, grpcServer{
		LocalTime: func(ctx context.Context) (*time.Time, error) {
			if s.clock == nil {
				return nil, ErrUnavailable
			}

			t := time.Now()
			return &t, nil
		},
		RemoteTime: func(ctx context.Context, pid peer.ID) (*time.Time, error) {
			if s.clock == nil {
				return nil, ErrUnavailable
			}

			return s.clock.RemoteTime(ctx, pid)
		},
		Connect: func(ctx context.Context, pi pstore.PeerInfo) error {
			if s.host == nil {
				return ErrUnavailable
			}

			return s.host.Connect(ctx, pi)
		},
	})
}
