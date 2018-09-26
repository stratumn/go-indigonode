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

// Package service defines types for the chat service.
package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/app/chat/grpc"
	chat "github.com/stratumn/go-node/app/chat/protocol"
	event "github.com/stratumn/go-node/core/app/event/service"

	"google.golang.org/grpc"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	pstore "gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
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

// Host represents a Stratumn Node host.
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
