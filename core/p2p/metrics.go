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

	"github.com/stratumn/go-indigonode/core/monitoring"

	metrics "gx/ipfs/QmVvu4bS5QLfS19ePkp5Wgzn2ZUma5oXTT9BgDFyQLxUZF/go-libp2p-metrics"
	protocol "gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	peer "gx/ipfs/QmcJukH2sAFjY3HdBKq35WDzWoL3UUu2gt9wdfqZTUyM74/go-libp2p-peer"
)

// Measures exposed by the p2p layer.
var (
	bandwidthIn = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/bandwidth-in",
		"incoming messages bandwidth",
		stats.UnitBytes,
	))

	bandwidthOut = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/bandwidth-out",
		"outgoing messages bandwidth",
		stats.UnitBytes,
	))

	connections = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/connections",
		"open connections",
		stats.UnitNone,
	))

	peers = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/peers",
		"connected peers",
		stats.UnitNone,
	))

	latency = monitoring.NewFloat64(stats.Float64(
		"github.com/stratumn/go-indigonode/measure/latency",
		"peer latency",
		stats.UnitMilliseconds,
	))
)

// Distributions used by p2p views.
var (
	DefaultLatencyDistribution = view.Distribution(0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

// Tags used by the p2p layer.
var (
	peerIDTag     = monitoring.NewTag("github.com/stratumn/go-indigonode/keys/peerid")
	protocolIDTag = monitoring.NewTag("github.com/stratumn/go-indigonode/keys/protocolid")
)

// Views exposed by the p2p layer.
var (
	BandwidthIn = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-in",
		Description: "incoming messages bandwidth",
		Measure:     bandwidthIn.Measure,
		TagKeys:     []tag.Key{peerIDTag.OCTag, protocolIDTag.OCTag},
		Aggregation: view.Count(),
	}

	BandwidthOut = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-out",
		Description: "outgoing messages bandwidth",
		Measure:     bandwidthOut.Measure,
		TagKeys:     []tag.Key{peerIDTag.OCTag, protocolIDTag.OCTag},
		Aggregation: view.Count(),
	}

	Connections = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/connections",
		Description: "open connections",
		Measure:     connections.Measure,
		Aggregation: view.LastValue(),
	}

	Peers = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/peers",
		Description: "connected peers",
		Measure:     peers.Measure,
		Aggregation: view.LastValue(),
	}

	Latency = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/latency",
		Description: "peer latency distribution",
		Measure:     latency.Measure,
		TagKeys:     []tag.Key{peerIDTag.OCTag},
		Aggregation: DefaultLatencyDistribution,
	}
)

// MetricsReporter collects stream and connection metrics.
type MetricsReporter struct{}

// LogSentMessage records the bandwidth used.
func (r *MetricsReporter) LogSentMessage(b int64) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(peerIDTag, "unknown").
		Tag(protocolIDTag, "unknown").
		Build()
	if err != nil {
		return
	}

	bandwidthOut.Record(ctx, b)
}

// LogRecvMessage records the bandwidth used.
func (r *MetricsReporter) LogRecvMessage(b int64) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(peerIDTag, "unknown").
		Tag(protocolIDTag, "unknown").
		Build()
	if err != nil {
		return
	}

	bandwidthIn.Record(ctx, b)
}

// LogSentMessageStream records the bandwidth used.
func (r *MetricsReporter) LogSentMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(peerIDTag, peerID.Pretty()).
		Tag(protocolIDTag, string(pid)).
		Build()
	if err != nil {
		return
	}

	bandwidthOut.Record(ctx, b)
}

// LogRecvMessageStream records the bandwidth used.
func (r *MetricsReporter) LogRecvMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(peerIDTag, peerID.Pretty()).
		Tag(protocolIDTag, string(pid)).
		Build()
	if err != nil {
		return
	}

	bandwidthIn.Record(ctx, b)
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
