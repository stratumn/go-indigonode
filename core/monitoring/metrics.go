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
