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
