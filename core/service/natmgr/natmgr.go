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

package natmgr

import (
	"context"

	"github.com/pkg/errors"

	ihost "gx/ipfs/Qmc1XhrFEiSeBNn3mpfg6gEuYCt5im2gYmNVmncsvmpeAk/go-libp2p-host"
	bhost "gx/ipfs/QmefgzMbKZYsmHFkLqxgaTBG9ypeEjrdWRD5WXH4j1cWDL/go-libp2p/p2p/host/basic"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// host represents an Alice host.
type host interface {
	ihost.Host

	SetNATManager(bhost.NATManager)
}

// Service is the NAT Manager service.
type Service struct {
	config *Config
	host   host
	mgr    bhost.NATManager
}

// Config contains configuration options for the NAT Manager service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "natmgr"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "NAT Manager"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Manages NAT port mappings."
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

	if s.host, ok = exposed[s.config.Host].(host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	return nil
}

// Expose exposes the NAT manager to other services.
//
// It exposes thhe type:
//	github.com/libp2p/go-libp2p/p2p/host/basic.NATManager
func (s *Service) Expose() interface{} {
	return s.mgr
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {
	s.mgr = bhost.NewNATManager(s.host.Network())
	s.host.SetNATManager(s.mgr)

	running <- struct{}{}
	<-ctx.Done()
	stopping <- struct{}{}

	s.host.SetNATManager(nil)

	if err := s.mgr.Close(); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(ctx.Err())
}