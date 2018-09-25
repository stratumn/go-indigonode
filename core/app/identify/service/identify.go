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

// Package service defines a service that identifies peers on the network.
package service

import (
	"context"

	"github.com/pkg/errors"

	identify "gx/ipfs/QmUEqyXr97aUbNmQADHYNknjwjjdVpJXEt1UZXmSG81EV4/go-libp2p/p2p/protocol/identify"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// Host represents an Indigo Node host.
type Host interface {
	ihost.Host

	SetIDService(*identify.IDService)
}

// Service is the Identify service.
type Service struct {
	config *Config
	host   Host
	ids    *identify.IDService
}

// Config contains configuration options for the Identify service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "identify"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Identify"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Identifies peers."
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

// Expose exposes the identify service to other services.
//
// It exposes the type:
//	github.com/libp2p/go-libp2p/p2p/protocols/*identify.IDService
func (s *Service) Expose() interface{} {
	return s.ids
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.ids = identify.NewIDService(s.host)
	s.host.SetIDService(s.ids)

	running()
	<-ctx.Done()
	stopping()

	s.host.SetIDService(nil)
	s.host.RemoveStreamHandler(identify.ID)
	s.ids = nil

	return errors.WithStack(ctx.Err())
}
