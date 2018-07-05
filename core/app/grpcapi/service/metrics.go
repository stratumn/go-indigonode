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

package service

import (
	"github.com/stratumn/go-indigonode/core/monitoring"

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
