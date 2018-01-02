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

//go:generate mockgen -package mockping -destination mockping/mockping.go github.com/stratumn/alice/core/service/ping Host

// Package ping defines a service that handles ping requests and responses.
package ping

import (
	"context"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/ping"
	"google.golang.org/grpc"

	ihost "gx/ipfs/QmP46LGWhzVZTMmt5akNNLfoV8qL4h5wTwmzQxLyDafggd/go-libp2p-host"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	pstore "gx/ipfs/QmYijbtjCxFEjSXaudaQAUz3LN5VKLssm8WCUsRoqzXmQR/go-libp2p-peerstore"
	ping "gx/ipfs/Qma23bpHwQrQyvKeBemaeJh7sAoRHggPkgnge1B9489ff5/go-libp2p/p2p/protocol/ping"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Host represents an Alice host.
type Host = ihost.Host

// Service is the Ping service.
type Service struct {
	config *Config
	host   Host
	ping   *ping.PingService
}

// Config contains configuration options for the Ping service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "ping"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Ping"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Handles ping requests and responses."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Host: "host",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
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

// Expose exposes the ping service to other services.
//
// It exposes the type:
//	github.com/libp2p/go-libp2p/p2p/protocols/*ping.PingService
func (s *Service) Expose() interface{} {
	return s.ping
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.ping = ping.NewPingService(s.host)

	running()
	<-ctx.Done()
	stopping()

	s.host.RemoveStreamHandler(ping.ID)
	s.ping = nil

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterPingServer(gs, grpcServer{
		func(ctx context.Context, pid peer.ID) (<-chan time.Duration, error) {
			if s.ping == nil {
				return nil, ErrUnavailable
			}

			return s.ping.Ping(ctx, pid)
		},
		func(ctx context.Context, pi pstore.PeerInfo) error {
			if s.host == nil {
				return ErrUnavailable
			}

			return s.host.Connect(ctx, pi)
		},
	})
}
