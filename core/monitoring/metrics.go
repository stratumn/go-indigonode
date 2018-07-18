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

package monitoring

import (
	"context"

	"go.opencensus.io/stats"
)

// Int64Metric is an integer metric that can be tagged and recorded.
type Int64Metric struct {
	Measure *stats.Int64Measure
}

// NewInt64 creates a new Int64 metrics.
func NewInt64(m *stats.Int64Measure) *Int64Metric {
	return &Int64Metric{Measure: m}
}

// Record a value for the receiver metric.
func (m Int64Metric) Record(ctx context.Context, val int64) {
	stats.Record(ctx, m.Measure.M(val))
}

// Float64Metric is a float metric that can be tagged and recorded.
type Float64Metric struct {
	Measure *stats.Float64Measure
}

// NewFloat64 creates a new Float64 metrics.
func NewFloat64(m *stats.Float64Measure) *Float64Metric {
	return &Float64Metric{Measure: m}
}

// Record a value for the receiver metric.
func (m Float64Metric) Record(ctx context.Context, val float64) {
	stats.Record(ctx, m.Measure.M(val))
}
