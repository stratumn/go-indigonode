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

// Package service defines the bootstrap service implementation.
package service

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-node/core/app/bootstrap/grpc"
	protocol "github.com/stratumn/go-node/core/app/bootstrap/protocol"
	"github.com/stratumn/go-node/core/app/bootstrap/protocol/proposal"
	swarm "github.com/stratumn/go-node/core/app/swarm/service"
	"github.com/stratumn/go-node/core/protector"
	protectorpb "github.com/stratumn/go-node/core/protector/pb"
	"github.com/stratumn/go-node/core/streamutil"
	"github.com/stratumn/go-node/release"

	"google.golang.org/grpc"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	inet "gx/ipfs/QmZNJyx9GGCX4GeuHnLB8fxaxMLs4MjTjHokxfQcCd6Nve/go-libp2p-net"
	pstore "gx/ipfs/Qmda4cPRvSRyox3SqgJN6DfSZGU5TtHufPTp9uXjFj71X6/go-libp2p-peerstore"
	ihost "gx/ipfs/QmeMYW7Nj8jnnEfs9qhm7SxKkoDPUWXu3MsxX6BFwz34tf/go-libp2p-host"
)

const (
	// ServiceID used by the bootstrap service.
	ServiceID = "bootstrap"
)

var (
	// ErrNotEnoughPeers is returned when not enough peers are connected.
	ErrNotEnoughPeers = errors.New("number of connected peers is under configuration threshold")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")

	// ErrNotAllowed is return from gRPC methods when the command is not allowed
	// in the current configuration.
	ErrNotAllowed = errors.New("operation not allowed")
)

// log is the logger for the service.
var log = logging.Logger(ServiceID)

// Host represents an Indigo Node host.
type Host ihost.Host

// Service is the Bootstrap service.
type Service struct {
	config   *Config
	peers    []pstore.PeerInfo
	interval time.Duration
	timeout  time.Duration
	host     Host
	swarm    *swarm.Swarm
	protocol protocol.Handler
	store    proposal.Store
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return ServiceID
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Bootstrap"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Bootstraps network connections."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Host:              "host",
		Swarm:             "swarm",
		Needs:             []string{"p2p"},
		Addresses:         release.BootstrapAddresses,
		MinPeerThreshold:  4,
		Interval:          "30s",
		ConnectionTimeout: "10s",
		StoreConfig: &StoreConfig{
			Type: InMemoryStore,
		},
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	var peers []pstore.PeerInfo

	for _, address := range conf.Addresses {
		addr, err := ma.NewMultiaddr(address)
		if err != nil {
			return errors.WithStack(err)
		}

		pi, err := pstore.InfoFromP2pAddr(addr)
		if err != nil {
			return errors.WithStack(err)
		}

		peers = append(peers, *pi)
	}

	interval, err := time.ParseDuration(conf.Interval)
	if err != nil {
		return errors.WithStack(err)
	}

	timeout, err := time.ParseDuration(conf.ConnectionTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	s.config = &conf
	s.peers = peers
	s.interval = interval
	s.timeout = timeout

	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.Host] = struct{}{}
	needs[s.config.Swarm] = struct{}{}

	for _, service := range s.config.Needs {
		needs[service] = struct{}{}
	}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.host, ok = exposed[s.config.Host].(Host); !ok {
		return errors.Wrap(ErrNotHost, s.config.Host)
	}

	if s.swarm, ok = exposed[s.config.Swarm].(*swarm.Swarm); !ok {
		return errors.Wrap(ErrNotSwarm, s.config.Swarm)
	}

	return nil
}

// Expose exposes an empty struct, so that other services know it is ready.
func (s *Service) Expose() interface{} {
	return struct{}{}
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	if s.config.MinPeerThreshold > 0 {
		// Bootstrap until we have at least one connected peer.
		for {
			if err := s.round(ctx); err != nil {
				log.Event(ctx, "roundError", logging.Metadata{
					"error": err.Error(),
				})
			}

			connected := s.host.Network().Peers()
			if len(connected) > 0 {
				break
			}

			time.Sleep(time.Second)
		}
	}

	protocolCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := s.config.NewStore(protocolCtx)
	if err != nil {
		return err
	}

	protocolHandler, err := protocol.New(
		s.host,
		streamutil.NewStreamProvider(),
		s.swarm.NetworkMode,
		s.swarm.NetworkConfig,
		store,
	)
	if err != nil {
		return err
	}

	err = protocolHandler.Handshake(protocolCtx)
	if err != nil {
		return err
	}

	s.store = store
	s.protocol = protocolHandler

	ticker := time.NewTicker(s.interval)

	running()

RUN_LOOP:
	for {
		select {
		case <-ticker.C:
			if err := s.round(ctx); err != nil {
				log.Event(ctx, "roundError", logging.Metadata{
					"error": err.Error(),
				})
			}
		case <-ctx.Done():
			break RUN_LOOP
		}
	}

	stopping()

	ticker.Stop()

	s.protocol.Close(protocolCtx)
	s.protocol = nil
	s.store = nil

	return errors.WithStack(ctx.Err())
}

