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

// Package service defines a service that maintains a swarm of connections
// between this node and its peers.
package service

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/stratumn/go-indigonode/core/app/swarm/grpc"
	"github.com/stratumn/go-indigonode/core/p2p"
	"github.com/stratumn/go-indigonode/core/protector"
	"google.golang.org/grpc"

	"gx/ipfs/QmRqfgh56f8CrqpwH7D2s6t8zQRsvPoftT3sp5Y6SUhNA3/go-libp2p-swarm"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	"gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
	"gx/ipfs/QmdeiKhUy1TVGBaKxt7y1QmBDLBdisSrLJ1x58Eoj4PXUh/go-libp2p-peerstore"
	"gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

const (
	// ServiceID used by the swarm service.
	ServiceID = "swarm"
)

var (
	// ErrPeerIDMismatch is returned when the peer ID does not match the
	// private key.
	ErrPeerIDMismatch = errors.New("the peer ID does not match the private key")

	// ErrNotStreamMuxer is returned when a specified service is not a
	// stream muxer.
	ErrNotStreamMuxer = errors.New("connected service is not a stream muxer")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Transport represents a transport.
type Transport = smux.Transport

// Service is the Swarm service.
type Service struct {
	config *Config

	peerID  peer.ID
	privKey crypto.PrivKey
	addrs   []ma.Multiaddr
	swarm   *swarm.Swarm
	smuxer  Transport

	networkConfig protector.NetworkConfig
	networkMode   *protector.NetworkMode
}

// Swarm wraps a swarm with other data that could be useful to services.
// It's the type exposed by the swarm service.
type Swarm struct {
	NetworkConfig protector.NetworkConfig
	NetworkMode   *protector.NetworkMode
	Swarm         *swarm.Swarm
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return ServiceID
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Swarm"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Connects to peers."
}

// Config returns the current service configuration or creates one with
// good default values.
//
// It can panic but it can only happen during `indigo-node init`.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		panic(errors.WithStack(err))
	}

	b, err := privKey.Bytes()
	if err != nil {
		panic(errors.WithStack(err))
	}

	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		panic(errors.WithStack(err))
	}

	return Config{
		PeerID:     peerID.Pretty(),
		PrivateKey: crypto.ConfigEncodeKey(b),
		Addresses: []string{
			"/ip4/0.0.0.0/tcp/8903",
			"/ip6/::/tcp/8903",
		},
		StreamMuxer: "mssmux",
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	b, err := crypto.ConfigDecodeKey(conf.PrivateKey)
	if err != nil {
		return errors.WithStack(err)
	}

	privKey, err := crypto.UnmarshalPrivateKey(b)
	if err != nil {
		return errors.WithStack(err)
	}

	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return errors.WithStack(err)
	}

	if peerID.Pretty() != conf.PeerID {
		return errors.WithStack(ErrPeerIDMismatch)
	}

	addrs := make([]ma.Multiaddr, len(conf.Addresses))

	for i, address := range conf.Addresses {
		addrs[i], err = ma.NewMultiaddr(address)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	networkMode, err := conf.ParseNetworkMode()
	if err != nil {
		return err
	}

	s.peerID = peerID
	s.privKey = privKey
	s.addrs = addrs
	s.networkMode = networkMode

	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.StreamMuxer] = struct{}{}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.smuxer, ok = exposed[s.config.StreamMuxer].(Transport); !ok {
		return errors.WithStack(ErrNotStreamMuxer)
	}

	return nil
}

// Expose exposes the swarm to other services.
//
// It exposes the type:
//	github.com/stratumn/go-indigonode/core/service/*swarm.Swarm
func (s *Service) Expose() interface{} {
	return &Swarm{
		NetworkConfig: s.networkConfig,
		NetworkMode:   s.networkMode,
		Swarm:         s.swarm,
	}
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) (err error) {
	pstore := peerstore.NewPeerstore()

	if err = pstore.AddPrivKey(s.peerID, s.privKey); err != nil {
		return errors.WithStack(err)
	}

	if err = pstore.AddPubKey(s.peerID, s.privKey.GetPublic()); err != nil {
		return errors.WithStack(err)
	}

	protectCfg, err := NewProtectorConfig(s.config)
	if err != nil {
		return err
	}

	protect, networkConfig, err := protectCfg.Configure(ctx, s, pstore)
	if err != nil {
		return err
	}

	swmCtx, swmCancel := context.WithCancel(ctx)
	defer swmCancel()

	swm, err := swarm.NewSwarmWithProtector(swmCtx, s.addrs, s.peerID, pstore, protect, s.smuxer, &p2p.MetricsReporter{})
	if err != nil {
		return errors.WithStack(err)
	}

	s.networkConfig = networkConfig
	s.swarm = swm

	running()
	<-ctx.Done()
	stopping()

	swmCancel()

	s.swarm = nil
	s.networkConfig = nil

	if err = swm.Close(); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterSwarmServer(gs, grpcServer{
		GetSwarm: func() *swarm.Swarm {
			return s.swarm
		},
	})
}
