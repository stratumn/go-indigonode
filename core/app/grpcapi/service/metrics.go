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

// Tags used by the grpcapi app.
var (
	methodTag = monitoring.NewTag("indigo-node/keys/grpc-method")
)

// Measures exposed by the grpcapi app.
var (
	requestReceived = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/grpcapi/request-received",
		"grpc request received",
		stats.UnitNone,
	))

	requestDuration = monitoring.NewFloat64(stats.Float64(
		"indigo-node/measure/grpcapi/request-duration",
		"grpc request duration",
		stats.UnitMilliseconds,
	))

	requestErr = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/grpcapi/request-error",
		"grpc request error",
		stats.UnitNone,
	))
)

// Views exposed by the grpcapi app.
var (
	RequestReceived = &view.View{
		Name:        "indigo-node/views/grpcapi/request-received",
		Description: "grpc request received",
		Measure:     requestReceived.Measure,
		TagKeys:     []tag.Key{methodTag.OCTag},
		Aggregation: view.Count(),
	}

	RequestDuration = &view.View{
		Name:        "indigo-node/views/grpcapi/request-duration",
		Description: "grpc request duration",
		Measure:     requestDuration.Measure,
		TagKeys:     []tag.Key{methodTag.OCTag},
		Aggregation: monitoring.DefaultLatencyDistribution,
	}

	RequestError = &view.View{
		Name:        "indigo-node/views/grpcapi/request-error",
		Description: "grpc request error",
		Measure:     requestErr.Measure,
		TagKeys:     []tag.Key{methodTag.OCTag},
		Aggregation: view.Count(),
	}
)
