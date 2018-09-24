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

// Package service defines a service for the P2P relay circuit, which enables
// nodes to send traffic through intermediary nodes in order to reach otherwise
// inaccessible nodes.
package service

import (
	"context"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	circuit "gx/ipfs/QmWX6RySJ3yAYmfjLSw1LtRZnDh5oVeA9kM3scNQJkysqa/go-libp2p-circuit"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// log is the logger for the service.
var log = logging.Logger("relay")

// Host represents an Indigo Node host.
type Host = ihost.Host

// Service is the Relay service.
type Service struct {
	config *Config
	host   Host
}

// Config contains configuration options for the Relay service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// EnableHop is whether to act as an intermediary node in relay
	// circuits.
	EnableHop bool `toml:"enable_hop" comment:"Whether to act as an intermediary node in relay circuits."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "relay"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Relay"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Enables the P2P circuit relay transport."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Host:      "host",
		EnableHop: false,
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

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	relayCtx, cancelRelay := context.WithCancel(ctx)
	defer cancelRelay()

	opts := []circuit.RelayOpt{circuit.OptActive}

	if s.config.EnableHop {
		log.Event(ctx, "hopEnabled")
		opts = append(opts, circuit.OptHop)
	} else {
		log.Event(ctx, "hopDisabled")
	}

	err := circuit.AddRelayTransport(relayCtx, s.host, nil, opts...)
	if err != nil {
		return errors.WithStack(err)
	}

	running()
	<-ctx.Done()
	stopping()

	s.host.RemoveStreamHandler(protocol.ID(circuit.ProtoID))

	cancelRelay()

	return errors.WithStack(ctx.Err())
}
