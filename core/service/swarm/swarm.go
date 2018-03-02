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

//go:generate mockgen -package mockswarm -destination mockswarm/mockswarm.go github.com/stratumn/alice/core/service/swarm Transport

// Package swarm defines a service that maintains a swarm of connections
// between this node and its peers.
package swarm

import (
	"context"

	gometrics "github.com/armon/go-metrics"
	"github.com/pkg/errors"
	"github.com/stratumn/alice/core/service/metrics"
	pb "github.com/stratumn/alice/grpc/swarm"
	"google.golang.org/grpc"

	swarm "gx/ipfs/QmSwZMWwFZSUpe5muU2xgTUwppH24KfMwdPXiwbEp2c6G5/go-libp2p-swarm"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	pstore "gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	peer "gx/ipfs/QmZoWKhxUmZ2seW4BzX6fJkNR8hh9PsGModr7q171yq2SS/go-libp2p-peer"
	crypto "gx/ipfs/QmaPbCnUMBohSGo3KnxEa2bHqyJVVeEEcwtqJAYxerieBo/go-libp2p-crypto"
)

var (
	// ErrPeerIDMismatch is returned whe the peer ID does not match the
	// private key.
	ErrPeerIDMismatch = errors.New("the peer ID does not match the private key")

	// ErrNotStreamMuxer is returned when a specified service is not a
	// stream muxer.
	ErrNotStreamMuxer = errors.New("connected service is not a stream muxer")

	// ErrNotMetrics is returned when a specified service is not of type
	// metrics.
	ErrNotMetrics = errors.New("connected service is not of type metrics")

	// ErrUnavailable is returned from gRPC methods when the service is not
	// available.
	ErrUnavailable = errors.New("the service is not available")
)

// Transport represents a transport.
type Transport = smux.Transport

// Service is the Swarm service.
type Service struct {
	config *Config

	metrics *metrics.Metrics
	smuxer  Transport

	peerID  peer.ID
	privKey crypto.PrivKey
	addrs   []ma.Multiaddr
	swarm   *swarm.Swarm
}

// Config contains configuration options for the Swarm service.
type Config struct {
	// PeerID is peer ID of the node.
	PeerID string `toml:"peer_id" comment:"The peer ID of the host."`

	// PrivateKey is the private key of the node.
	PrivateKey string `toml:"private_key" comment:"The private key of the host."`

	// Addresses are the list of addresses to bind to.
	Addresses []string `toml:"addresses" comment:"List of addresses to bind to."`

	// StreamMuxer is the name of the stream muxer service.
	StreamMuxer string `toml:"stream_muxer" comment:"The name of the stream muxer service."`

	// Metrics is the name of the metrics service.
	Metrics string `toml:"metrics" comment:"The name of the metrics service (blank = disabled)."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "swarm"
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
// It can panic but it can only happen during `alice init`.
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
		Metrics:     "metrics",
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

	s.peerID = peerID
	s.privKey = privKey
	s.addrs = addrs

	s.config = &conf
	return nil
}

// Needs returns the set of services this service depends on.
func (s *Service) Needs() map[string]struct{} {
	needs := map[string]struct{}{}
	needs[s.config.StreamMuxer] = struct{}{}

	if s.config.Metrics != "" {
		needs[s.config.Metrics] = struct{}{}
	}

	return needs
}

// Plug sets the connected services.
func (s *Service) Plug(exposed map[string]interface{}) error {
	var ok bool

	if s.smuxer, ok = exposed[s.config.StreamMuxer].(Transport); !ok {
		return errors.WithStack(ErrNotStreamMuxer)
	}

	if s.config.Metrics != "" {
		mtrx := exposed[s.config.Metrics]
		if s.metrics, ok = mtrx.(*metrics.Metrics); !ok {
			return errors.Wrap(ErrNotMetrics, s.config.Metrics)
		}
	}

	return nil
}

// Expose exposes the swarm to other services.
//
// It exposes the type:
//	github.com/libp2p/*go-libp2p-swarm.Swarm
func (s *Service) Expose() interface{} {
	return s.swarm
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	pstore := pstore.NewPeerstore()

	if err := pstore.AddPrivKey(s.peerID, s.privKey); err != nil {
		return errors.WithStack(err)
	}

	if err := pstore.AddPubKey(s.peerID, s.privKey.GetPublic()); err != nil {
		return errors.WithStack(err)
	}

	swmCtx, swmCancel := context.WithCancel(ctx)
	defer swmCancel()

	// TODO: protector?
	swm, err := swarm.NewSwarmWithProtector(swmCtx, s.addrs, s.peerID, pstore, nil, s.smuxer, s.metrics)
	if err != nil {
		return errors.WithStack(err)
	}

	s.swarm = swm

	var cancelPeriodicMetrics func()

	if s.metrics != nil {
		cancelPeriodicMetrics = s.metrics.AddPeriodicHandler(s.periodicMetrics)
	}

	running()
	<-ctx.Done()
	stopping()

	if cancelPeriodicMetrics != nil {
		cancelPeriodicMetrics()
	}

	swmCancel()

	s.swarm = nil

	if err := swm.Close(); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(ctx.Err())
}

// AddToGRPCServer adds the service to a gRPC server.
func (s *Service) AddToGRPCServer(gs *grpc.Server) {
	pb.RegisterSwarmServer(gs, grpcServer{func() *swarm.Swarm {
		return s.swarm
	}})
}

// periodicMetrics sends periodic stats about peers, connections, and total
// bandwidth usage.
func (s *Service) periodicMetrics(sink gometrics.MetricSink) {
	labels := []gometrics.Label{{
		Name:  "service",
		Value: s.ID(),
	}}

	sink.SetGaugeWithLabels([]string{"peers"}, float32(len(s.swarm.Peers())), labels)
	sink.SetGaugeWithLabels([]string{"connections"}, float32(len(s.swarm.Connections())), labels)

	stats := s.metrics.GetBandwidthTotals()

	sink.SetGaugeWithLabels([]string{"bandwidthTotalIn"}, float32(stats.TotalIn), labels)
	sink.SetGaugeWithLabels([]string{"bandwidthTotalOut"}, float32(stats.TotalOut), labels)
	sink.SetGaugeWithLabels([]string{"bandwidthRateIn"}, float32(stats.RateIn), labels)
	sink.SetGaugeWithLabels([]string{"bandwidthRateOut"}, float32(stats.RateOut), labels)
}
