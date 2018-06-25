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

// Package service defines types for the chat service.
package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/app/chat/grpc"
	chat "github.com/stratumn/go-indigonode/app/chat/protocol"
	event "github.com/stratumn/go-indigonode/core/app/event/service"

	"google.golang.org/grpc"

	inet "gx/ipfs/QmXoz9o2PT3tEzf7hicegwex5UgVP54n3k82K7jrWFyN86/go-libp2p-net"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	pstore "gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
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

// Service is the Chat service.
type Service struct {
	config       *Config
	host         Host
	eventEmitter event.Emitter
	chat         *chat.Chat
	mgr          *Manager
}

// Config contains configuration options for the Chat service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Event is the name of the event service.
	Event string `toml:"event" comment:"The name of the event service."`

	// HistoryDBPath is the path of the chat history db file.
	HistoryDBPath string `toml:"history_db_path" comment:"The path of the chat history db file."`
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

	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.WithStack(err))
	}

	filename, err := filepath.Abs(filepath.Join(cwd, "data", "chat-history.db"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	// Set the default configuration settings of your service here.
	return Config{
		Host:          "host",
		Event:         "event",
		HistoryDBPath: filename,
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
	msgReceivedCh := make(chan *pb.DatedMessage)
	s.chat = chat.NewChat(s.host, s.eventEmitter, msgReceivedCh)
	mgr, err := NewManager(s.config.HistoryDBPath)
	if err != nil {
		return nil
	}

	s.mgr = mgr

	// Wrap the stream handler with the context.
	s.host.SetStreamHandler(chat.ProtocolID, func(stream inet.Stream) {
		s.chat.StreamHandler(ctx, stream)
	})

	running()
LOOP:
	for {
		select {
		case m := <-msgReceivedCh:
			from, err := m.FromPeer()
			if err != nil {
				return err
			}
			err = s.mgr.Add(from, m)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			break LOOP
		}
	}
	stopping()

	// Stop accepting streams.
	s.host.RemoveStreamHandler(chat.ProtocolID)
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
			err := s.chat.Send(ctx, pid, message)
			if err != nil {
				return err
			}
			return s.mgr.Add(pid, pb.NewDatedMessageSent(pid, message))
		},
		GetPeerHistory: func(pid peer.ID) (PeerHistory, error) {
			return s.mgr.Get(pid)
		},
	})
}
