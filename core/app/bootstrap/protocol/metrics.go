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

package protocol

import (
	"github.com/stratumn/go-node/core/monitoring"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// Measures exposed by the bootstrap app.
var (
	participants = monitoring.NewInt64(stats.Int64(
		"stratumn-node/measure/bootstrap/participants",
		"number of network participants",
		stats.UnitNone,
	))
)

// Views exposed by the bootstrap app.
var (
	Participants = &view.View{
		Name:        "stratumn-node/views/bootstrap/participants",
		Description: "number of network participants",
		Measure:     participants.Measure,
		Aggregation: view.LastValue(),
	}
)
