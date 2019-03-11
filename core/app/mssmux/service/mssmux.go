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

// Package service defines a service that routes transport protocols to stream
// multiplexers.
package service

import (
	"context"

	"github.com/pkg/errors"

	logging "github.com/ipfs/go-log"
	smux "github.com/libp2p/go-stream-muxer"
	mssmux "github.com/whyrusleeping/go-smux-multistream"
)

var (
	// ErrNoStreamMuxers is returned when no stream muxers were specified.
	ErrNoStreamMuxers = errors.New("at least one stream muxer is required")

	// ErrNotStreamMuxer is returned when a specified service is not a stream muxer.
	ErrNotStreamMuxer = errors.New("connected service is not a stream muxer")
)

// log is the logger for the service.
var log = logging.Logger("mssmux")

// Service is the Stream Muxer Router service.
type Service struct {
	config *Config
	tpt    smux.Transport
}

// Config contains configuration options for the Stream Muxer Router
// service.
type Config struct {
	// Routes maps protocols to stream muxer services.
	Routes map[string]string `toml:"routes" comment:"A map of protocols to stream muxers (protocol = service)."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "mssmux"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Stream Muxer Router"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Routes protocols to stream muxers."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Routes: map[string]string{
			"/yamux/v1.0.0": "yamux",
		},
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	if len(conf.Routes) < 1 {
		return errors.WithStack(ErrNoStreamMuxers)
	}

	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}

	for _, service := range s.config.Routes {
		needs[service] = struct{}{}
	}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	tpt := mssmux.NewBlankTransport()

	for protocol, service := range s.config.Routes {
		smuxer, ok := exposed[service].(smux.Transport)
		if !ok {
			return errors.Wrap(ErrNotStreamMuxer, service)
		}

		tpt.AddTransport(protocol, smuxer)
		log.Event(context.Background(), "routeAdded", logging.Metadata{
			"protocol": protocol,
			"service":  service,
		})
	}

	s.tpt = tpt

	return nil
}

// Expose exposes the stream muxer to other services.
//
// It exposes the type:
//	github.com/libp2p/go-stream-muxer.Transport
func (s *Service) Expose() interface{} {
	return s.tpt
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	running()
	<-ctx.Done()
	stopping()

	s.tpt = nil

	return errors.WithStack(ctx.Err())
}
