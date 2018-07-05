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

package protocol

import (
	"github.com/stratumn/go-indigonode/core/monitoring"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// Measures exposed by the bootstrap app.
var (
	participants = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/bootstrap/participants",
		"number of network participants",
		stats.UnitNone,
	))
)

// Views exposed by the bootstrap app.
var (
	Participants = &view.View{
		Name:        "indigo-node/views/bootstrap/participants",
		Description: "number of network participants",
		Measure:     participants.Measure,
		Aggregation: view.LastValue(),
	}
)
