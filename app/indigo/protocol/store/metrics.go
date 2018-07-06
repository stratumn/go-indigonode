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

package store

import (
	"github.com/stratumn/go-indigonode/core/monitoring"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Measures exposed by the indigo store app.
var (
	segmentsCreated = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/indigo/store/segments-created",
		"number of segments created",
		stats.UnitNone,
	))

	segmentsReceived = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/indigo/store/segments-received",
		"number of segments received",
		stats.UnitNone,
	))

	invalidSegments = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/indigo/store/segments-invalid",
		"number of invalid segments received",
		stats.UnitNone,
	))
)

// Views exposed by the indigo store app.
var (
	SegmentsCreated = &view.View{
		Name:        "indigo-node/views/indigo/store/segments-created",
		Description: "number of segments created",
		Measure:     segmentsCreated.Measure,
		TagKeys:     []tag.Key{monitoring.ErrorTag.OCTag},
		Aggregation: view.Count(),
	}

	SegmentsReceived = &view.View{
		Name:        "indigo-node/views/indigo/store/segments-received",
		Description: "number of segments received",
		Measure:     segmentsReceived.Measure,
		TagKeys:     []tag.Key{monitoring.ErrorTag.OCTag},
		Aggregation: view.Count(),
	}

	InvalidSegments = &view.View{
		Name:        "indigo-node/views/indigo/store/segments-invalid",
		Description: "number of invalid segments received",
		Measure:     invalidSegments.Measure,
		Aggregation: view.Count(),
	}
)
