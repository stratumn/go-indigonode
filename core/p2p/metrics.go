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

package p2p

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	metrics "gx/ipfs/QmVvu4bS5QLfS19ePkp5Wgzn2ZUma5oXTT9BgDFyQLxUZF/go-libp2p-metrics"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// Measures exposed by the p2p layer.
var (
	bandwidthIn = stats.Int64(
		"github.com/stratumn/go-indigonode/measure/bandwidth-in",
		"incoming messages bandwidth",
		stats.UnitBytes,
	)

	bandwidthOut = stats.Int64(
		"github.com/stratumn/go-indigonode/measure/bandwidth-out",
		"outgoing messages bandwidth",
		stats.UnitBytes,
	)

	connections = stats.Int64(
		"github.com/stratumn/go-indigonode/measure/connections",
		"open connections",
		stats.UnitNone,
	)

	peers = stats.Int64(
		"github.com/stratumn/go-indigonode/measure/peers",
		"connected peers",
		stats.UnitNone,
	)

	latency = stats.Int64(
		"github.com/stratumn/go-indigonode/measure/latency",
		"peer latency",
		stats.UnitMilliseconds,
	)
)

// Distributions used by p2p views.
var (
	DefaultLatencyDistribution = view.Distribution(0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

// Tags used by the p2p layer.
var (
	peerIDKey, _     = tag.NewKey("github.com/stratumn/go-indigonode/keys/peerid")
	protocolIDKey, _ = tag.NewKey("github.com/stratumn/go-indigonode/keys/protocolid")
)

// Views exposed by the p2p layer.
var (
	BandwidthInView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-in",
		Description: "incoming messages bandwidth",
		Measure:     bandwidthIn,
		TagKeys:     []tag.Key{peerIDKey, protocolIDKey},
		Aggregation: view.Count(),
	}

	BandwidthOutView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-out",
		Description: "outgoing messages bandwidth",
		Measure:     bandwidthOut,
		TagKeys:     []tag.Key{peerIDKey, protocolIDKey},
		Aggregation: view.Count(),
	}

	ConnectionsView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/connections",
		Description: "open connections",
		Measure:     connections,
		Aggregation: view.LastValue(),
	}

	PeersView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/peers",
		Description: "connected peers",
		Measure:     peers,
		Aggregation: view.LastValue(),
	}

	LatencyView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/latency",
		Description: "peer latency distribution",
		Measure:     latency,
		TagKeys:     []tag.Key{peerIDKey},
		Aggregation: DefaultLatencyDistribution,
	}
)

// MetricsReporter collects stream and connection metrics.
type MetricsReporter struct {
}

// LogSentMessage records the bandwidth used.
func (r *MetricsReporter) LogSentMessage(b int64) {
	ctx, err := tag.New(
		context.Background(),
		tag.Insert(peerIDKey, "none"),
		tag.Insert(protocolIDKey, "none"),
	)
	if err != nil {
		return
	}

	stats.Record(ctx, bandwidthOut.M(b))
}

// LogRecvMessage records the bandwidth used.
func (r *MetricsReporter) LogRecvMessage(b int64) {
	ctx, err := tag.New(
		context.Background(),
		tag.Insert(peerIDKey, "none"),
		tag.Insert(protocolIDKey, "none"),
	)
	if err != nil {
		return
	}

	stats.Record(ctx, bandwidthIn.M(b))
}

// LogSentMessageStream records the bandwidth used.
func (r *MetricsReporter) LogSentMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
	ctx, err := tag.New(
		context.Background(),
		tag.Insert(peerIDKey, peerID.Pretty()),
		tag.Insert(protocolIDKey, string(pid)),
	)
	if err != nil {
		return
	}

	stats.Record(ctx, bandwidthOut.M(b))
}

// LogRecvMessageStream records the bandwidth used.
func (r *MetricsReporter) LogRecvMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
	ctx, err := tag.New(
		context.Background(),
		tag.Insert(peerIDKey, peerID.Pretty()),
		tag.Insert(protocolIDKey, string(pid)),
	)
	if err != nil {
		return
	}

	stats.Record(ctx, bandwidthIn.M(b))
}

// GetBandwidthForPeer shouldn't be used.
func (r *MetricsReporter) GetBandwidthForPeer(peer.ID) metrics.Stats {
	return metrics.Stats{}
}

// GetBandwidthForProtocol shouldn't be used.
func (r *MetricsReporter) GetBandwidthForProtocol(protocol.ID) metrics.Stats {
	return metrics.Stats{}
}

// GetBandwidthTotals shouldn't be used.
func (r *MetricsReporter) GetBandwidthTotals() metrics.Stats {
	return metrics.Stats{}
}
