// Copyright © 2017  Stratumn SAS
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

package kaddht

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"

	levelds "gx/ipfs/QmPdvXuXWAR6gtxxqZw42RtSADMwz4ijVmYHGS542b6cMz/go-ds-leveldb"
	pstore "gx/ipfs/QmPgDWmTmuzvP7QE5zwo1TmjbJme9pmZHNujB2453jkCTr/go-libp2p-peerstore"
	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
	kaddht "gx/ipfs/QmWRBYr99v8sjrpbyNWMuGkQekn7b9ELoLSCe8Ny7Nxain/go-libp2p-kad-dht"
	peer "gx/ipfs/QmXYjuNuxVzXKJCfWasQk1RqkhVLDM9jtUKhqc2WPQmFSB/go-libp2p-peer"
	ihost "gx/ipfs/Qmc1XhrFEiSeBNn3mpfg6gEuYCt5im2gYmNVmncsvmpeAk/go-libp2p-host"
)

var (
	// ErrNotHost is returned when the connected service is not a host.
	ErrNotHost = errors.New("connected service is not a host")
)

// log is the logger for the service.
var log = logging.Logger("kaddht")

// host represents an Alice host.
type host interface {
	ihost.Host

	SetRouter(func(context.Context, peer.ID) (pstore.PeerInfo, error))
}

// Service is the Kademlia DHT service.
type Service struct {
	config *Config
	host   host
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

	if s.host, ok = exposed[s.config.Host].(host); !ok {
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
func (s *Service) Run(ctx context.Context, running, stopping chan struct{}) error {
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

	running <- struct{}{}

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
	stopping <- struct{}{}

	s.host.SetRouter(nil)

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

	proc, err := s.dht.BootstrapWithConfig(bsConfig)
	if err != nil {
		return errors.WithStack(err)
	}

	go func() {
		<-ctx.Done()
		if err := proc.Close(); err != nil && err != context.Canceled {
			log.Event(ctx, "procCloseError", logging.Metadata{
				"error": err,
			})
		}
	}()

	return nil
}
