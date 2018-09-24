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

// Package service defines a service the wraps a P2P host.
package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/host/grpc"
	swarmSvc "github.com/stratumn/go-indigonode/core/app/swarm/service"
	"github.com/stratumn/go-indigonode/core/p2p"

	"google.golang.org/grpc"

	mamask "gx/ipfs/QmSMZwvs3n4GBikZ7hKzT17c3bk65FmyZo2JqtJ16swqCv/multiaddr-filter"
	mafilter "gx/ipfs/QmSW4uNHbvQia8iZDXzbwjiyHQtnyo9aFqfQAMasj3TJ6Y/go-maddr-filter"
	ifconnmgr "gx/ipfs/QmWGGN1nysi1qgqto31bENwESkmZBY4YGK4sZC3qhnqhSv/go-libp2p-interface-connmgr"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
)

var (
	// ErrNotNetwork is returned when a specified service is not a network.
	ErrNotNetwork = errors.New("connected service is not a network or swarm")

	// ErrNotMonitoring is returned when a specified service is not a monitoring.
	ErrNotMonitoring = errors.New("connected service is not a monitoring")

	// ErrNotConnManager is returned when a specified service is not a
	// connection manager.
	ErrNotConnManager = errors.New("connected service is not a connection manager")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Service is the Host service.
type Service struct {
	config *Config

	netw inet.Network
	cmgr ifconnmgr.ConnManager

	negTimeout      time.Duration
	addrsFilters    *mafilter.Filters
	metricsInterval time.Duration

	host *p2p.Host
}

// Config contains configuration options for the Host service.
type Config struct {
	// Network is the name of the network or swarm service.
	Network string `toml:"network" comment:"The name of the network or swarm service."`

	// Monitoring is the name of the monitoring service.
	Monitoring string `toml:"monitoring" comment:"The name of the monitoring service."`

	// ConnectionManager is the name of the connection manager service.
	ConnectionManager string `toml:"connection_manager" comment:"The name of the connection manager service."`

	// NegotiationTimeout is the negotiation timeout.
	NegotiationTimeout string `toml:"negotiation_timeout" comment:"The negotiation timeout."`

	// AddressesNetmasks are CIDR netmasks to filter announced addresses.
	AddressesNetmasks []string `toml:"addresses_netmasks" comment:"CIDR netmasks to filter announced addresses."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "host"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Host"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Starts a P2P host."
}

// Config returns the current service configuration or creates one with
// good default values.
//
// It can panic but it can only happen during `indigo-node init`.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Network:            "swarm",
		Monitoring:         "monitoring",
		ConnectionManager:  "connmgr",
		NegotiationTimeout: "1m",
		AddressesNetmasks:  []string{},
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	negTimeout, err := time.ParseDuration(conf.NegotiationTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(conf.AddressesNetmasks) > 0 {
		addrsFilters := mafilter.NewFilters()

		for _, address := range conf.AddressesNetmasks {
			mask, err := mamask.NewMask(address)
			if err != nil {
				return errors.WithStack(err)
			}

			addrsFilters.AddDialFilter(mask)
		}

		s.addrsFilters = addrsFilters
	}

	s.negTimeout = negTimeout
	s.config = &conf

	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Network] = struct{}{}
	needs[s.config.Monitoring] = struct{}{}

	if s.config.ConnectionManager != "" {
		needs[s.config.ConnectionManager] = struct{}{}
	}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	// Try network first, then swarm.
	netw := exposed[s.config.Network]
	if s.netw, ok = netw.(inet.Network); !ok {
		swm, ok := netw.(*swarmSvc.Swarm)
		if !ok {
			return errors.Wrap(ErrNotNetwork, s.config.Network)
		}

		s.netw = swm.Swarm
	}

	if s.config.ConnectionManager != "" {
		cmgr := exposed[s.config.ConnectionManager]
		if s.cmgr, ok = cmgr.(ifconnmgr.ConnManager); !ok {
			return errors.Wrap(ErrNotConnManager, s.config.ConnectionManager)
		}
	}

	s.metricsInterval, ok = exposed[s.config.Monitoring].(time.Duration)
	if !ok {
		return errors.Wrap(ErrNotMonitoring, s.config.Monitoring)
	}

	return nil
}

// Expose exposes the service to other services.
//
// It exposes the type:
//
//	github.com/stratumn/go-indigonode/core/*p2p.Host
func (s *Service) Expose() interface{} {
	return s.host
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	opts := []p2p.HostOption{
		p2p.OptNegTimeout(s.negTimeout),
	}

	if s.cmgr != nil {
		opts = append(opts, p2p.OptConnManager(s.cmgr))
	}

	if s.addrsFilters != nil {
		opts = append(opts, p2p.OptAddrsFilters(s.addrsFilters))
	}

	s.host = p2p.NewHost(ctx, s.netw, opts...)

	running()
	<-ctx.Done()
	stopping()

	h := s.host
	s.host = nil

	if err := h.Close(); err != nil {
		return err
	}

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterHostServer(gs, grpcServer{
		GetHost: func() *p2p.Host { return s.host },
	})
}
