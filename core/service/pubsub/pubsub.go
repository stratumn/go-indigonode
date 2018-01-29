// Copyright © 2017-2018  Stratumn SAS
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

//go:generate mockgen -package mockpubsub -destination mockpubsub/mockpubsub.go github.com/stratumn/alice/core/service/pubsub Host
//go:generate mockgen -package mockpubsub -destination mockpubsub/mocknet.go gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net Network

package pubsub

import (
	"context"

	"github.com/pkg/errors"

	floodsub "gx/ipfs/QmSjoxpBJV71bpSojnUY1K382Ly3Up55EspnDx6EKAmQX4/go-libp2p-floodsub"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	ihost "gx/ipfs/QmfCtHMCd9xFvehvHeVxtKVXJTMVTuHhyPRVHEXetn87vL/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// log is the logger for the service.
var log = logging.Logger("pubsub")

// Host represents an Alice host.
type Host = ihost.Host

// Service is the PubSub service.
type Service struct {
	config *Config
	host   Host
	pubsub *floodsub.PubSub
}

// Config contains configuration options for the PubSub service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "pubsub"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "PubSub"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Subscribes and pubishes to topics."
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

// Expose exposes the service to other services.
//
// It exposes the type:
//	github.com/libp2p/*go-libp2p-floodsub.PubSub
func (s *Service) Expose() interface{} {
	return s.pubsub
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	pubsub, err := floodsub.NewFloodSub(ctx, s.host)
	if err != nil {
		return errors.WithStack(err)
	}

	s.pubsub = pubsub

	running()
	<-ctx.Done()
	stopping()

	s.host.RemoveStreamHandler(floodsub.ID)
	s.pubsub = nil

	return errors.WithStack(ctx.Err())
}
