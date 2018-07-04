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

	streamsIn = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/streams-in",
		"incoming streams",
		stats.UnitNone,
	))

	streamsOut = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/streams-out",
		"outgoing streams",
		stats.UnitNone,
	))

	streamsErr = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/streams-err",
		"errored streams",
		stats.UnitNone,
	))

	latency = monitoring.NewFloat64(stats.Float64(
		"github.com/stratumn/go-indigonode/measure/latency",
		"peer latency",
		stats.UnitMilliseconds,
	))
)

// Views exposed by the p2p layer.
var (
	BandwidthIn = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-in",
		Description: "incoming messages bandwidth",
		Measure:     bandwidthIn.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag, monitoring.ProtocolIDTag.OCTag},
		Aggregation: view.Count(),
	}

	BandwidthOut = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-out",
		Description: "outgoing messages bandwidth",
		Measure:     bandwidthOut.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag, monitoring.ProtocolIDTag.OCTag},
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

	StreamsIn = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/streams-in",
		Description: "incoming streams",
		Measure:     streamsIn.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag, monitoring.ProtocolIDTag.OCTag},
		Aggregation: view.Count(),
	}

	StreamsOut = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/streams-out",
		Description: "outgoing streams",
		Measure:     streamsOut.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag, monitoring.ProtocolIDTag.OCTag},
		Aggregation: view.Count(),
	}

	StreamsErr = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/streams-error",
		Description: "errored streams",
		Measure:     streamsErr.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag, monitoring.ErrorTag.OCTag},
		Aggregation: view.Count(),
	}

	Latency = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/latency",
		Description: "peer latency distribution",
		Measure:     latency.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: monitoring.DefaultLatencyDistribution,
	}
)

// MetricsReporter collects stream and connection metrics.
type MetricsReporter struct{}

// LogSentMessage records the bandwidth used.
func (r *MetricsReporter) LogSentMessage(b int64) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(monitoring.PeerIDTag, "unknown").
		Tag(monitoring.ProtocolIDTag, "unknown").
		Build()
	if err != nil {
		return
	}

	bandwidthOut.Record(ctx, b)
}

// LogRecvMessage records the bandwidth used.
func (r *MetricsReporter) LogRecvMessage(b int64) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(monitoring.PeerIDTag, "unknown").
		Tag(monitoring.ProtocolIDTag, "unknown").
		Build()
	if err != nil {
		return
	}

	bandwidthIn.Record(ctx, b)
}

// LogSentMessageStream records the bandwidth used.
func (r *MetricsReporter) LogSentMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(monitoring.PeerIDTag, peerID.Pretty()).
		Tag(monitoring.ProtocolIDTag, string(pid)).
		Build()
	if err != nil {
		return
	}

	bandwidthOut.Record(ctx, b)
}

// LogRecvMessageStream records the bandwidth used.
func (r *MetricsReporter) LogRecvMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
	ctx, err := monitoring.NewTaggedContext(context.Background()).
		Tag(monitoring.PeerIDTag, peerID.Pretty()).
		Tag(monitoring.ProtocolIDTag, string(pid)).
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
