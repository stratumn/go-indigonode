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

package test

import (
	"context"

	"github.com/stratumn/go-node/core/monitoring"

	logging "gx/ipfs/QmSpJByNKFX1sCsHBEp3R73FL4NF6FnQTEGyNAXHm2GS52/go-log"
)

// TestLogger can be used in tests to generate events.
var TestLogger = logging.Logger("test")

// NewEvent creates a new empty event.
func NewEvent() *logging.EventInProgress {
	return TestLogger.EventBegin(context.Background(), "test")
}

// NewSpan creates a new empty span.
func NewSpan() *monitoring.Span {
	_, span := monitoring.StartSpan(context.Background(), "test", "test")
	return span
}
