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

package fossilizer

import (
	"github.com/stratumn/go-indigonode/core/monitoring"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Measures exposed by the indigo fossilizer app.
var (
	fossils = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/indigo/fossilizer/fossils",
		"number of fossilized segments",
		stats.UnitNone,
	))
)

// Views exposed by the indigo fossilizer app.
var (
	Fossils = &view.View{
		Name:        "indigo-node/views/indigo/fossilizer/fossils",
		Description: "number of fossilized segments",
		Measure:     fossils.Measure,
		TagKeys:     []tag.Key{monitoring.ErrorTag.OCTag},
		Aggregation: view.Count(),
	}
)
