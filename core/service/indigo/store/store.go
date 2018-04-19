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

//go:generate mockgen -package mockstore -destination mockstore/mockstore.go github.com/stratumn/alice/core/service/indigo/store Host

// Package store contains the Indigo Store service.
package store

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/cfg"
	protocol "github.com/stratumn/alice/core/protocol/indigo/store"
	rpcpb "github.com/stratumn/alice/grpc/indigo/store"
	"github.com/stratumn/go-indigocore/cs"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"

	"google.golang.org/grpc"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Host represents an Alice host.
type Host = ihost.Host

// Service is the Indigo service.
type Service struct {
	config *Config
	host   Host
	store  *protocol.Store
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "indigostore"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Indigo Store"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "A service to use Stratumn's Indigo Store technology."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	// Set the default configuration settings of your service here.
	return Config{
		Host:        "host",
		Version:     "0.1.0",
		NetworkID:   "",
		PrivateKey:  cfg.ConfZeroPK,
		StorageType: "in-memory",
	}
}

// SetConfig configures the service handler.
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
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	hostPrivateKey, err := s.config.UnmarshalPrivateKey()
	if err != nil {
		return err
	}

	indigoStore, err := s.config.CreateIndigoStore()
	if err != nil {
		return err
	}

	auditStore, err := s.config.CreateAuditStore()
	if err != nil {
		return err
	}

	// We can't use the input context as a parent because it is cancelled
	// before we do the cleanup (see the <-ctx.Done() line).
	// For part of the floodsub cleanup, we need an active context.
	networkCtx, cancelNetwork := context.WithCancel(context.Background())
	defer cancelNetwork()

	networkMgr := protocol.NewNetworkManager(hostPrivateKey)
	if err := networkMgr.Join(networkCtx, s.config.NetworkID, s.host); err != nil {
		return err
	}

	s.store = protocol.New(networkMgr, indigoStore, auditStore)

	errChan := make(chan error)
	listenCtx, cancelListen := context.WithCancel(networkCtx)
	go func() { errChan <- networkMgr.Listen(listenCtx) }()

	running()
	<-ctx.Done()
	stopping()

	// First close the store that uses the PoP network.
	s.store.Close(networkCtx)
	s.store = nil

	// Then leave the PoP network.
	cancelListen()
	<-errChan

	err = networkMgr.Leave(networkCtx, s.config.NetworkID)
	if err != nil {
		return err
	}

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	rpcpb.RegisterIndigoStoreServer(gs, grpcServer{
		DoGetInfo: func() (interface{}, error) {
			if s.store == nil {
				return nil, ErrUnavailable
			}

			return s.store.GetInfo(context.Background())
		},
		DoCreateLink: func(ctx context.Context, link *cs.Link) (*types.Bytes32, error) {
			if s.store == nil {
				return nil, ErrUnavailable
			}

			return s.store.CreateLink(ctx, link)
		},
		DoGetSegment: func(ctx context.Context, linkHash *types.Bytes32) (*cs.Segment, error) {
			if s.store == nil {
				return nil, ErrUnavailable
			}

			return s.store.GetSegment(ctx, linkHash)
		},
		DoFindSegments: func(ctx context.Context, filter *indigostore.SegmentFilter) (cs.SegmentSlice, error) {
			if s.store == nil {
				return nil, ErrUnavailable
			}

			return s.store.FindSegments(ctx, filter)
		},
		DoGetMapIDs: func(ctx context.Context, filter *indigostore.MapFilter) ([]string, error) {
			if s.store == nil {
				return nil, ErrUnavailable
			}

			return s.store.GetMapIDs(ctx, filter)
		},
		DoAddEvidence: func(ctx context.Context, linkHash *types.Bytes32, evidence *cs.Evidence) error {
			if s.store == nil {
				return ErrUnavailable
			}

			return s.store.AddEvidence(ctx, linkHash, evidence)
		},
		DoGetEvidences: func(ctx context.Context, linkHash *types.Bytes32) (*cs.Evidences, error) {
			if s.store == nil {
				return nil, ErrUnavailable
			}

			return s.store.GetEvidences(ctx, linkHash)
		},
	})
}
