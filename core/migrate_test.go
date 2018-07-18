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

package core

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
	kaddht "github.com/stratumn/go-indigonode/core/app/kaddht/service"
	swarm "github.com/stratumn/go-indigonode/core/app/swarm/service"
	"github.com/stratumn/go-indigonode/core/cfg"
	logger "github.com/stratumn/go-indigonode/core/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "")`)

	filename := filepath.Join(dir, "cfg.toml")

	// Save original configuration.
	err = ioutil.WriteFile(filename, []byte(confZero), 0600)
	require.NoError(t, err, "ioutil.WriteFile(filename, []byte(confZero), 0600)")

	set := NewConfigurableSet(BuiltinServices())

	// Migrate and load.
	err = LoadConfig(set, filename)
	require.NoError(t, err, "LoadConfig(set, filename)")

	migratedConf := set.Configs()

	// Create default config.
	defConf := NewConfigurableSet(BuiltinServices()).Configs()

	// Make sure both configs use the same private key and point to the
	// same files.
	swarmCfg := defConf["swarm"].(swarm.Config)
	swarmCfg.PeerID = cfg.ConfZeroPID
	swarmCfg.PrivateKey = cfg.ConfZeroPK
	defConf["swarm"] = swarmCfg
	dhtCfg := defConf["kaddht"].(kaddht.Config)
	dhtCfg.LevelDBPath = "data/kaddht"
	defConf["kaddht"] = dhtCfg
	logCfg := defConf["log"].(logger.Config)
	logCfg.Writers[0].Filename = "log.jsonld"
	defConf["log"] = logCfg

	// If migrations are consistent, both configs should be the same.
	for k, gotVal := range migratedConf {
		gotBuf := bytes.NewBuffer(nil)
		enc := toml.NewEncoder(gotBuf)
		enc.QuoteMapKeys(true)
		err := enc.Encode(gotVal)
		if !assert.NoErrorf(t, err, "%s: enc.Encode(gotVal)", k) {
			continue
		}

		wantVal := defConf[k]
		wantBuf := bytes.NewBuffer(nil)
		enc = toml.NewEncoder(wantBuf)
		enc.QuoteMapKeys(true)
		err = enc.Encode(wantVal)
		if !assert.NoErrorf(t, err, "%s: enc.Encode(wantVal)", k) {
			continue
		}

		assert.Equalf(t, wantBuf.String(), gotBuf.String(), "%s", k)
	}
}

