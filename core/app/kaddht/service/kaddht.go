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

/*
Package service defines a service that runs an instance of a Kademlia
distributed hash table server or client that can be used to route peer IDs to
network addresses.

It may also store other key values in addition to node IDs.

For more information on Kademlia DHT, checkout:

	https://en.wikipedia.org/wiki/Kademlia
*/
package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"

	levelds "github.com/ipfs/go-ds-leveldb"
	logging "github.com/ipfs/go-log"
	ihost "github.com/libp2p/go-libp2p-host"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	kaddhtopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// log is the logger for the service.
var log = logging.Logger("kaddht")

// Host represents a Stratumn Node host.
type Host interface {
	ihost.Host

	SetRouter(func(context.Context, peer.ID) (pstore.PeerInfo, error))
}

// Service is the Kademlia DHT service.
type Service struct {
	config *Config
	host   Host
	dht    *kaddht.IpfsDHT

	bsMu           sync.Mutex
	bsHostComplete bool
	bsNeeded       bool
	bsChan         chan struct{}
	bsInterval     time.Duration
	bsTimeout      time.Duration
}

// Config contains configuration options for the Kademlia DHT service.
type Config struct {
	// Host is the name of the host service.
	Host string `toml:"host" comment:"The name of the host service."`

	// Bootstrap is the name of the bootstrap service.
	Bootstrap string `toml:"bootstrap" comment:"The name of the bootstrap service."`

	// LevelDBPath is the path to the LevelDB database directory.
	LevelDBPath string `toml:"level_db_path" comment:"The path to the LevelDB database directory."`

	// BootstrapQueries is the number of queries during a bootstrap job.
	BootstrapQueries int `toml:"bootstrap_queries" comment:"The number of queries during a bootstrap job."`

	// BootstrapInterval is how long to wait between bootstrap jobs.
	BootstrapInterval string `toml:"bootstrap_interval" comment:"How long to wait between bootstrap jobs."`

	// BootstrapTimeout is the timeout for a bootstrap job.
	BootstrapTimeout string `toml:"bootstrap_timeout" comment:"The timeout for a bootstrap job."`

	// EnableClientMode is whether to run only as a client and not to store
	// DHT values.
	EnableClientMode bool `toml:"enable_client_mode" comment:"Whether to run only as a client and not store DHT values."`

	// TODO: LevelDB options.
}

// ID returns the unique identifier of the service.
func (s *Service) ID() string {
	return "kaddht"
}

// Name returns the human friendly name of the service.
func (s *Service) Name() string {
	return "Kademlia DHT"
}

// Desc returns a description of what the service does.
func (s *Service) Desc() string {
	return "Manages a Kademlia distributed hash table."
}

// Config returns the current service configuration or creates one with
// good default values.
func (s *Service) Config() interface{} {
	if s.config != nil {
		return *s.config
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(errors.WithStack(err))
	}

	dbPath, err := filepath.Abs(filepath.Join(cwd, "data", "kaddht"))
	if err != nil {
		panic(errors.WithStack(err))
	}

	return Config{
		Host:              "host",
		Bootstrap:         "bootstrap",
		LevelDBPath:       dbPath,
		BootstrapQueries:  3,
		BootstrapInterval: "1m",
		BootstrapTimeout:  "10s",
		EnableClientMode:  false,
	}
}

// SetConfig configures the service.
func (s *Service) SetConfig(config interface{}) error {
	conf := config.(Config)

	interval, err := time.ParseDuration(conf.BootstrapInterval)
	if err != nil {
		return errors.WithStack(err)
	}

	timeout, err := time.ParseDuration(conf.BootstrapTimeout)
	if err != nil {
		return errors.WithStack(err)
	}

	s.config = &conf
	s.bsInterval = interval
	s.bsTimeout = timeout

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

// Expose exposes the Kademlia DHT to other services.
//
// It exposes the type:
//	github.com/libp2p/*go-libp2p-kad-dht.IpfsDHT
func (s *Service) Expose() interface{} {
	return s.dht
}

// Likes returns the set of services this service can work with.
func (s *Service) Likes() map[string]struct{} {
	if s.config.Bootstrap == "" {
		return nil
	}

	likes := map[string]struct{}{}
	likes[s.config.Bootstrap] = struct{}{}

	return likes
}

// Befriend sets the liked services.
func (s *Service) Befriend(serviceID string, exposed interface{}) {
	s.bsMu.Lock()
	defer s.bsMu.Unlock()

	if exposed == nil {
		s.bsHostComplete = false
		return
	}

	s.bsHostComplete = true

	if s.bsNeeded {
		close(s.bsChan)
		s.bsNeeded = false
	}
}

// Run starts the service.
func (s *Service) Run(ctx context.Context, running, stopping func()) error {
	ds, err := levelds.NewDatastore(s.config.LevelDBPath, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	dhtCtx, closeDHT := context.WithCancel(ctx)
	defer closeDHT()

	if s.config.EnableClientMode {
		s.dht = kaddht.NewDHTClient(dhtCtx, s.host, ds)
		log.Event(ctx, "newDHTClient")
	} else {
		s.dht = kaddht.NewDHT(dhtCtx, s.host, ds)
		log.Event(ctx, "newDHT")
	}

	s.host.SetRouter(s.dht.FindPeer)

	s.bsChan = make(chan struct{}, 1)

	s.bsMu.Lock()
	if s.bsHostComplete {
		// Host already bootstrapped, so we can bootstrap the DHT
		// right away.
		close(s.bsChan)
	} else {
		s.bsNeeded = true
	}
	s.bsMu.Unlock()

	bsCtx, closeBS := context.WithCancel(ctx)
	defer closeBS()

	running()

	select {
	case <-s.bsChan:
		if err := s.bootstrap(bsCtx); err != nil {
			log.Event(ctx, "bootstrapError", logging.Metadata{
				"error": err.Error(),
			})
		}
	case <-ctx.Done():
	}

	<-ctx.Done()
	stopping()

	s.host.SetRouter(nil)
	s.host.RemoveStreamHandler(kaddhtopts.ProtocolDHT)
	s.host.RemoveStreamHandler(kaddhtopts.ProtocolDHTOld)

	closeBS()
	closeDHT()

	s.dht = nil

	if err := ds.Close(); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(ctx.Err())
}

// bootstrap launches periodic bootstrap jobs.
func (s *Service) bootstrap(ctx context.Context) error {
	bsConfig := kaddht.BootstrapConfig{
		Queries: s.config.BootstrapQueries,
		Period:  s.bsInterval,
		Timeout: s.bsTimeout,
	}

	err := s.dht.BootstrapWithConfig(ctx, bsConfig)
	if err != nil {
		log.Event(ctx, "bootstrapError", logging.Metadata{
			"error": err,
		})
		return errors.WithStack(err)
	}

	return nil
}
