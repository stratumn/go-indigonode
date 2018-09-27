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

package service

import (
	"github.com/stratumn/go-node/core/monitoring"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Measures exposed by the swarm network layer.
var (
	connections = monitoring.NewInt64(stats.Int64(
		"stratumn-node/measure/connections",
		"open connections",
		stats.UnitNone,
	))

	peers = monitoring.NewInt64(stats.Int64(
		"stratumn-node/measure/peers",
		"connected peers",
		stats.UnitNone,
	))

	latency = monitoring.NewFloat64(stats.Float64(
		"stratumn-node/measure/latency",
		"peer latency",
		stats.UnitMilliseconds,
	))
)

// Views exposed by the swarm network layer.
var (
	Connections = &view.View{
		Name:        "stratumn-node/views/connections",
		Description: "open connections",
		Measure:     connections.Measure,
		Aggregation: view.LastValue(),
	}

	Peers = &view.View{
		Name:        "stratumn-node/views/peers",
		Description: "connected peers",
		Measure:     peers.Measure,
		Aggregation: view.LastValue(),
	}

	Latency = &view.View{
		Name:        "stratumn-node/views/latency",
		Description: "peer latency distribution",
		Measure:     latency.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: monitoring.DefaultLatencyDistribution,
	}
)