// Original configuration before migrations.
const confZero = `
# Indigo Node configuration file. Keep private!!!

# Settings for the bootstrap module.
[bootstrap]

  # A list of known peer addresses.
  addresses = ["/dnsaddr/impulse.io/ipfs/Qmc1QbSba7RtPgxEw4NqXNeDpB5CpCTwv9dvdZRdTkche1","/dnsaddr/impulse.io/ipfs/QmQVdocY8ZbYxrKRSrff2Vxmm27Mhu6DgXyWXQwmuz1b6P","/dnsaddr/impulse.io/ipfs/QmQJib6mnEMgdCe3bGH1YP7JswHbQQejyNucvW9BjFqmWr","/dnsaddr/impulse.io/ipfs/Qmc1rLFp5stHrjtq4duFg6KakBcDCpB3bTjjMZVSAdnHLj"]

  # The connection timeout. It should be less than the interval.
  connection_timeout = "10s"

  # The name of the host service.
  host = "host"

  # Interval between bootstrap jobs.
  interval = "30s"

  # The number of peers under which to bootstrap connections.
  min_peer_threshold = 4

  # Services that should be started in addition to the host before bootstrapping.
  needs = ["p2p"]

  # The name of the swarm service.
  swarm = "swarm"

  # Configure the store used for network update proposals.
  [bootstrap.store_config]

    # Type of store to use.
    # Supported values: in-memory and file.
    type = "in-memory"

# Settings for the chat module.
[chat]
  
  # The name of the event service.
  event = "event"

  # The name of the host service.
  host = "host"

# Settings for the clock module.
[clock]

  # The name of the host service.
  host = "host"

  # How long to wait before closing the stream when writing the time to a peer.
  write_timeout = "10s"

# Settings for the coin module.
[coin]

  # The difficulty for block production.
  block_difficulty = 30

  # The name of the host service.
  host = "host"

  # The name of the kaddht service.
  kaddht = "kaddht"

  # The maximum number of transactions in a block.
  max_tx_per_block = 100

  # The reward miners should get when producing blocks.
  miner_reward = 10

  # The name of the pubsub service.
  pubsub = "pubsub"

  # The version of the coin service.
  version = 1

# Settings for the connmgr module.
[connmgr]

  # How long to keep a connection before it can be closed.
  grace_period = "20s"

  # Maximum number of connections to keep open (0 = disabled).
  high_water = 900

  # Minimum number of connections to keep open (0 = disabled).
  low_water = 600

# Settings for the core module.
[core]

  # Service to launch when starting the node.
  boot_service = "boot"

  # Whether to show the boot screen when starting the node.
  enable_boot_screen = true

  # Name of the host service used by the boot screen to display metrics and host addresses.
  boot_screen_host = "host"

  # The version of Indigo Node that generated this file.
  generated_by_version = "v0.0.1"

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts boot services."

    # Unique identifier of the service group.
    id = "boot"

    # Name of the service group.
    name = "Boot Services"

    # Services started by the group.
    services = ["system","bootstrap","api", "util"]

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts system services."

    # Unique identifier of the service group.
    id = "system"

    # Name of the service group.
    name = "System Services"

    # Services started by the group.
    services = ["signal","pruner","monitoring"]

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts P2P services."

    # Unique identifier of the service group.
    id = "p2p"

    # Name of the service group.
    name = "P2P Services"

    # Services started by the group.
    services = ["identify","relay","kaddht","ping","clock","pubsub"]

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts network services."

    # Unique identifier of the service group.
    id = "network"

    # Name of the service group.
    name = "Network Services"

    # Services started by the group.
    services = ["host","natmgr"]

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts API services."

    # Unique identifier of the service group.
    id = "api"

    # Name of the service group.
    name = "API Services"

    # Services started by the group.
    services = ["grpcapi","grpcweb"]

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts utility services."

    # Unique identifier of the service group.
    id = "util"

    # Name of the service group.
    name = "Utility Services"

    # Services started by the group.
    services = ["contacts","event"]

  [[core.service_groups]]

    # Description of the service group.
    description = "Starts Stratumn Indigo services for Proof-of-Process networks."

    # Unique identifier of the service group.
    id = "indigo"

    # Name of the service group.
    name = "Stratumn Indigo Services"

    # Services started by the group.
    services = ["indigostore","indigofossilizer"]

# Settings for the event module.
[event]

  # How long to wait before dropping a message when listeners are too slow.
  write_timeout = "100ms"    

# Settings for the grpcapi module.
[grpcapi]

  # Address to bind to.
  address = "/ip4/127.0.0.1/tcp/8904"

  # Whether to log requests.
  enable_request_logger = true

  # The name of the manager service.
  manager = "manager"

  # Path to a TLS certificate.
  tls_certificate_file = ""

  # Path to a TLS key.
  tls_key_file = ""

# Settings for the grpcweb module.
[grpcweb]

  # Address to bind to.
  address = "/ip4/127.0.0.1/tcp/8906"

  # The name of the grpcapi service.
  grpcapi = "grpcapi"

# Settings for the host module.
[host]

  # CIDR netmasks to filter announced addresses.
  addresses_netmasks = []

  # The name of the connection manager service.
  connection_manager = "connmgr"

  # The name of the monitoring service.
  monitoring = "monitoring"

  # The negotiation timeout.
  negotiation_timeout = "1m"

  # The name of the network or swarm service.
  network = "swarm"

# Settings for the identify module.
[identify]

  # The name of the host service.
  host = "host"

# Settings for the indigofossilizer module.
[indigofossilizer]

  # amount of the fee to use when sending transactions to the bitcoin blockchain (only applicable to the bitcoin fossilizer).
  bitcoin_fee = 15000

  # The type of fossilizer (eg: dummy, dummybatch, bitcoin...).
  fossilizer_type = "dummy"

  # The time interval between batches expressed in seconds (only applicable to fossilizers using batches).
  interval = 0

  # The maximum number of leaves of a merkle tree in a batch (only applicable to fossilizers using batches).
  max_leaves = 0

  # The version of the indigo fossilizer service.
  version = "0.1.0"

# Settings for the indigostore module.
[indigostore]

  # The name of the host service.
  host = "host"

  # The type of storage to use.
  # Supported values: in-memory and postgreSQL.
  storage_type = "in-memory"

  # The name of the swarm service.
  swarm = "swarm"

  # The ID of your Indigo PoP network.
  network_id = "indigo"

  # The version of the indigo service.
  version = "0.1.0"

  # Configure settings for the Indigo PostgreSQL Store in the following section.
  [indigostore.postgres]

    # If external storage is used, the url of that storage.
    storage_db_url = "postgres://postgres@postgres/postgres?sslmode=disable"

  # Configure settings for the validation rules of your indigo network in the following section.
  [indigostore.validation]

    # The directory where the validator scripts are located.
    plugins_path = ""

    # The path to the validation rules file.
    rules_path = ""

# Settings for the kaddht module.
[kaddht]

  # The name of the bootstrap service.
  bootstrap = "bootstrap"

  # How long to wait between bootstrap jobs.
  bootstrap_interval = "1m"

  # The number of queries during a bootstrap job.
  bootstrap_queries = 3

  # The timeout for a bootstrap job.
  bootstrap_timeout = "10s"

  # Whether to run only as a client and not store DHT values.
  enable_client_mode = false

  # The name of the host service.
  host = "host"

  # The path to the LevelDB database directory.
  level_db_path = "data/kaddht"

# Settings for the log module.
[log]

  [[log.writers]]

    # Whether to compress the file.
    compress = false

    # The file for a file logger.
    filename = "log.jsonld"

    # The formatter for the writer (json, text, color, journald).
    formatter = "json"

    # The log level for the writer (info, error, all).
    level = "all"

    # The maximum age of the file in days before a rotation.
    maximum_age = 7

    # The maximum number of backups.
    maximum_backups = 4

    # The maximum size of the file in megabytes before a rotation.
    maximum_size = 128

    # The type of writer (file, stdout, stderr).
    type = "file"

    # Whether to use local time instead of UTC for backups.
    use_local_time = false

# Settings for the monitoring module.
[monitoring]

  # Interval between updates of periodic stats.
  interval = "10s"

  # Name of the metrics exporter (prometheus or stackdriver). Leave empty to disable metrics.
  metrics_exporter = "prometheus"

  # Name of the trace exporter (jaeger or stackdriver). Leave empty to disable tracing.
  trace_exporter = ""
  
  # Fraction of traces to record.
  trace_sampling_ratio = 1.0

  # Jaeger configuration options (if enabled).
  [monitoring.jaeger]

    # Address of the Jaeger agent to collect traces.
    endpoint = "/ip4/127.0.0.1/tcp/14268"

  # Prometheus configuration options (if enabled).
  [monitoring.prometheus]

    # Address of the endpoint to expose Prometheus metrics.
    endpoint = "/ip4/127.0.0.1/tcp/8905"

  # Stackdriver configuration options (if enabled).
  [monitoring.stackdriver]

    # Identifier of the Stackdriver project.
    project_id = "your-stackdriver-project-id"

# Settings for the mssmux module.
[mssmux]

  # A map of protocols to stream muxers (protocol = service).
  [mssmux.routes]
    "/yamux/v1.0.0" = "yamux"

# Settings for the natmgr module.
[natmgr]

  # The name of the host service.
  host = "host"

# Settings for the ping module.
[ping]

  # The name of the host service.
  host = "host"

# Settings for the pruner module.
[pruner]

  # Interval between prune jobs.
  interval = "1m"

  # The name of the manager service.
  manager = "manager"

# Settings for the pubsub module.
[pubsub]

  # The name of the host service.
  host = "host"

# Settings for the raft module.
[raft]

  # the number of Node.Tick invocations that must pass between elections
  election_tick = 10

  # the number of Node.Tick invocations that must pass between heartbeats
  heartbeat_tick = 1

  # limits the max number of in-flight append messages during optimistic replication phase
  max_inflight_msgs = 256

  # limits the max size of each append message
  max_size_per_msg = 1048576

  # defines the unit of time in ms
  ticker_interval = 100

# Settings for the relay module.
[relay]

  # Whether to act as an intermediary node in relay circuits.
  enable_hop = false

  # The name of the host service.
  host = "host"

# Settings for the signal module.
[signal]

  # Allow forced shutdown by sending second signal.
  allow_forced_shutdown = true

  # The name of the manager service.
  manager = "manager"

# Settings for the storage module.
[storage]

  # The name of the host service.
  host = "host"

  # The time after which an upload session will be reset (and the partial file deleted)
  upload_timeout = "10m"

# Settings for the swarm module.
[swarm]

  # List of addresses to bind to.
  addresses = ["/ip4/0.0.0.0/tcp/8903","/ip6/::/tcp/8903"]

  # The peer ID of the host.
  peer_id = "` + cfg.ConfZeroPID + `"

  # The private key of the host.
    private_key = "` + cfg.ConfZeroPK + `"

  # The name of the stream muxer service.
  stream_muxer = "mssmux"

# Settings for the yamux module.
[yamux]

  # The size of the accept backlog.
  accept_backlog = 512

  # The connection write timeout.
  connection_write_timeout = "10s"

  # The keep alive interval.
  keep_alive_interval = "30s"

  # The maximum stream window size.
  max_stream_window_size = "512KB"
`
