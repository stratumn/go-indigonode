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

package connmgr

import (
	"context"
	"time"

	"github.com/pkg/errors"

	connmgr "gx/ipfs/QmYr5RRXTXwEsG7sq1nP1JioZ8RmiFUktnBdoNBvTovgGj/go-libp2p-connmgr"
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
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {
	s.cmgr = connmgr.NewConnManager(s.config.LowWater, s.config.HighWater, s.grace)

	running <- struct{}{}
	<-ctx.Done()
	stopping <- struct{}{}

	s.cmgr = nil

	return errors.WithStack(ctx.Err())
}
