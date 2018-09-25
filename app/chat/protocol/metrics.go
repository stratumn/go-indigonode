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
	"go.opencensus.io/tag"
)

// Measures exposed by the chat app.
var (
	msgReceived = monitoring.NewInt64(stats.Int64(
		"stratumn-node/measure/chat/message-received",
		"chat message received",
		stats.UnitNone,
	))

	msgSent = monitoring.NewInt64(stats.Int64(
		"stratumn-node/measure/chat/message-sent",
		"chat message sent",
		stats.UnitNone,
	))

	msgError = monitoring.NewInt64(stats.Int64(
		"stratumn-node/measure/chat/message-error",
		"chat message error",
		stats.UnitNone,
	))
)

// Views exposed by the chat app.
var (
	MessageReceived = &view.View{
		Name:        "stratumn-node/views/chat/message-received",
		Description: "chat message received",
		Measure:     msgReceived.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: view.Count(),
	}

	MessageSent = &view.View{
		Name:        "stratumn-node/views/chat/message-sent",
		Description: "chat message sent",
		Measure:     msgSent.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: view.Count(),
	}

	MessageError = &view.View{
		Name:        "stratumn-node/views/chat/message-error",
		Description: "chat message error",
		Measure:     msgError.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: view.Count(),
	}
)
