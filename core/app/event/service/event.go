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

package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	eventgrpc "github.com/stratumn/go-indigonode/core/app/event/grpc"

	"google.golang.org/grpc"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// log is the logger for the service.
var log = logging.Logger("event")

// Service is the Event service.
type Service struct {
	config  *Config
	timeout time.Duration
	emitter Emitter
}

// Config contains configuration options for the Event service.
type Config struct {
	// WriteTimeout sets how long to wait before dropping a message when listeners are too slow.
	WriteTimeout string `toml:"write_timeout" comment:"How long to wait before dropping a message when listeners are too slow."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "event"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Event"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "An event emitter. Clients can connect to receive events from Indigo Node services."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		WriteTimeout: DefaultTimeout.String(),
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	timeout, err := time.ParseDuration(conf.WriteTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	s.timeout = timeout
	s.config = &conf

	return nil
}

// Expose exposes the event service to other services.
// It allows services to emit events in a standard way.
func (s *Service) Expose() interface{} {
	return s.emitter
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	s.emitter = NewEmitter(s.timeout)

	running()
	<-ctx.Done()
	stopping()

	s.emitter.Close()
	s.emitter = nil

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	eventgrpc.RegisterEmitterServer(gs, grpcServer{
		GetEmitter: func() Emitter { return s.emitter },
	})
}