// round boostraps connections if needed.
func (s *Service) round(ctx context.Context) (err error) {
	event := log.EventBegin(ctx, "round")
	defer func() {
		if err != nil {
			event.SetError(err)
		}

		event.Append(logging.Metadata{
			"connected": len(s.host.Network().Peers()),
		})

		event.Done()
	}()

	if s.swarm.NetworkMode == nil || s.swarm.NetworkMode.ProtectionMode == "" {
		return s.publicRound(ctx)
	}

	return s.privateRound(ctx)
}

// privateRound boostraps connections in a private network.
func (s *Service) privateRound(ctx context.Context) error {
	event := log.EventBegin(ctx, "private_round")
	defer event.Done()

	networkState := s.swarm.NetworkConfig.NetworkState(ctx)
	event.Append(logging.Metadata{
		"network_state": protectorpb.NetworkState_name[int32(networkState)],
	})

	peers := s.swarm.NetworkConfig.AllowedPeers(ctx)
	wg := sync.WaitGroup{}
	eventLock := sync.Mutex{}

	for _, peerID := range peers {
		if peerID == s.host.ID() {
			continue
		}

		if s.host.Network().Connectedness(peerID) != inet.Connected {
			wg.Add(1)

			peerStore := s.host.Peerstore()
			// If addresses have been evicted from the peer store, we need to
			// re-add them.
			// This can happen when a node has been down too long, libp2p will
			// erase its entry from the peer store.
			if len(peerStore.Addrs(peerID)) == 0 {
				addrs := s.swarm.NetworkConfig.AllowedAddrs(ctx, peerID)
				peerStore.AddAddrs(peerID, addrs, pstore.PermanentAddrTTL)
			}

			pi := peerStore.PeerInfo(peerID)

			go func(pi pstore.PeerInfo) {
				defer wg.Done()

				err := s.host.Connect(ctx, pi)

				eventLock.Lock()
				defer eventLock.Unlock()

				if err != nil {
					event.Append(logging.Metadata{
						pi.ID.Pretty(): err.Error(),
					})
				} else {
					event.Append(logging.Metadata{
						pi.ID.Pretty(): "connected",
					})
				}
			}(pi)
		}
	}

	wg.Wait()

	return nil
}

// publicRound boostraps connections in a public network.
func (s *Service) publicRound(ctx context.Context) error {
	event := log.EventBegin(ctx, "public_round", logging.Metadata{
		"threshold": s.config.MinPeerThreshold,
	})
	defer event.Done()

	threshold := s.config.MinPeerThreshold
	connected := s.host.Network().Peers()
	current := len(connected)

	if current >= threshold {
		return nil
	}

	var candidates []pstore.PeerInfo

	for _, pi := range s.peers {
		if s.host.Network().Connectedness(pi.ID) != inet.Connected {
			candidates = append(candidates, pi)
		}
	}

	needed := threshold - current
	selected := randPeers(candidates, needed)
	numSelected := len(selected) // could be less than value of needed

	connCtx, cancelConn := context.WithTimeout(ctx, s.timeout)
	defer cancelConn()

	var wg sync.WaitGroup
	wg.Add(numSelected)

	for _, pi := range selected {
		go func(pi pstore.PeerInfo) {
			defer wg.Done()

			if err := s.connect(connCtx, pi); err != nil {
				return
			}
		}(pi)
	}

	wg.Wait()

	if err := connCtx.Err(); err != nil {
		err = errors.WithStack(err)
		event.SetError(err)
	}

	if err := errors.WithStack(ctx.Err()); err != nil {
		event.SetError(err)
		return err
	}

	return nil
}

// connect connects to a peer.
func (s *Service) connect(ctx context.Context, pi pstore.PeerInfo) error {
	s.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.PermanentAddrTTL)

	if err := s.host.Connect(ctx, pi); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(ctx.Err())
}

// randPeers selects random peers from a pool of peers.
func randPeers(pool []pstore.PeerInfo, n int) []pstore.PeerInfo {
	l := len(pool)
	if n > l {
		n = l
	}

	peers := make([]pstore.PeerInfo, n)
	perm := rand.Perm(l)
	rand.Seed(time.Now().UTC().UnixNano())

	for i := range peers {
		peers[i] = pool[perm[i]]
	}

	return peers
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterBootstrapServer(gs, grpcServer{
		GetNetworkMode: func() *protector.NetworkMode {
			if s.swarm == nil {
				return nil
			}

			return s.swarm.NetworkMode
		},
		GetProtocolHandler: func() protocol.Handler {
			return s.protocol
		},
		GetProposalStore: func() proposal.Store {
			return s.store
		},
	})
}
