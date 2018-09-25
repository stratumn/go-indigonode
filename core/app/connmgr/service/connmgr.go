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

// Package service defines a service that manages the number of connections
// kept open.
package service

import (
	"context"
	"time"

	"github.com/pkg/errors"

	connmgr "gx/ipfs/QmW9pfNup4hcWxyMxDGSe25tG9xepvLqqmQUoTDaawzTZe/go-libp2p-connmgr"
)

// Service is the Connection Manager service.
type Service struct {
	config *Config
	grace  time.Duration
	cmgr   *connmgr.BasicConnMgr
}

// Config contains configuration options for the Connection Manager service.
type Config struct {
	// LowWater is the minimum number of connections to keep open.
	LowWater int `toml:"low_water" comment:"Minimum number of connections to keep open (0 = disabled)."`

	// HighWater is the maximum number of connections to keep open.
	HighWater int `toml:"high_water" comment:"Maximum number of connections to keep open (0 = disabled)."`

	// GracePeriod is how long to keep a connection before it can be closed.
	GracePeriod string `toml:"grace_period" comment:"How long to keep a connection before it can be closed."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "connmgr"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Connection Manager"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Manages connections to peers."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		LowWater:    600,
		HighWater:   900,
		GracePeriod: "20s",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	grace, err := time.ParseDuration(conf.GracePeriod)
	if err != nil {
		return errors.WithStack(err)
	}

	s.grace = grace
	s.config = &conf

	return nil
}

// Expose exposes the connection manager to other services.
//
// It exposes the type:
//  	github.com/libp2p/*go-libp2p-connmgr.BasicConnMgr
func (s *Service) Expose() interface{} {
	return s.cmgr
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.cmgr = connmgr.NewConnManager(s.config.LowWater, s.config.HighWater, s.grace)

	running()
	<-ctx.Done()
	stopping()

	s.cmgr = nil

	return errors.WithStack(ctx.Err())
}
