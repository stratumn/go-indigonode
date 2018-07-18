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

// Package store contains the Indigo Store service.
package store

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/postgresstore"
	indigostore "github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	rpcpb "github.com/stratumn/go-indigonode/app/indigo/grpc/store"
	protocol "github.com/stratumn/go-indigonode/app/indigo/protocol/store"
	"github.com/stratumn/go-indigonode/app/indigo/protocol/store/sync"
	swarmSvc "github.com/stratumn/go-indigonode/core/app/swarm/service"
	"github.com/stratumn/go-indigonode/core/streamutil"

	"google.golang.org/grpc"

	ihost "gx/ipfs/QmfZTdmunzKzAGJrSvXXQbQ5kLLUiEMX5vdwux7iXkdk7D/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrNotSwarm is returned when the connected service is not a swarm.
	ErrNotSwarm = errors.New("connected service is not a swarm")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Host represents an Indigo Node host.
type Host = ihost.Host

// Service is the Indigo Store service.
type Service struct {
	config *Config
	host   Host
	swarm  *swarmSvc.Swarm
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
		Swarm:       "swarm",
		Version:     "0.1.0",
		NetworkID:   "indigo",
		StorageType: InMemoryStorage,
		PostgresConfig: &PostgresConfig{
			StorageDBURL: postgresstore.DefaultURL,
		},
		ValidationConfig: &ValidationConfig{},
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
	needs[s.config.Swarm] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.host, ok = exposed[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	if s.host.Peerstore().PrivKey(s.host.ID()) == nil {
		return errors.Wrap(ErrMissingPrivateKey, s.config.Host)
	}

	if s.swarm, ok = exposed[s.config.Swarm].(*swarmSvc.Swarm); !ok {
		return errors.Wrap(ErrNotSwarm, s.config.Swarm)
	}

	return nil
}

// Expose exposes the service to other services.
func (s *Service) Expose() interface{} {
	return nil
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	indigoStore, err := s.config.CreateIndigoStore(ctx)
	if err != nil {
		return err
	}

	auditStore, err := s.config.CreateAuditStore(ctx)
	if err != nil {
		return err
	}

	governanceManager, err := s.config.CreateValidator(ctx, indigoStore)
	if err != nil {
		return err
	}

	// We can't use the input context as a parent because it is cancelled
	// before we do the cleanup (see the <-ctx.Done() line).
	// For part of the PoP network cleanup, we need an active context.
	networkCtx, cancelNetwork := context.WithCancel(context.Background())
	defer cancelNetwork()

	networkMgr, err := s.config.JoinIndigoNetwork(networkCtx, s.host, s.swarm)
	if err != nil {
		return err
	}

	syncEngine := sync.NewSingleNodeEngine(s.host, indigoStore, streamutil.NewStreamProvider())
	s.store = protocol.New(ctx, networkMgr, syncEngine, indigoStore, auditStore, governanceManager)

	errChan := make(chan error)
	listenCtx, cancelListen := context.WithCancel(networkCtx)
	go func() { errChan <- networkMgr.Listen(listenCtx) }()

	running()
	<-ctx.Done()
	stopping()

	// Stop responding to sync requests.
	syncEngine.Close(networkCtx)

	// Close the store that uses the PoP network. If an error occurs, the context needs to be canceled before returning.
	err = s.store.Close(networkCtx)
	s.store = nil

	// Then leave the PoP network.
	cancelListen()
	<-errChan

	if err != nil {
		return err
	}

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
