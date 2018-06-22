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

// Package service defines a service that routes transport protocols to stream
// multiplexers.
package service

import (
	"context"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	mssmux "gx/ipfs/QmVniQJkdzLZaZwzwMdd3dJTvWiJ1DQEkreVy6hs6h7Vk5/go-smux-multistream"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
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
