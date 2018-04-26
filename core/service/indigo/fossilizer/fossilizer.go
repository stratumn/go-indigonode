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

// Package fossilizer contains the Indigo Fossilizer service.
package fossilizer

import (
	"context"

	"google.golang.org/grpc"

	"github.com/pkg/errors"

	protocol "github.com/stratumn/alice/core/protocol/indigo/fossilizer"
	rpcpb "github.com/stratumn/alice/grpc/indigo/fossilizer"
	"github.com/stratumn/go-indigocore/blockchain/btc/btctimestamper"
)

var (
	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Service is the Indigo Fossilizer service.
type Service struct {
	config     *Config
	fossilizer *protocol.Fossilizer
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "indigofossilizer"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Indigo Fossilizer"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "A service to use Stratumn's Indigo Fossilizer technology."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Version:        "0.1.0",
		FossilizerType: Dummy,
		BtcFee:         btctimestamper.DefaultFee,
	}
}

// SetConfig configures the service handler.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)
	s.config = &conf
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	indigoFossilizer, err := s.config.CreateIndigoFossilizer(ctx)
	if err != nil {
		return err
	}

	s.fossilizer = protocol.New(indigoFossilizer)

	// we need to start the fossilizer in case it uses batches.
	errChan := make(chan error)
	go func() { errChan <- s.fossilizer.Start(ctx) }()
	<-s.fossilizer.Started(ctx)

	running()
	<-ctx.Done()
	stopping()

	if err := <-errChan; err != nil {
		return err
	}

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	rpcpb.RegisterIndigoFossilizerServer(gs, grpcServer{
		DoGetInfo: func(ctx context.Context) (interface{}, error) {
			if s.fossilizer == nil {
				return nil, ErrUnavailable
			}
			return s.fossilizer.GetInfo(ctx)
		},
		DoFossilize: func(ctx context.Context, data, meta []byte) error {
			if s.fossilizer == nil {
				return ErrUnavailable
			}
			return s.fossilizer.Fossilize(ctx, data, meta)
		},
	})
}
