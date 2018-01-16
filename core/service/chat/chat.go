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

// Package chat is a simple service that allows two peers to exchange messages.
package chat

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/service/event"
	pb "github.com/stratumn/alice/grpc/chat"

	"google.golang.org/grpc"

	ihost "gx/ipfs/QmP46LGWhzVZTMmt5akNNLfoV8qL4h5wTwmzQxLyDafggd/go-libp2p-host"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	inet "gx/ipfs/QmU4vCDZTPLDqSDKguWbHCiUe46mZUtmM2g2suBZ9NE8ko/go-libp2p-net"
	peer "gx/ipfs/QmWNY7dV54ZDYmTA1ykVdwNCqC11mpU4zSUp6XDpLTH9eG/go-libp2p-peer"
	pstore "gx/ipfs/QmYijbtjCxFEjSXaudaQAUz3LN5VKLssm8WCUsRoqzXmQR/go-libp2p-peerstore"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrNotEvent is returned when the connected service is not an event emitter.
	ErrNotEvent = errors.New("connected service is not an event emitter")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Host represents an Alice host.
type Host = ihost.Host

// log is the logger for the service.
var log = logging.Logger("chat")

// Service is the Chat service.
type Service struct {
	config       *Config
	host         Host
	eventEmitter event.Emitter
	chat         *Chat
}

// Config contains configuration options for the Chat service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Event is the name of the event service.
	Event string `toml:"event" comment:"The name of the event service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "chat"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Chat"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "A basic p2p chat service."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	// Set the default configuration settings of your service here.
	return Config{
		Host:  "host",
		Event: "event",
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
	needs := map[string]struct{}{}
	needs[s.config.Host] = struct{}{}
	needs[s.config.Event] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(services map[string]interface{}) error {
	var ok bool

	if s.host, ok = services[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	if s.eventEmitter, ok = services[s.config.Event].(event.Emitter); !ok {
		return errors.Wrap(ErrNotEvent, s.config.Event)
	}

	return nil
}

// Expose exposes the chat service to other services.
// It is currently not exposed.
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.chat = NewChat(s.host, s.eventEmitter)

	// Wrap the stream handler with the context.
	s.host.SetStreamHandler(ProtocolID, func(stream inet.Stream) {
		s.chat.StreamHandler(ctx, stream)
	})

	running()
	<-ctx.Done()
	stopping()

	// Stop accepting streams.
	s.host.RemoveStreamHandler(ProtocolID)
	s.chat = nil

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterChatServer(gs, grpcServer{
		Connect: func(ctx context.Context, pi pstore.PeerInfo) error {
			if s.host == nil {
				return ErrUnavailable
			}

			return s.host.Connect(ctx, pi)
		},
		Send: func(ctx context.Context, pid peer.ID, message string) error {
			return s.chat.Send(ctx, pid, message)
		},
	})
}
