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
	methodTag = monitoring.NewTag("github.com/stratumn/go-indigonode/keys/grpc-method")
)

// Measures exposed by the grpcapi app.
var (
	requestReceived = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/grpcapi/request-received",
		"grpc request received",
		stats.UnitNone,
	))

	requestErr = monitoring.NewInt64(stats.Int64(
		"github.com/stratumn/go-indigonode/measure/grpcapi/request-error",
		"grpc request error",
		stats.UnitNone,
	))
)

// Views exposed by the grpcapi app.
var (
	RequestReceived = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/grpcapi/request-received",
		Description: "grpc request received",
		Measure:     requestReceived.Measure,
		TagKeys:     []tag.Key{methodTag.OCTag},
		Aggregation: view.Count(),
	}

	RequestError = &view.View{
		Name:        "github.com/stratumn/go-indigonode/views/grpcapi/request-error",
		Description: "grpc request error",
		Measure:     requestErr.Measure,
		TagKeys:     []tag.Key{methodTag.OCTag},
		Aggregation: view.Count(),
	}
)
