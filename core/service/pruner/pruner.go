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

// Package pruner defines a service that periodically prunes the service
// manager.
package pruner

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrNotManager is returned when the connected service is not a
	// manager.
	ErrNotManager = errors.New("connected service is not a manager")
)

// Manager represents a service manager.
type Manager interface {
	Prune()
}

// Service is the Service Pruner service.
type Service struct {
	config   *Config
	interval time.Duration
	mgr      Manager
}

// Config contains configuration options for the Service Pruner service.
type Config struct {
	// Manager is the name of the manager service.
	Manager string `toml:"manager" comment:"The name of the manager service."`

	// Interval is the duration of the interval between prune jobs.
	Interval string `toml:"interval" comment:"Interval between prune jobs."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "pruner"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Service Pruner"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Prunes unused services."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Manager:  "manager",
		Interval: "1m",
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	interval, err := time.ParseDuration(conf.Interval)
	if err != nil {
		return errors.WithStack(err)
	}

	s.config = &conf
	s.interval = interval

	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Manager] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(handlers map[string]interface{}) error {
	var ok bool

	if s.mgr, ok = handlers[s.config.Manager].(Manager); !ok {
		return errors.Wrap(ErrNotManager, s.config.Manager)
	}

	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	running()

	for {
		select {
		case <-time.After(s.interval):
			// Avoid race condition if pruner is stopped.
			done := make(chan struct{})
			go func() {
				s.mgr.Prune()
				close(done)
			}()

			select {
			case <-done:
				// We're still alive!
			case <-ctx.Done():
				// We stopped ourself.
				stopping()
				return errors.WithStack(ctx.Err())
			}
		case <-ctx.Done():
			stopping()
			return errors.WithStack(ctx.Err())
		}
	}
}
