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
	"go.opencensus.io/tag"
)

// Measures exposed by the chat app.
var (
	msgReceived = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/chat/message-received",
		"chat message received",
		stats.UnitNone,
	))

	msgSent = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/chat/message-sent",
		"chat message sent",
		stats.UnitNone,
	))

	msgError = monitoring.NewInt64(stats.Int64(
		"indigo-node/measure/chat/message-error",
		"chat message error",
		stats.UnitNone,
	))
)

// Views exposed by the chat app.
var (
	MessageReceived = &view.View{
		Name:        "indigo-node/views/chat/message-received",
		Description: "chat message received",
		Measure:     msgReceived.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: view.Count(),
	}

	MessageSent = &view.View{
		Name:        "indigo-node/views/chat/message-sent",
		Description: "chat message sent",
		Measure:     msgSent.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: view.Count(),
	}

	MessageError = &view.View{
		Name:        "indigo-node/views/chat/message-error",
		Description: "chat message error",
		Measure:     msgError.Measure,
		TagKeys:     []tag.Key{monitoring.PeerIDTag.OCTag},
		Aggregation: view.Count(),
	}
)
