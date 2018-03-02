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

//go:generate mockgen -package mockbootstrap -destination mockbootstrap/mockbootstrap.go github.com/stratumn/alice/core/service/bootstrap Host
//go:generate mockgen -package mockbootstrap -destination mockbootstrap/mocknet.go gx/ipfs/QmQm7WmgYCa4RSz76tKEYpRjApjnRw8ZTUVQC15b8JM4a2/go-libp2p-net Network

// Package bootstrap defines a service that bootstraps a host from a set of
// well known peers.
package bootstrap

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/stratumn/alice/release"

	ihost "gx/ipfs/QmNmJZL7FQySMtE2BQuLMuZg2EB2CLEunJJUSVSc9YnnbV/go-libp2p-host"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	ma "gx/ipfs/QmWWQ2Txc2c6tqjsBpzg5Ar652cHPGNsQQp2SejkNmkUMb/go-multiaddr"
	pstore "gx/ipfs/QmXauCuJzmzapetmC6W4TuDJLL1yFFrVzSHoWv8YdbmnxH/go-libp2p-peerstore"
	inet "gx/ipfs/QmXfkENeeBvh3zYA51MaSdGUdBjhQ99cP5WQe8zgr6wchG/go-libp2p-net"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")

	// ErrNotEnoughPeers is returned when peers are connected.
	ErrNotEnoughPeers = errors.New("number of connected peers is under configuration threeshold")
)

// log is the logger for the service.
var log = logging.Logger("bootstrap")

// Host represents an Alice host.
type Host ihost.Host

// Service is the Bootstrap service.
type Service struct {
	config   *Config
	peers    []pstore.PeerInfo
	interval time.Duration
	timeout  time.Duration
	host     Host
}

// Config contains configuration options for the Bootstrap service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Needs are services that should be started in addition to the host
	// before bootstrapping.
	Needs []string `toml:"needs" comment:"Services that should be started in addition to the host before bootstrapping."`

	// Addresses is a list of known peer addresses.
	Addresses []string `toml:"addresses" comment:"A list of known peer addresses."`

	// MinPeerThreshold is the number of peers under which to bootstrap
	// connections.
	MinPeerThreshold int `toml:"min_peer_threshold" comment:"The number of peers under which to bootstrap connections."`

	// Interval is the duration of the interval between bootstrap jobs.
	Interval string `toml:"interval" comment:"Interval between bootstrap jobs."`

	// ConnectionTimeout is the connection timeout. It should be less than
	// the interval.
	ConnectionTimeout string `toml:"connection_timeout" comment:"The connection timeout. It should be less than the interval."`
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "bootstrap"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Bootstrap"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Periodically connects to known peers."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	return Config{
		Host:              "host",
		Needs:             []string{"network", "p2p"},
		Addresses:         release.BootstrapAddresses,
		MinPeerThreshold:  4,
		Interval:          "30s",
		ConnectionTimeout: "10s",
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

	return nil
}

// Expose exposes an empty struct, so that other services know it is ready.
func (s *Service) Expose() interface{} {
	return struct{}{}
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
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

	return errors.WithStack(ctx.Err())
}

// round boostraps connections if needed.
func (s *Service) round(ctx context.Context) error {
	event := log.EventBegin(ctx, "round", logging.Metadata{
		"threshold": s.config.MinPeerThreshold,
	})
	defer event.Done()
	defer func() {
		event.Append(logging.Metadata{
			"connected": len(s.host.Network().Peers()),
		})
	}()

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
