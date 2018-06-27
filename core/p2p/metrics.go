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

func init() {
	var err error

	peerIDKey, err = tag.NewKey("github.com/stratumn/go-indigonode/keys/peerid")
	if err != nil {
		panic(err)
	}

	protocolIDKey, err = tag.NewKey("github.com/stratumn/go-indigonode/keys/protocolid")
	if err != nil {
		panic(err)
	}
}

// Tags used by the p2p layer.
var (
	peerIDKey     tag.Key
	protocolIDKey tag.Key
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
)

// Views exposed by the p2p layer.
var (
	BandwidthInView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-in",
		Description: "incoming messages bandwidth",
		Measure:     bandwidthIn,
		Aggregation: view.Count(),
	}

	BandwidthOutView = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/bandwidth-out",
		Description: "outgoing messages bandwidth",
		Measure:     bandwidthOut,
		Aggregation: view.Count(),
	}
)

// streamReporter collects stream metrics.
type streamReporter struct {
}

func (r *streamReporter) LogSentMessage(b int64) {
	stats.Record(context.Background(), bandwidthOut.M(b))
}

func (r *streamReporter) LogRecvMessage(b int64) {
	stats.Record(context.Background(), bandwidthIn.M(b))
}

func (r *streamReporter) LogSentMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
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

func (r *streamReporter) LogRecvMessageStream(b int64, pid protocol.ID, peerID peer.ID) {
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

func (r *streamReporter) GetBandwidthForPeer(peer.ID) metrics.Stats {
	return metrics.Stats{}
}

func (r *streamReporter) GetBandwidthForProtocol(protocol.ID) metrics.Stats {
	return metrics.Stats{}
}

func (r *streamReporter) GetBandwidthTotals() metrics.Stats {
	return metrics.Stats{}
}
