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

package relay

import (
	"context"

	"github.com/pkg/errors"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	host "gx/ipfs/Qmc1XhrFEiSeBNn3mpfg6gEuYCt5im2gYmNVmncsvmpeAk/go-libp2p-host"
	circuit "gx/ipfs/Qmf7GSJ4omRJsvA9uzTqzbnVhq4RWLPzjzW4xJzUta4dKE/go-libp2p-circuit"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// log is the logger for the service.
var log = logging.Logger("relay")

// Service is the Relay service.
type Service struct {
	config *Config
	host   host.Host
}

// Config contains configuration options for the Relay service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Example is an example setting.
	EnableHop bool `toml:"enable_hop" comment:"Whether to act as an intermediate node in relay circuits."`
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

	if s.host, ok = exposed[s.config.Host].(host.Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {
	relayCtx, cancelRelay := context.WithCancel(ctx)
	defer cancelRelay()

	opts := []circuit.RelayOpt{circuit.OptActive}

	if s.config.EnableHop {
		log.Event(ctx, "hopEnabled")
		opts = append(opts, circuit.OptHop)
	} else {
		log.Event(ctx, "hopDisabled")
	}

	err := circuit.AddRelayTransport(relayCtx, s.host, opts...)
	if err != nil {
		return errors.WithStack(err)
	}

	running <- struct{}{}
	<-ctx.Done()
	stopping <- struct{}{}

	cancelRelay()

	return errors.WithStack(ctx.Err())
}
