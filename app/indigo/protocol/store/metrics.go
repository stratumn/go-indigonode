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
