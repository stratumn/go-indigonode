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

// Package service wraps coreos/raft library.
package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/app/raft/grpc"
	"github.com/stratumn/go-indigonode/app/raft/protocol"
	swarmSvc "github.com/stratumn/go-indigonode/core/app/swarm/service"

	"google.golang.org/grpc"

	swarm "gx/ipfs/QmRqfgh56f8CrqpwH7D2s6t8zQRsvPoftT3sp5Y6SUhNA3/go-libp2p-swarm"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
	// ErrNotSwarm is returned when the connected service is not a swarm.
	ErrNotSwarm = errors.New("connected service is not a swarm")
)

type host = ihost.Host

// Service is the Raft service.
type Service struct {
	host         host
	config       *Config
	swarm        *swarm.Swarm
	grpcReceiver grpcReceiver
}

// Config contains configuration options for the Raft service.
type Config struct {
	// ElectionTick is the number of Node.Tick invocations that must pass between elections
	ElectionTick int `toml:"election_tick" comment:"the number of Node.Tick invocations that must pass between elections"`
	// HeartbeatTick is the number of Node.Tick invocations that must pass between heartbeats
	HeartbeatTick int `toml:"heartbeat_tick" comment:"the number of Node.Tick invocations that must pass between heartbeats"`
	// MaxSizePerMsg limits the max size of each append message
	MaxSizePerMsg uint64 `toml:"max_size_per_msg" comment:"limits the max size of each append message"`
	// MaxInflightMsgs limits the max number of in-flight append messages during optimistic replication phase
	MaxInflightMsgs int `toml:"max_inflight_msgs" comment:"limits the max number of in-flight append messages during optimistic replication phase"`
	// TickerInterval defines the unit of time in ms
	TickerInterval uint64 `toml:"ticker_interval" comment:"defines the unit of time in ms"`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "raft"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Raft"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Non-Byzantine consensus engine"
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		ElectionTick:    10,
		HeartbeatTick:   1,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
		TickerInterval:  100,
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
	needs["host"] = struct{}{}
	needs["swarm"] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.host, ok = exposed["host"].(host); !ok {
		return errors.Wrap(ErrNotHost, "host")
	}

	pluggedSwarm, ok := exposed["swarm"].(*swarmSvc.Swarm)
	if !ok {
		return errors.Wrap(ErrNotSwarm, "swarm")
	}

	s.swarm = pluggedSwarm.Swarm

	return nil
}

// Expose exposes nothing for the moment
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	hub := protocol.NewHub()

	s.grpcReceiver.Init(hub)

	circleProcess := protocol.NewCircleProcess(
		raftImpl{},
		hub,
		[]byte(s.swarm.LocalPeer()),
		s.config.ElectionTick,
		s.config.HeartbeatTick,
		s.config.MaxSizePerMsg,
		s.config.MaxInflightMsgs,
		s.config.TickerInterval,
	)

	netProcess := protocol.NewNetProcess(s.host, hub)

	raftErrChan := make(chan error)
	netErrChan := make(chan error)

	s.host.SetStreamHandler(protocol.CircleProtocol, netProcess.CircleHandler)
	s.host.SetStreamHandler(protocol.DiscoverProtocol, netProcess.DiscoverHandler)

	go func() {
		raftErrChan <- circleProcess.Run(ctx)
	}()

	go func() {
		netErrChan <- netProcess.Run(ctx)
	}()

	running()
	<-ctx.Done()
	stopping()

	if err := <-raftErrChan; err != nil {
		return errors.WithStack(err)
	}
	if err := <-netErrChan; err != nil {
		return errors.WithStack(err)
	}

	s.host.RemoveStreamHandler(protocol.DiscoverProtocol)
	s.host.RemoveStreamHandler(protocol.CircleProtocol)

	return nil

}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterRaftServer(gs, &s.grpcReceiver)
}
