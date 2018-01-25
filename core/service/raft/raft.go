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

//go:generate mockgen -package mockraft -destination mockraft/mockraft.go github.com/stratumn/alice/core/service/raft Host

// Package raft is a simple service that sends the local time to a peer every
// time it receives a byte from that peer.
//
// It is meant to illustrate how to create network services.
package raft

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/alice/grpc/raft"
	"google.golang.org/grpc"

	protocol "github.com/stratumn/alice/core/protocol/raft"

	swarm "gx/ipfs/QmSD9fajyipwNQw3Hza2k2ifcBfbhGoC1ZHHgQBy4yqU8d/go-libp2p-swarm"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	ihost "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
	// ErrNotSwarm is returned when the connected service is not a swarm.
	ErrNotSwarm = errors.New("connected service is not a swarm")
)

type Host = ihost.Host

// log is the logger for the service.
var log = logging.Logger("raft")

// Service is the Raft service.
type Service struct {
	host         Host
	swarm        *swarm.Swarm
	grpcReceiver grpcReceiver
}

// Config returns the current service configuration or creates one with
// good default values.
type Config struct{}

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
	return Config{}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
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

	if s.host, ok = exposed["host"].(Host); !ok {
		return errors.Wrap(ErrNotHost, "host")
	}

	if s.swarm, ok = exposed["swarm"].(*swarm.Swarm); !ok {
		return errors.Wrap(ErrNotSwarm, "swarm")
	}

	return nil
}

// Expose exposes nothing for the moment
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {

	msgStartC := make(chan protocol.MessageStart)
	msgStopC := make(chan protocol.MessageStop)
	msgStatusC := make(chan protocol.MessageStatus)
	msgPeersC := make(chan protocol.MessagePeers)
	msgDiscoverC := make(chan protocol.MessageDiscover)
	msgInviteC := make(chan protocol.MessageInvite)
	msgJoinC := make(chan protocol.MessageJoin)
	msgExpelC := make(chan protocol.MessageExpel)
	msgProposeC := make(chan protocol.MessagePropose)
	msgLogC := make(chan protocol.MessageLog)
	msgFromNetC := make(chan protocol.MessageRaft)
	msgToNetC := make(chan protocol.MessageRaft)

	s.grpcReceiver.Init(
		msgStartC,
		msgStopC,
		msgStatusC,
		msgPeersC,
		msgDiscoverC,
		msgInviteC,
		msgJoinC,
		msgExpelC,
		msgProposeC,
		msgLogC,
	)

	raftProcess := protocol.NewRaftProcess(
		[]byte(s.swarm.LocalPeer()),
		msgStartC,
		msgStopC,
		msgStatusC,
		msgPeersC,
		msgDiscoverC,
		msgInviteC,
		msgJoinC,
		msgExpelC,
		msgProposeC,
		msgLogC,
		msgFromNetC,
		msgToNetC,
	)

	netProcess := protocol.NewNetProcess(
		s.host,
		msgDiscoverC,
		msgPeersC,
		msgFromNetC,
		msgToNetC,
	)

	// ctx = logging.ContextWithLoggable(ctx, logging.Metadata{
	// 	"origin": "xxxx",
	// })

	go raftProcess.Run(ctx)
	go netProcess.Run(ctx)

	running()
	<-ctx.Done()
	stopping()

	// TODO: wait

	return nil

}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterRaftServer(gs, &s.grpcReceiver)
}
